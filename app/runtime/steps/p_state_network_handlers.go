package steps

import (
	"context"

	"github.com/imfact-labs/currency-model/app/runtime/contracts"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	isaacdatabase "github.com/imfact-labs/mitum2/isaac/database"
	isaacnetwork "github.com/imfact-labs/mitum2/isaac/network"
	isaacstates "github.com/imfact-labs/mitum2/isaac/states"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/network/quicmemberlist"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/logging"
	"github.com/pkg/errors"
)

func PStatesNetworkHandlers(pctx context.Context) (context.Context, error) {
	if err := launch.AttachHandlerOperation(pctx); err != nil {
		return pctx, err
	}

	if err := AttachHandlerSendOperation(pctx); err != nil {
		return pctx, err
	}

	if err := launch.AttachHandlerStreamOperations(pctx); err != nil {
		return pctx, err
	}

	if err := launch.AttachHandlerProposals(pctx); err != nil {
		return pctx, err
	}

	return pctx, nil
}

func AttachHandlerSendOperation(pctx context.Context) error {
	var log *logging.Logging
	var params *launch.LocalParams
	var db isaac.Database
	var pool *isaacdatabase.TempPool
	var states *isaacstates.States
	var svVoteF isaac.SuffrageVoteFunc
	var memberList *quicmemberlist.Memberlist

	if err := util.LoadFromContext(pctx,
		launch.LoggingContextKey, &log,
		launch.LocalParamsContextKey, &params,
		launch.CenterDatabaseContextKey, &db,
		launch.PoolDatabaseContextKey, &pool,
		launch.StatesContextKey, &states,
		launch.SuffrageVotingVoteFuncContextKey, &svVoteF,
		launch.MemberlistContextKey, &memberList,
	); err != nil {
		return err
	}

	sendOperationFilterF, err := SendOperationFilterFunc(pctx)
	if err != nil {
		return err
	}

	var gerror error

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameSendOperation,
		isaacnetwork.QuicstreamHandlerSendOperation(
			params.ISAAC.NetworkID(),
			pool,
			db.ExistsInStateOperation,
			sendOperationFilterF,
			svVoteF,
			func(ctx context.Context, id string, op base.Operation, b []byte) error {
				if broker := states.HandoverXBroker(); broker != nil {
					if err := broker.SendData(ctx, isaacstates.HandoverMessageDataTypeOperation, op); err != nil {
						log.Log().Error().Err(err).
							Interface("operation", op.Hash()).
							Msg("send operation data to handover y broker; ignored")
					}
				}

				return memberList.CallbackBroadcast(b, id, nil)
			},
			params.MISC.MaxMessageSize,
		),
		nil,
	)

	return gerror
}

func SendOperationFilterFunc(ctx context.Context) (
	func(base.Operation) (bool, error),
	error,
) {
	var db isaac.Database
	var oprs *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc]
	var oprsB *hint.CompatibleSet[contracts.NewOperationProcessorInternalWithProposalFunc]
	var f contracts.ProposalOperationFactHintFunc

	if err := util.LoadFromContextOK(ctx,
		launch.CenterDatabaseContextKey, &db,
		launch.OperationProcessorsMapContextKey, &oprs,
		contracts.OperationProcessorsMapBContextKey, &oprsB,
		contracts.ProposalOperationFactHintContextKey, &f,
	); err != nil {
		return nil, err
	}

	operationFilterF := f()

	return func(op base.Operation) (bool, error) {
		switch hinter, ok := op.Fact().(hint.Hinter); {
		case !ok:
			return false, nil
		case !operationFilterF(hinter.Hint()):
			return false, errors.Errorf("Not supported operation")
		}
		var height base.Height

		switch m, found, err := db.LastBlockMap(); {
		case err != nil:
			return false, err
		case !found:
			return true, nil
		default:
			height = m.Manifest().Height()
		}

		f, closeF, err := OperationPreProcess(db, oprs, oprsB, op, height)
		if err != nil {
			return false, err
		}

		defer func() {
			_ = closeF()
		}()

		_, reason, err := f(context.Background(), db.State)
		if err != nil {
			return false, err
		}

		return reason == nil, reason
	}, nil
}

func OperationPreProcess(
	db isaac.Database,
	oprsA *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc],
	oprsB *hint.CompatibleSet[contracts.NewOperationProcessorInternalWithProposalFunc],
	op base.Operation,
	height base.Height,
) (
	preprocess func(context.Context, base.GetStateFunc) (context.Context, base.OperationProcessReasonError, error),
	cancel func() error,
	_ error,
) {
	fA, foundA := oprsA.Find(op.Hint())
	fB, foundB := oprsB.Find(op.Hint())
	if !foundA && !foundB {
		return op.PreProcess, util.EmptyCancelFunc, nil
	}

	if foundA {
		switch opp, err := fA(height, db.State); {
		case err != nil:
			return nil, nil, err
		default:
			return func(pctx context.Context, getStateFunc base.GetStateFunc) (
				context.Context, base.OperationProcessReasonError, error,
			) {
				return opp.PreProcess(pctx, op, getStateFunc)
			}, opp.Close, nil
		}
	}
	switch opp, err := fB(height, nil, db.State); {
	case err != nil:
		return nil, nil, err
	default:
		return func(pctx context.Context, getStateFunc base.GetStateFunc) (
			context.Context, base.OperationProcessReasonError, error,
		) {
			return opp.PreProcess(pctx, op, getStateFunc)
		}, opp.Close, nil
	}

}
