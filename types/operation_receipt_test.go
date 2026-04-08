package types_test

import (
	"testing"

	"github.com/imfact-labs/currency-model/app/runtime/steps"
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

func TestCurrencyOperationReceiptRoundTrip(t *testing.T) {
	encs, benc := newTestEncoders(t)
	gasUsed := uint64(33)
	receipt := types.NewCurrencyOperationReceipt(&types.FeeReceipt{
		CurrencyID: types.CurrencyID("MCC"),
		Amount:     "10",
	}, &gasUsed)

	t.Run("json", func(t *testing.T) {
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

		if got.Fee == nil || got.Fee.CurrencyID != types.CurrencyID("MCC") || got.Fee.Amount != "10" {
			t.Fatalf("unexpected json fee receipt: %+v", got.Fee)
		}

		if got.GasUsed == nil || *got.GasUsed != gasUsed {
			t.Fatalf("unexpected json gas used: %v", got.GasUsed)
		}
	})

	t.Run("bson", func(t *testing.T) {
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

		if got.Fee == nil || got.Fee.CurrencyID != types.CurrencyID("MCC") || got.Fee.Amount != "10" {
			t.Fatalf("unexpected bson fee receipt: %+v", got.Fee)
		}

		if got.GasUsed == nil || *got.GasUsed != gasUsed {
			t.Fatalf("unexpected bson gas used: %v", got.GasUsed)
		}
	})
}
