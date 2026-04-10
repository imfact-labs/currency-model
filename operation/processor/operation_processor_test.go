package processor_test

import (
	"context"
	"testing"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/operation/currency"
	"github.com/imfact-labs/currency-model/operation/extras"
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
	setCurrencyDesign(tp, cid, design)
}

func setCurrencyDesign(tp *operationtest.TestProcessor, cid types.CurrencyID, design types.CurrencyDesign) {

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
	fee, ok := receipt.Fee.(types.FixedFeeReceipt)
	if !ok {
		t.Fatalf("unexpected fee receipt type: %T", receipt.Fee)
	}

	if fee.CurrencyID != tp.GenesisCurrency || fee.Amount != "10" || fee.BaseAmount != "10" {
		t.Fatalf("unexpected fee receipt: %+v", fee)
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

func TestOperationProcessorSetsDetailedFeeReceiptForFixedItemDataSizeExecutionFeeer(t *testing.T) {
	getter := operationtest.NewMockStateGetter()
	var tp operationtest.TestProcessor
	tp.Setup(getter)

	senderPrivSeed := tp.NewPrivateKey("sender-detailed")
	sender, _, senderPriv := tp.NewTestAccountState(senderPrivSeed, true)
	tp.NewTestBalanceState(sender, tp.GenesisCurrency, 1000, true)

	receiverPrivSeed := tp.NewPrivateKey("receiver-detailed")
	receiver, _, _ := tp.NewTestAccountState(receiverPrivSeed, true)

	feeer := types.NewFixedItemDataSizeExecutionFeeer(
		tp.GenesisAddr,
		common.NewBig(10),
		common.NewBig(2),
		common.NewBig(3),
		100000,
		common.NewBig(5),
	)
	design := types.NewCurrencyDesign(
		common.ZeroBig,
		tp.GenesisCurrency,
		common.NewBig(9),
		tp.GenesisAddr,
		types.NewCurrencyPolicy(common.ZeroBig, feeer),
	)
	setCurrencyDesign(&tp, tp.GenesisCurrency, design)

	opr := newWrappedProcessor(t, tp.GetStateFunc)

	item := currency.NewTransferItemMultiAmounts(receiver, []types.Amount{
		types.NewAmount(common.NewBig(100), tp.GenesisCurrency),
	})

	transferOp, err := currency.NewTransfer(currency.NewTransferFact(
		[]byte("transfer-detailed"),
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

	feeable, ok := transferOp.Fact().(extras.FeeAble)
	if !ok {
		t.Fatalf("transfer fact is not feeable: %T", transferOp.Fact())
	}

	_, itemCount, dataSize, _ := feeable.FeeBase()
	expectedItemFee := feeer.ItemFee(itemCount)
	expectedDataSizeFee := feeer.DataSizeFee(dataSize)
	expectedExecutionFee := feeer.ExecutionFee()
	expectedTotal := feeer.Fee().Add(expectedItemFee).Add(expectedDataSizeFee).Add(expectedExecutionFee)

	receipt := receiptAsCurrency(t, opr.OperationReceipt())
	fee, ok := receipt.Fee.(types.FixedItemDataSizeExecutionFeeReceipt)
	if !ok {
		t.Fatalf("unexpected fee receipt type: %T", receipt.Fee)
	}

	if fee.CurrencyID != tp.GenesisCurrency || fee.TotalAmount != expectedTotal.String() || fee.BaseAmount != feeer.Fee().String() {
		t.Fatalf("unexpected total/base fee receipt: %+v", fee)
	}

	if fee.ItemCount != itemCount || fee.ItemFeeAmount != common.NewBig(2).String() || fee.ItemFee != expectedItemFee.String() {
		t.Fatalf("unexpected item fee receipt: %+v", fee)
	}

	if fee.DataSize != dataSize || fee.DataSizeUnit != feeer.DataSizeUnit() ||
		fee.DataSizeFeeAmount != common.NewBig(3).String() || fee.DataSizeFee != expectedDataSizeFee.String() {
		t.Fatalf("unexpected data size fee receipt: %+v", fee)
	}

	if fee.ExecutionFeeAmount != common.NewBig(5).String() || fee.ExecutionFee != expectedExecutionFee.String() {
		t.Fatalf("unexpected execution fee receipt: %+v", fee)
	}
}
