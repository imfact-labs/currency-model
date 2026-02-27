package processor

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/operation/currency"
	"github.com/imfact-labs/currency-model/operation/extras"
	isaacoperation "github.com/imfact-labs/currency-model/operation/isaac"
	"github.com/imfact-labs/currency-model/state"
	ccstate "github.com/imfact-labs/currency-model/state/currency"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/logging"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var operationProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(OperationProcessor)
	},
}

type GetLastBlockFunc func() (base.BlockMap, bool, error)

type BaseOperationProcessor interface {
	PreProcess(base.Operation, base.GetStateFunc) (base.OperationProcessReasonError, error)
	Process(base.Operation, base.GetStateFunc) ([]base.StateMergeValue, base.OperationProcessReasonError, error)
	Close() error
}

type OperationProcessor struct {
	// id string
	sync.RWMutex
	*logging.Logging
	*base.BaseOperationProcessor
	processorHintSet             *hint.CompatibleSet[types.GetNewProcessor]
	processorHintSetWithProposal *hint.CompatibleSet[types.GetNewProcessorWithProposal]
	Duplicated                   map[string]struct{}
	duplicatedNewAddress         map[string]struct{}
	processorClosers             *sync.Map
	proposal                     *base.ProposalSignFact
	GetStateFunc                 base.GetStateFunc
	CollectFee                   func(*OperationProcessor, types.AddFee) error
	CheckDuplicationFunc         func(*OperationProcessor, base.Operation) error
	GetNewProcessorFunc          func(*OperationProcessor, base.Operation) (base.OperationProcessor, bool, error)
}

func NewOperationProcessor() *OperationProcessor {
	m := sync.Map{}
	return &OperationProcessor{
		// id: util.UUID().String(),
		Logging: logging.NewLogging(func(c zerolog.Context) zerolog.Context {
			return c.Str("module", "mitum-currency-operations-processor")
		}),
		processorHintSet:             hint.NewCompatibleSet[types.GetNewProcessor](1 << 9),
		processorHintSetWithProposal: hint.NewCompatibleSet[types.GetNewProcessorWithProposal](1 << 9),
		Duplicated:                   map[string]struct{}{},
		duplicatedNewAddress:         map[string]struct{}{},
		processorClosers:             &m,
	}
}

func (opr *OperationProcessor) New(
	height base.Height,
	getStateFunc base.GetStateFunc,
	newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	newProcessConstraintFunc base.NewOperationProcessorProcessFunc) (*OperationProcessor, error) {
	e := util.StringError("create new OperationProcessor")

	nopr := operationProcessorPool.Get().(*OperationProcessor)
	if nopr.processorHintSet == nil {
		nopr.processorHintSet = opr.processorHintSet
	}

	if nopr.processorHintSetWithProposal == nil {
		nopr.processorHintSetWithProposal = opr.processorHintSetWithProposal
	}

	if nopr.Duplicated == nil {
		nopr.Duplicated = make(map[string]struct{})
	}

	if nopr.proposal == nil && opr.proposal != nil {
		nopr.proposal = opr.proposal
	}

	if nopr.duplicatedNewAddress == nil {
		nopr.duplicatedNewAddress = make(map[string]struct{})
	}

	if nopr.Logging == nil {
		nopr.Logging = opr.Logging
	}

	b, err := base.NewBaseOperationProcessor(
		height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
	if err != nil {
		return nil, e.Wrap(err)
	}

	nopr.BaseOperationProcessor = b
	nopr.GetStateFunc = getStateFunc
	nopr.CheckDuplicationFunc = opr.CheckDuplicationFunc
	nopr.GetNewProcessorFunc = opr.GetNewProcessorFunc
	return nopr, nil
}

func (opr *OperationProcessor) SetProcessor(
	hint hint.Hint,
	newProcessor types.GetNewProcessor,
) error {
	if err := opr.processorHintSet.Add(hint, newProcessor); err != nil {
		if !errors.Is(err, util.ErrFound) {
			return err
		}
	}

	return nil
}

func (opr *OperationProcessor) SetProcessorWithProposal(
	hint hint.Hint,
	newProcessor types.GetNewProcessorWithProposal,
) error {
	if err := opr.processorHintSetWithProposal.Add(hint, newProcessor); err != nil {
		if !errors.Is(err, util.ErrFound) {
			return err
		}
	}

	return nil
}

func (opr *OperationProcessor) SetProposal(
	proposal *base.ProposalSignFact,
) error {
	if proposal == nil {
		return errors.Errorf("Set nil proposal to OperationProcessor")
	}
	opr.proposal = proposal

	return nil
}

func (opr *OperationProcessor) GetProposal() *base.ProposalSignFact {
	return opr.proposal
}

func (opr *OperationProcessor) SetCheckDuplicationFunc(
	f func(*OperationProcessor, base.Operation) error,
) error {
	if f == nil {
		return errors.Errorf("Set nil func to CheckDuplicationFunc")
	}
	opr.CheckDuplicationFunc = f

	return nil
}

func (opr *OperationProcessor) SetGetNewProcessorFunc(
	f func(*OperationProcessor, base.Operation) (base.OperationProcessor, bool, error),
) error {
	if f == nil {
		return errors.Errorf("Set nil func to GetNewProcessorFunc")
	}
	opr.GetNewProcessorFunc = f

	return nil
}

func (opr *OperationProcessor) PreProcess(ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringError("preprocess for OperationProcessor")

	if err := opr.CheckDuplicationFunc(opr, op); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("duplication found; %w", err), nil
	}

	if opr.processorClosers == nil {
		opr.processorClosers = &sync.Map{}
	}

	var opp base.OperationProcessor

	if opr.GetNewProcessorFunc == nil {
		return ctx, nil, e.Errorf("GetNewProcessorFunc is nil")
	}
	switch i, known, err := opr.GetNewProcessorFunc(opr, op); {
	case err != nil:
		return ctx, base.NewBaseOperationProcessReasonError(err.Error()), nil
	case !known:
		return ctx, nil, e.Errorf("getNewProcessor, %T", op)
	default:
		opp = i
	}

	if extOp, ok := op.(extras.OperationExtensions); ok {
		auth := extOp.Extension(extras.AuthenticationExtensionType)
		settlement := extOp.Extension(extras.SettlementExtensionType)
		if settlement != nil && auth != nil {
			if err := extOp.Verify(op, getStateFunc); err != nil {
				return ctx, base.NewBaseOperationProcessReasonError(
					common.ErrMPreProcess.Errorf("%v", err)), nil
			}
		} else {
			fact := op.Fact()
			signerFact, ok := fact.(currency.Signer)
			if ok {
				if err := state.CheckFactSignsByState(signerFact.Signer(), op.Signs(), getStateFunc); err != nil {
					return ctx,
						base.NewBaseOperationProcessReasonError(
							common.ErrMPreProcess.
								Wrap(common.ErrMSignInvalid).
								Errorf("%v", err),
						), nil
				}
			} else {
				return ctx,
					base.NewBaseOperationProcessReasonError(
						common.ErrMPreProcess.Wrap(common.ErrMTypeMismatch).
							Errorf("expected Signer but %T", fact)), nil
			}
		}
	}

	if fact, ok := op.Fact().(extras.FeeAble); ok {
		if err := extras.VerifyFeeAble(fact, getStateFunc); err != nil {
			return ctx, err, nil
		}
	}

	if fact, ok := op.Fact().(extras.FactUser); ok {
		if err := extras.VerifyFactUser(fact, getStateFunc); err != nil {
			return ctx, err, nil
		}
	}

	if fact, ok := op.Fact().(extras.InActiveContractOwnerHandlerOnly); ok {
		if err := extras.VerifyInActiveContractOwnerHandlerOnly(fact, getStateFunc); err != nil {
			return ctx, err, nil
		}
	}

	if fact, ok := op.Fact().(extras.ActiveContractOwnerHandlerOnly); ok {
		if err := extras.VerifyActiveContractOwnerHandlerOnly(fact, getStateFunc); err != nil {
			return ctx, err, nil
		}
	}

	if fact, ok := op.Fact().(extras.ContractOwnerOnly); ok {
		if err := extras.VerifyContractOwnerOnly(fact, getStateFunc); err != nil {
			return ctx, err, nil
		}
	}

	if fact, ok := op.Fact().(extras.ActiveContract); ok {
		if err := extras.VerifyActiveContract(fact, getStateFunc); err != nil {
			return ctx, err, nil
		}
	}

	switch _, reasonErr, err := opp.PreProcess(ctx, op, getStateFunc); {
	case err != nil:
		return ctx, nil, e.Wrap(err)
	case reasonErr != nil:
		return ctx, reasonErr, nil
	}

	return ctx, nil, nil
}

func (opr *OperationProcessor) Process(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	e := util.StringError("process for OperationProcessor")

	var sp base.OperationProcessor
	if opr.GetNewProcessorFunc == nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Errorf("failed to GetNewProcessorFunc")), nil
	}

	switch i, known, err := opr.GetNewProcessorFunc(opr, op); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Errorf("%v", err)), nil
	case !known:
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Errorf("getNewProcessor for op %T", op)), nil
	default:
		sp = i
	}

	stateMergeValues, reasonErr, err := sp.Process(ctx, op, getStateFunc)
	if reasonErr != nil {
		return nil, reasonErr, nil
	}
	if err != nil {
		return nil, nil, e.Wrap(err)
	}

	var payer base.Address
	switch i := op.Fact().(type) {
	case extras.FeeAble:
		feeBase := i.FeeBase()
		payer = i.FeePayer()
		switch k := op.(type) {
		case extras.OperationExtensions:
			iAuth := k.Extension(extras.AuthenticationExtensionType)
			iSettlement := k.Extension(extras.SettlementExtensionType)
			iProxyPayer := k.Extension(extras.ProxyPayerExtensionType)
			if iAuth != nil && iSettlement != nil {
				settlement, ok := iSettlement.(extras.Settlement)
				if !ok {
					return nil, base.NewBaseOperationProcessReasonError(
						common.ErrMPreProcess.Wrap(common.ErrMTypeMismatch).
							Errorf("expected Settlement, but %T", iSettlement)), nil
				}
				opSender := settlement.OpSender()
				if opSender == nil {
					return nil, base.NewBaseOperationProcessReasonError(
						common.ErrMPreProcess.
							Errorf("failed to get op sender, empty op sender")), nil
				}
				payer = opSender
			}
			if iProxyPayer != nil {
				proxyPayer, ok := iProxyPayer.(extras.ProxyPayer)
				if !ok {
					return nil, base.NewBaseOperationProcessReasonError(
						common.ErrMPreProcess.Wrap(common.ErrMTypeMismatch).
							Errorf("expected ProxyPayer, but %T", iProxyPayer)), nil
				}

				if proxyPayer := proxyPayer.ProxyPayer(); proxyPayer != nil {
					payer = proxyPayer
				}
			}
		default:
		}

		feeReceiveSts := map[types.CurrencyID]base.State{}
		var feeRequired = make(map[types.CurrencyID]common.Big)
		var policy *types.CurrencyPolicy
		var err error

		for cid, amounts := range feeBase {
			policy, err = state.ExistsCurrencyPolicy(cid, getStateFunc)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError(
					common.ErrMPreProcess.
						Errorf("%v", err)), nil
			}
			receiver := policy.Feeer().Receiver()
			if receiver == nil {
				continue
			}

			if err := state.CheckExistsState(ccstate.AccountStateKey(receiver), getStateFunc); err != nil {
				return nil, base.NewBaseOperationProcessReasonError(
						common.ErrMAccountNF.Errorf("Feeer receiver, %v", receiver)),
					nil
			} else if st, found, err := getStateFunc(ccstate.BalanceStateKey(receiver, cid)); err != nil {
				return nil, base.NewBaseOperationProcessReasonError(
						common.ErrMStateNF.Errorf("Feeer receiver, %v BalanceState: %v", receiver, err)),
					nil
			} else if !found {
				return nil, base.NewBaseOperationProcessReasonError(
						common.ErrMStateNF.Errorf("Feeer receiver, %v BalanceState", receiver)),
					nil
			} else {
				feeReceiveSts[cid] = st
			}

			rq := common.ZeroBig
			total := common.ZeroBig
			for _, big := range amounts {
				total = total.Add(big)
				f, ok := policy.Feeer().(types.ItemFeeer)
				if ok {
					itmFee, _ := f.ItemFee(big)
					rq = rq.Add(itmFee)
				} else {
					switch bsFee, err := policy.Feeer().Fee(big); {
					case err != nil:
						return nil,
							base.NewBaseOperationProcessReasonError("check fee of currency %v; %w", cid, err),
							nil
					default:
						rq = rq.Add(bsFee)
					}
				}
			}
			_, ok := policy.Feeer().(types.ItemFeeer)
			if ok {
				switch bsFee, err := policy.Feeer().Fee(total); {
				case err != nil:
					return nil,
						base.NewBaseOperationProcessReasonError("check fee of currency %v; %w", cid, err),
						nil
				default:
					rq = rq.Add(bsFee)
				}
			}
			if v, found := feeRequired[cid]; !found {
				feeRequired[cid] = rq
			} else {
				feeRequired[cid] = v.Add(rq)
			}
		}

		for cid, rq := range feeRequired {
			payerSt, err := state.ExistsState(ccstate.BalanceStateKey(payer, cid), fmt.Sprintf("balance of fee payer, %v", payer), getStateFunc)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError(
						common.ErrMStateNF.Errorf("fee payer, %v BalanceState: %v", payer, err)),
					nil
			}

			payerBalValue, ok := payerSt.Value().(ccstate.BalanceStateValue)
			if !ok {
				return nil, base.NewBaseOperationProcessReasonError(
						common.ErrMPreProcess.Wrap(common.ErrMTypeMismatch).
							Errorf("expected %T, not %T",
								ccstate.BalanceStateValue{},
								payerSt.Value())),
					nil
			}

			feeReceiverSt, feeReceiverFound := feeReceiveSts[cid]
			if feeReceiverFound {
				if payerSt.Key() != feeReceiverSt.Key() {
					stateMergeValues = append(stateMergeValues, common.NewBaseStateMergeValue(
						payerSt.Key(),
						ccstate.NewDeductBalanceStateValue(payerBalValue.Amount.WithBig(rq)),
						func(height base.Height, st base.State) base.StateValueMerger {
							return ccstate.NewBalanceStateValueMerger(height, st.Key(), cid, st)
						},
					))
					r, ok := feeReceiveSts[cid].Value().(ccstate.BalanceStateValue)
					if !ok {
						return nil, base.NewBaseOperationProcessReasonError(
								"expected %T, not %T",
								ccstate.BalanceStateValue{},
								feeReceiveSts[cid].Value()),
							nil
					}
					stateMergeValues = append(
						stateMergeValues,
						common.NewBaseStateMergeValue(
							feeReceiveSts[cid].Key(),
							ccstate.NewAddBalanceStateValue(r.Amount.WithBig(rq)),
							func(height base.Height, st base.State) base.StateValueMerger {
								return ccstate.NewBalanceStateValueMerger(height, feeReceiveSts[cid].Key(), cid, st)
							},
						),
					)
				}
			}
		}
	default:
	}

	reasonErr, err = CheckBalanceStateMergeValue(stateMergeValues, getStateFunc)
	if reasonErr != nil {
		return nil, reasonErr, nil
	}
	if err != nil {
		return nil, nil, e.Wrap(err)
	}

	return stateMergeValues, reasonErr, err
}

type DupKeySet map[types.DuplicationKeyType][]string

func NewDupKeySet() *DupKeySet {
	d := DupKeySet(make(map[types.DuplicationKeyType][]string))
	return &d
}

func (d DupKeySet) Add(t types.DuplicationKeyType, key string) {
	d[t] = append(d[t], DuplicationKey(t, key))
}

func DuplicationKey(dupType types.DuplicationKeyType, key string) string {
	return fmt.Sprintf("%s:%s", key, dupType)
}

func CheckDuplication(opr *OperationProcessor, op base.Operation) error {
	opr.Lock()
	defer opr.Unlock()

	dupKeySet := NewDupKeySet()

	switch t := op.Fact().(type) {
	case extras.DeDupeKeyer:
		dkSet, err := t.DupKey()
		if err != nil {
			return err
		}
		for k, v := range dkSet {
			for _, dk := range v {
				dupKeySet.Add(k, dk)
			}
		}
	case isaacoperation.NetworkPolicyFact, isaacoperation.GenesisNetworkPolicyFact,
		isaacoperation.SuffrageCandidateFact, isaacoperation.SuffrageDisjoinFact,
		isaacoperation.SuffrageGenesisJoinFact, isaacoperation.SuffrageJoinFact,
		base.SuffrageExpelFact:
	default:
		return errors.Errorf(
			"%T not implemented DeDupeKeyer", t,
		)
	}

	pending := make(map[string]struct{})
	for kType, kSet := range *dupKeySet {
		for _, dk := range kSet {
			if _, found := pending[dk]; found {
				return errors.Errorf(
					"cannot use a duplicated %v for %v within one operation",
					dk, kType,
				)
			}
			if _, found := opr.Duplicated[dk]; found {
				return errors.Errorf(
					"cannot use a duplicated %v for %v within a proposal",
					dk, kType,
				)
			}
			pending[dk] = struct{}{}
		}
	}

	for dk := range pending {
		opr.Duplicated[dk] = struct{}{}
	}

	return nil
}

func (opr *OperationProcessor) Close() error {
	opr.Lock()
	defer opr.Unlock()

	defer opr.close()

	return nil
}

func (opr *OperationProcessor) Cancel() error {
	opr.Lock()
	defer opr.Unlock()

	defer opr.close()

	return nil
}

func GetNewProcessor(opr *OperationProcessor, op base.Operation) (base.OperationProcessor, bool, error) {
	switch i, err := opr.GetNewProcessorFromHintset(op); {
	case err != nil:
		return nil, false, err
	case i != nil:
		return i, true, nil
	}

	return nil, false, errors.Errorf("%T needs SetProcessor", op)
}

func (opr *OperationProcessor) GetNewProcessorFromHintset(op base.Operation) (base.OperationProcessor, error) {
	var fA types.GetNewProcessor
	var fB types.GetNewProcessorWithProposal
	var iA types.GetNewProcessor
	var iB types.GetNewProcessorWithProposal
	var foundA, foundB bool
	if hinter, ok := op.(hint.Hinter); !ok {
		return nil, nil
	} else if iA, foundA = opr.processorHintSet.Find(hinter.Hint()); foundA {
		fA = iA
	} else if iB, foundB = opr.processorHintSetWithProposal.Find(hinter.Hint()); foundB {
		fB = iB
	} else {
		return nil, nil
	}

	var opp base.OperationProcessor
	var err error
	if foundA {
		opp, err = fA(opr.Height(), opr.GetStateFunc, nil, nil)
	}
	if foundB {
		opp, err = fB(opr.Height(), opr.proposal, opr.GetStateFunc, nil, nil)
	}

	if err != nil {
		return nil, err
	}

	h := op.(util.Hasher).Hash().String()
	_, isCloser := opp.(io.Closer)
	if isCloser {
		opr.processorClosers.Store(h, opp)
		isCloser = true
	}

	opr.Log().Debug().
		Str("operation", h).
		Str("processor", fmt.Sprintf("%T", opp)).
		Bool("is_closer", isCloser).
		Msg("operation processor created")

	return opp, nil
}

func (opr *OperationProcessor) close() {
	opr.processorClosers.Range(func(_, v interface{}) bool {
		err := v.(io.Closer).Close()
		if err != nil {
			opr.Log().Error().Err(err).Str("op", fmt.Sprintf("%T", v)).Msg("close operation processor")
		} else {
			opr.Log().Debug().Str("processor", fmt.Sprintf("%T", v)).Msg("operation processor closed")
		}

		return true
	})

	//opr.pool = nil
	opr.proposal = nil
	opr.Duplicated = nil
	opr.duplicatedNewAddress = nil
	opr.processorClosers = &sync.Map{}

	operationProcessorPool.Put(opr)

	opr.Log().Debug().Msg("operation processors closed")
}

func CheckBalanceStateMergeValue(stateMergeValues []base.StateMergeValue, getStateFunc base.GetStateFunc) (base.OperationProcessReasonError, error) {
	type BalanceValue struct {
		address  string
		add      common.Big
		remove   common.Big
		currency types.CurrencyID
	}

	balanceValues := make(map[string]BalanceValue)
	for i := range stateMergeValues {
		if ccstate.IsBalanceStateKey(stateMergeValues[i].Key()) {
			parsed, err := ccstate.ParseBalanceStateKey(stateMergeValues[i].Key())
			if err != nil {
				return nil, err
			}
			bv, found := balanceValues[stateMergeValues[i].Key()]
			if !found {
				bv = BalanceValue{
					address:  parsed[0],
					add:      common.ZeroBig,
					remove:   common.ZeroBig,
					currency: types.CurrencyID(parsed[1]),
				}
			}
			switch t := stateMergeValues[i].Value().(type) {
			case ccstate.AddBalanceStateValue:
				bv.add = bv.add.Add(t.Amount.Big())
			case ccstate.DeductBalanceStateValue:
				bv.remove = bv.remove.Add(t.Amount.Big())
			default:
				return nil, errors.Errorf("Unsupported balance state value, %T", stateMergeValues[i].Value())
			}

			balanceValues[stateMergeValues[i].Key()] = bv
		}
	}

	for stk, bv := range balanceValues {
		switch st, _, err := getStateFunc(stk); {
		case err != nil:
			return nil, err
		default:
			var existing common.Big
			var amount common.Big
			if st == nil {
				existing = common.ZeroBig
			} else if st.Value() != nil {
				value, ok := st.Value().(ccstate.BalanceStateValue)
				if ok {
					existing = value.Amount.Big()
					amount = value.Amount.Big()
				} else {
					return nil, errors.Errorf("expected BalanceStateValue, but %T", st.Value())
				}
			}

			if bv.add.OverZero() {
				existing = existing.Add(bv.add)
			}
			if bv.remove.OverZero() {
				existing = existing.Sub(bv.remove)
			}
			if !existing.OverNil() {
				return base.NewBaseOperationProcessReasonError(
					"account, %s balance insufficient; %d < required %d", bv.address, amount, amount.Sub(existing)), nil
			}
		}
	}

	return nil, nil
}
