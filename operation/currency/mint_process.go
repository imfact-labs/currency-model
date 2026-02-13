package currency

import (
	"context"
	"sync"

	"github.com/imfact-labs/imfact-currency/common"
	"github.com/imfact-labs/imfact-currency/state"
	"github.com/imfact-labs/imfact-currency/state/currency"
	"github.com/imfact-labs/imfact-currency/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/util"
)

var mintProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(MintProcessor)
	},
}

type MintProcessor struct {
	*base.BaseOperationProcessor
	suffrage  base.Suffrage
	threshold base.Threshold
}

func NewMintProcessor(threshold base.Threshold) types.GetNewProcessor {
	return func(height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("create new MintProcessor")

		nopp := mintProcessorPool.Get()
		opp, ok := nopp.(*MintProcessor)
		if !ok {
			return nil, e.Errorf("expected %T, not %T", &MintProcessor{}, nopp)
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		opp.BaseOperationProcessor = b
		opp.threshold = threshold

		switch i, found, err := getStateFunc(isaac.SuffrageStateKey); {
		case err != nil:
			return nil, e.Wrap(err)
		case !found, i == nil:
			return nil, e.Errorf("Empty state")
		default:
			sufstv := i.Value().(base.SuffrageNodesStateValue) //nolint:forcetypeassert //...

			suf, err := sufstv.Suffrage()
			if err != nil {
				return nil, e.Errorf("get suffrage from state")
			}

			opp.suffrage = suf
		}

		return opp, nil
	}
}

func (opp *MintProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	nop, ok := op.(Mint)
	if !ok {
		return ctx,
			base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Wrap(common.ErrMTypeMismatch).
					Errorf("expected %T, not %T", Mint{}, op),
			),
			nil
	}

	fact, ok := op.Fact().(MintFact)
	if !ok {
		return ctx,
			base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Wrap(common.ErrMTypeMismatch).
					Errorf("expected %T, not %T", MintFact{}, op.Fact()),
			),
			nil
	}

	if err := base.CheckFactSignsBySuffrage(opp.suffrage, opp.threshold, nop.NodeSigns()); err != nil {
		return ctx,
			base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Wrap(common.ErrMSignInvalid).
					Errorf("%v", common.ErrSignNE),
			), nil
	}

	_, err := state.ExistsCurrencyPolicy(fact.Amount().Currency(), getStateFunc)
	if err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err),
		), nil
	}

	if _, _, _, cErr := state.ExistsCAccount(
		fact.Receiver(), "receiver", true, false, getStateFunc); cErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCAccountNA).Errorf("%v: receiver %v is contract account", cErr, fact.Receiver())), nil
	}

	return ctx, nil, nil
}

func (opp *MintProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	e := util.StringError("process Mint")

	fact, ok := op.Fact().(MintFact)
	if !ok {
		return nil, nil, e.Errorf("expected %T, not %T", MintFact{}, op.Fact())
	}

	var sts []base.StateMergeValue

	smv, err := state.CreateNotExistAccount(fact.Receiver(), getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"%w", err), nil
	} else if smv != nil {
		sts = append(sts, smv)
	}

	cid := fact.Amount().Currency()
	k := currency.BalanceStateKey(fact.Receiver(), cid)
	switch st, found, err := getStateFunc(k); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError(
			"find receiver account balance state, %v; %w", k, err), nil
	//case !found:
	//	ab = types.NewZeroAmount(item.Amount().Currency())
	case found:
		_, err := currency.StateBalanceValue(st)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("get balance value, %v: %w", k, err), nil
		}
	}

	sts = append(sts, common.NewBaseStateMergeValue(
		k,
		currency.NewAddBalanceStateValue(types.NewAmount(fact.Amount().Big(), cid)),
		func(height base.Height, st base.State) base.StateValueMerger {
			return currency.NewBalanceStateValueMerger(
				height,
				k,
				cid,
				st,
			)
		},
	))

	var de types.CurrencyDesign

	k = currency.DesignStateKey(cid)
	switch st, found, err := getStateFunc(k); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError("find currency design state, %v: %w", cid, err), nil
	case !found:
		return nil, base.NewBaseOperationProcessReasonError("Currency not found, %v: %w", cid, err), nil
	default:
		d, err := currency.GetDesignFromState(st)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("get currency design value, %v: %w", cid, err), nil
		}
		de = d
	}

	ade, err := de.AddTotalSupply(fact.Amount().Big())
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("add aggregate, %v: %w", cid, err), nil
	}

	sts = append(sts, state.NewStateMergeValue(k, currency.NewCurrencyDesignStateValue(ade)))

	return sts, nil, nil
}

func (opp *MintProcessor) Close() error {
	opp.suffrage = nil
	opp.threshold = 0

	mintProcessorPool.Put(opp)

	return nil
}
