package did_registry

import (
	"context"
	"sync"

	"github.com/imfact-labs/currency-model/common"

	"github.com/imfact-labs/currency-model/state"
	dstate "github.com/imfact-labs/currency-model/state/did-registry"
	cestate "github.com/imfact-labs/currency-model/state/extension"
	"github.com/imfact-labs/currency-model/types"
	ctypes "github.com/imfact-labs/currency-model/types"

	ccstate "github.com/imfact-labs/currency-model/state/currency"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/pkg/errors"
)

var registerModelProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(RegisterModelProcessor)
	},
}

type RegisterModelProcessor struct {
	*base.BaseOperationProcessor
}

func NewRegisterModelProcessor() ctypes.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new RegisterModelProcessor")

		nopp := registerModelProcessorPool.Get()
		opp, ok := nopp.(*RegisterModelProcessor)
		if !ok {
			return nil, errors.Errorf("expected RegisterModelProcessor, not %T", nopp)
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		opp.BaseOperationProcessor = b

		return opp, nil
	}
}

func (opp *RegisterModelProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(RegisterModelFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", RegisterModelFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	if err := state.CheckExistsState(ccstate.DesignStateKey(fact.Currency()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCurrencyNF).Errorf("currency id, %v", fact.Currency())), nil
	}

	if found, _ := state.CheckNotExistsState(dstate.DesignStateKey(fact.Contract()), getStateFunc); found {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceE).Errorf("did service for contract account %v",
				fact.Contract(),
			)), nil
	}

	return ctx, nil, nil
}

func (opp *RegisterModelProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process RegisterModel")

	fact, ok := op.Fact().(RegisterModelFact)
	if !ok {
		return nil, nil, e.Errorf("expected RegisterModelFact, not %T", op.Fact())
	}

	var sts []base.StateMergeValue

	design := types.NewDesign(fact.DIDMethod())
	if err := design.IsValid(nil); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("invalid did design, %q; %w", fact.Contract(), err), nil
	}

	sts = append(sts, state.NewStateMergeValue(
		dstate.DesignStateKey(fact.Contract()),
		dstate.NewDesignStateValue(design),
	))

	st, err := state.ExistsState(cestate.StateKeyContractAccount(fact.Contract()), "contract account", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("contract account not found, %q; %w", fact.Contract(), err), nil
	}

	ca, err := cestate.StateContractAccountValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to get state value of contract account, %q; %w", fact.Contract(), err), nil
	}
	ca.SetActive(true)
	h := op.Hint()
	ca.SetRegisterOperation(&h)

	sts = append(sts, state.NewStateMergeValue(
		cestate.StateKeyContractAccount(fact.Contract()),
		cestate.NewContractAccountStateValue(ca),
	))

	return sts, nil, nil
}

func (opp *RegisterModelProcessor) Close() error {
	registerModelProcessorPool.Put(opp)

	return nil
}
