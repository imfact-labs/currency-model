package processor_test

import (
	"context"
	"testing"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/operation/currency"
	"github.com/imfact-labs/currency-model/operation/processor"
	operationtest "github.com/imfact-labs/currency-model/operation/test"
	ccstate "github.com/imfact-labs/currency-model/state/currency"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
)

func newWrappedProcessor(t *testing.T, getStateFunc base.GetStateFunc) *processor.OperationProcessor {
	t.Helper()

	root := processor.NewOperationProcessor()

	if err := root.SetCheckDuplicationFunc(processor.CheckDuplication); err != nil {
		t.Fatalf("set duplication func: %v", err)
	}

	if err := root.SetGetNewProcessorFunc(processor.GetNewProcessor); err != nil {
		t.Fatalf("set new processor func: %v", err)
	}

	if err := root.SetProcessor(currency.TransferHint, currency.NewTransferProcessor()); err != nil {
		t.Fatalf("set transfer processor: %v", err)
	}

	if err := root.SetProcessor(currency.MintHint, currency.NewMintProcessor(base.MaxThreshold)); err != nil {
		t.Fatalf("set mint processor: %v", err)
	}

	opr, err := root.New(base.GenesisHeight, getStateFunc, nil, nil)
	if err != nil {
		t.Fatalf("new wrapped processor: %v", err)
	}

	return opr
}

func setFixedFeeer(tp *operationtest.TestProcessor, cid types.CurrencyID, receiver base.Address, fee int64) {
	design := types.NewCurrencyDesign(
		common.ZeroBig,
		cid,
		common.NewBig(9),
		receiver,
		types.NewCurrencyPolicy(common.ZeroBig, types.NewFixedFeeer(receiver, common.NewBig(fee))),
	)

	st := common.NewBaseState(
		base.Height(1),
		ccstate.DesignStateKey(cid),
		ccstate.NewCurrencyDesignStateValue(design),
		nil,
		[]util.Hash{},
	)

	tp.SetState(st, true)
}

func receiptAsCurrency(t *testing.T, receipt base.OperationReceipt) types.CurrencyOperationReceipt {
	t.Helper()

	switch r := receipt.(type) {
	case types.CurrencyOperationReceipt:
		return r
	case *types.CurrencyOperationReceipt:
		if r == nil {
			t.Fatal("nil currency receipt pointer")
		}

		return *r
	default:
		t.Fatalf("unexpected receipt type: %T", receipt)

		return types.CurrencyOperationReceipt{}
	}
}

func TestOperationProcessorSetsAndResetsReceipt(t *testing.T) {
	getter := operationtest.NewMockStateGetter()
	var tp operationtest.TestProcessor
	tp.Setup(getter)

	senderPrivSeed := tp.NewPrivateKey("sender")
	sender, _, senderPriv := tp.NewTestAccountState(senderPrivSeed, true)
	tp.NewTestBalanceState(sender, tp.GenesisCurrency, 1000, true)

	receiverPrivSeed := tp.NewPrivateKey("receiver")
	receiver, _, _ := tp.NewTestAccountState(receiverPrivSeed, true)

	setFixedFeeer(&tp, tp.GenesisCurrency, tp.GenesisAddr, 10)

	opr := newWrappedProcessor(t, tp.GetStateFunc)

	item := currency.NewTransferItemMultiAmounts(receiver, []types.Amount{
		types.NewAmount(common.NewBig(100), tp.GenesisCurrency),
	})

	transferOp, err := currency.NewTransfer(currency.NewTransferFact(
		[]byte("transfer"),
		sender,
		[]currency.TransferItem{item},
		tp.GenesisCurrency,
	))
	if err != nil {
		t.Fatalf("new transfer: %v", err)
	}

	if err := transferOp.Sign(senderPriv, tp.NetworkID); err != nil {
		t.Fatalf("sign transfer: %v", err)
	}

	states, reason, err := opr.Process(context.Background(), transferOp, tp.GetStateFunc)
	if err != nil {
		t.Fatalf("process transfer: %v", err)
	}

	if reason != nil {
		t.Fatalf("unexpected transfer reason: %v", reason)
	}

	if len(states) == 0 {
		t.Fatal("expected transfer state merge values")
	}

	receipt := receiptAsCurrency(t, opr.OperationReceipt())
	if receipt.Fee == nil || receipt.Fee.CurrencyID != tp.GenesisCurrency || receipt.Fee.Amount != "10" {
		t.Fatalf("unexpected fee receipt: %+v", receipt.Fee)
	}

	if receipt.GasUsed != nil {
		t.Fatalf("unexpected gas used: %v", receipt.GasUsed)
	}

	mintOp, err := currency.NewMint(currency.NewMintFact(
		[]byte("mint"),
		tp.GenesisAddr,
		types.NewAmount(common.NewBig(5), tp.GenesisCurrency),
	))
	if err != nil {
		t.Fatalf("new mint: %v", err)
	}

	if err := mintOp.NodeSign(tp.NodePriv, tp.NetworkID, tp.NodeAddr); err != nil {
		t.Fatalf("sign mint: %v", err)
	}

	states, reason, err = opr.Process(context.Background(), mintOp, tp.GetStateFunc)
	if err != nil {
		t.Fatalf("process mint: %v", err)
	}

	if reason != nil {
		t.Fatalf("unexpected mint reason: %v", reason)
	}

	if len(states) == 0 {
		t.Fatal("expected mint state merge values")
	}

	if opr.OperationReceipt() != nil {
		t.Fatalf("expected receipt reset after non-fee operation, got %T", opr.OperationReceipt())
	}
}
