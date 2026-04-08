package digest_test

import (
	"testing"
	"time"

	"github.com/imfact-labs/currency-model/app/runtime/steps"
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/digest"
	"github.com/imfact-labs/currency-model/operation/currency"
	operationtest "github.com/imfact-labs/currency-model/operation/test"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/base"
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

func TestOperationValueBSONRoundTripWithReceipt(t *testing.T) {
	_, benc := newTestEncoders(t)

	getter := operationtest.NewMockStateGetter()
	var tp operationtest.TestProcessor
	tp.Setup(getter)

	op, err := currency.NewMint(currency.NewMintFact(
		[]byte("mint"),
		tp.GenesisAddr,
		types.NewAmount(common.NewBig(5), tp.GenesisCurrency),
	))
	if err != nil {
		t.Fatalf("new mint operation: %v", err)
	}

	if err := op.NodeSign(tp.NodePriv, tp.NetworkID, tp.NodeAddr); err != nil {
		t.Fatalf("sign mint operation: %v", err)
	}

	receipt := types.NewCurrencyOperationReceipt(&types.FeeReceipt{
		CurrencyID: tp.GenesisCurrency,
		Amount:     "10",
	}, nil)

	value := digest.NewOperationValue(
		op,
		base.Height(2),
		time.Unix(123, 0).UTC(),
		true,
		"",
		7,
		receipt,
	)

	b, err := benc.Marshal(value)
	if err != nil {
		t.Fatalf("marshal operation value: %v", err)
	}

	i, err := benc.Decode(b)
	if err != nil {
		t.Fatalf("decode operation value: %v", err)
	}

	got, ok := i.(digest.OperationValue)
	if !ok {
		t.Fatalf("decoded operation value type = %T", i)
	}

	if got.Index() != 7 {
		t.Fatalf("unexpected index: %d", got.Index())
	}

	if !got.Operation().Fact().Hash().Equal(op.Fact().Hash()) {
		t.Fatalf("fact hash mismatch")
	}

	decodedReceipt, ok := got.Receipt().(types.CurrencyOperationReceipt)
	if !ok {
		t.Fatalf("decoded receipt type = %T", got.Receipt())
	}

	if decodedReceipt.Fee == nil || decodedReceipt.Fee.CurrencyID != tp.GenesisCurrency || decodedReceipt.Fee.Amount != "10" {
		t.Fatalf("unexpected decoded fee receipt: %+v", decodedReceipt.Fee)
	}
}
