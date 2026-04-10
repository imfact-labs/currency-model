package types_test

import (
	"testing"

	"github.com/imfact-labs/currency-model/app/runtime/steps"
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util/encoder"
	jsonenc "github.com/imfact-labs/mitum2/util/encoder/json"
)

func newTestEncoders(t *testing.T) (*encoder.Encoders, *bsonenc.Encoder) {
	t.Helper()

	jenc := jsonenc.NewEncoder()
	encs := encoder.NewEncoders(jenc, jenc)
	benc := bsonenc.NewEncoder()

	if err := encs.AddEncoder(benc); err != nil {
		t.Fatalf("add bson encoder: %v", err)
	}

	if err := steps.LoadHinters(encs); err != nil {
		t.Fatalf("load hinters: %v", err)
	}

	return encs, benc
}

func requireFixedFeeReceipt(t *testing.T, fee types.FeeReceipt, cid types.CurrencyID, amount string) {
	t.Helper()

	var got types.FixedFeeReceipt
	switch r := fee.(type) {
	case types.FixedFeeReceipt:
		got = r
	case *types.FixedFeeReceipt:
		if r == nil {
			t.Fatal("nil fixed fee receipt")
		}

		got = *r
	default:
		t.Fatalf("unexpected fee receipt type: %T", fee)
	}

	if got.CurrencyID != cid || got.Amount != amount || got.BaseAmount != amount {
		t.Fatalf("unexpected fixed fee receipt: %+v", got)
	}
}

func requireFixedItemDataSizeExecutionFeeReceipt(t *testing.T, fee types.FeeReceipt) {
	t.Helper()

	var got types.FixedItemDataSizeExecutionFeeReceipt
	switch r := fee.(type) {
	case types.FixedItemDataSizeExecutionFeeReceipt:
		got = r
	case *types.FixedItemDataSizeExecutionFeeReceipt:
		if r == nil {
			t.Fatal("nil fixed item data size execution fee receipt")
		}

		got = *r
	default:
		t.Fatalf("unexpected fee receipt type: %T", fee)
	}

	if got.CurrencyID != types.CurrencyID("MCC") {
		t.Fatalf("unexpected currency id: %v", got.CurrencyID)
	}

	if got.TotalAmount != "18" || got.BaseAmount != "10" {
		t.Fatalf("unexpected total/base amount: %+v", got)
	}

	if got.ItemCount != 2 || got.ItemFeeAmount != "2" || got.ItemFee != "4" {
		t.Fatalf("unexpected item fee detail: %+v", got)
	}

	if got.DataSize != 30 || got.DataSizeUnit != 10 || got.DataSizeFeeAmount != "1" || got.DataSizeFee != "3" {
		t.Fatalf("unexpected data size fee detail: %+v", got)
	}

	if got.ExecutionFeeAmount != "1" || got.ExecutionFee != "1" {
		t.Fatalf("unexpected execution fee detail: %+v", got)
	}
}

func TestCurrencyOperationReceiptRoundTrip(t *testing.T) {
	encs, benc := newTestEncoders(t)
	gasUsed := uint64(33)

	tests := []struct {
		name   string
		fee    types.FeeReceipt
		assert func(*testing.T, types.FeeReceipt)
	}{
		{
			name: "fixed",
			fee: types.NewFixedFeeReceipt(
				types.CurrencyID("MCC"),
				common.NewBig(10),
			),
			assert: func(t *testing.T, fee types.FeeReceipt) {
				requireFixedFeeReceipt(t, fee, types.CurrencyID("MCC"), "10")
			},
		},
		{
			name: "fixed-item-data-size-execution",
			fee: types.NewFixedItemDataSizeExecutionFeeReceipt(
				types.CurrencyID("MCC"),
				common.NewBig(18),
				common.NewBig(10),
				2,
				common.NewBig(2),
				common.NewBig(4),
				30,
				10,
				common.NewBig(1),
				common.NewBig(3),
				common.NewBig(1),
				common.NewBig(1),
			),
			assert: requireFixedItemDataSizeExecutionFeeReceipt,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name+"-json", func(t *testing.T) {
			receipt := types.NewCurrencyOperationReceipt(tc.fee, &gasUsed)

			b, err := encs.JSON().Marshal(receipt)
			if err != nil {
				t.Fatalf("marshal json: %v", err)
			}

			i, err := encs.JSON().Decode(b)
			if err != nil {
				t.Fatalf("decode json: %v", err)
			}

			got, ok := i.(types.CurrencyOperationReceipt)
			if !ok {
				t.Fatalf("decoded json receipt type = %T", i)
			}

			if err := got.IsValid(nil); err != nil {
				t.Fatalf("validate json receipt: %v", err)
			}

			if got.Fee == nil {
				t.Fatal("expected json fee receipt")
			}

			tc.assert(t, got.Fee)

			if got.GasUsed == nil || *got.GasUsed != gasUsed {
				t.Fatalf("unexpected json gas used: %v", got.GasUsed)
			}
		})

		t.Run(tc.name+"-bson", func(t *testing.T) {
			receipt := types.NewCurrencyOperationReceipt(tc.fee, &gasUsed)

			b, err := benc.Marshal(receipt)
			if err != nil {
				t.Fatalf("marshal bson: %v", err)
			}

			i, err := benc.Decode(b)
			if err != nil {
				t.Fatalf("decode bson: %v", err)
			}

			got, ok := i.(types.CurrencyOperationReceipt)
			if !ok {
				t.Fatalf("decoded bson receipt type = %T", i)
			}

			if err := got.IsValid(nil); err != nil {
				t.Fatalf("validate bson receipt: %v", err)
			}

			if got.Fee == nil {
				t.Fatal("expected bson fee receipt")
			}

			tc.assert(t, got.Fee)

			if got.GasUsed == nil || *got.GasUsed != gasUsed {
				t.Fatalf("unexpected bson gas used: %v", got.GasUsed)
			}
		})
	}
}

func TestFixedFeeReceiptValidationRejectsMismatchedAmounts(t *testing.T) {
	receipt := types.NewFixedFeeReceipt(types.CurrencyID("MCC"), common.NewBig(10))
	receipt.Amount = "11"

	if err := receipt.IsValid(nil); err == nil {
		t.Fatal("expected fixed fee receipt validation error")
	}
}

func TestFixedItemDataSizeExecutionFeeReceiptValidationRejectsInconsistentBreakdown(t *testing.T) {
	t.Run("item-fee", func(t *testing.T) {
		receipt := types.NewFixedItemDataSizeExecutionFeeReceipt(
			types.CurrencyID("MCC"),
			common.NewBig(18),
			common.NewBig(10),
			2,
			common.NewBig(2),
			common.NewBig(5),
			30,
			10,
			common.NewBig(1),
			common.NewBig(3),
			common.NewBig(1),
			common.NewBig(1),
		)

		if err := receipt.IsValid(nil); err == nil {
			t.Fatal("expected item fee validation error")
		}
	})

	t.Run("total-amount", func(t *testing.T) {
		receipt := types.NewFixedItemDataSizeExecutionFeeReceipt(
			types.CurrencyID("MCC"),
			common.NewBig(17),
			common.NewBig(10),
			2,
			common.NewBig(2),
			common.NewBig(4),
			30,
			10,
			common.NewBig(1),
			common.NewBig(3),
			common.NewBig(1),
			common.NewBig(1),
		)

		if err := receipt.IsValid(nil); err == nil {
			t.Fatal("expected total amount validation error")
		}
	})
}
