package types

import (
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
)

func (fa NilFeeer) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.NewHintedDoc(fa.Hint()))
}

func (fa *NilFeeer) UnmarsahlBSON(b []byte) error {
	e := util.StringError("unmarshal bson of NilFeeer")

	var head bsonenc.HintedHead
	if err := bsonenc.Unmarshal(b, &head); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(head.H)
	if err != nil {
		return e.Wrap(err)
	}

	fa.BaseHinter = hint.NewBaseHinter(ht)

	return nil
}

func (fa FixedFeeer) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    fa.Hint().String(),
			"receiver": fa.receiver,
			"amount":   fa.amount.String(),
		},
	)

}

type FixedFeeerBSONUnmarshaler struct {
	Hint     string `bson:"_hint"`
	Receiver string `bson:"receiver"`
	Amount   string `bson:"amount"`
}

func (fa *FixedFeeer) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode bson of FixedFeeer")

	var ufa FixedFeeerBSONUnmarshaler
	if err := enc.Unmarshal(b, &ufa); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(ufa.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return fa.unpack(enc, ht, ufa.Receiver, ufa.Amount)
}

func (fa FixedItemDataSizeExecutionFeeer) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":                fa.Hint().String(),
			"receiver":             fa.receiver,
			"amount":               fa.amount.String(),
			"item_fee_amount":      fa.itemFeeAmount.String(),
			"data_size_fee_amount": fa.dataSizeFeeAmount.String(),
			"data_size_unit":       fa.dataSizeUnit,
			"execution_fee_amount": fa.executionFeeAmount.String(),
		},
	)
}

type FixedItemDataSizeExecutionFeeerBSONUnmarshaler struct {
	Hint               string `bson:"_hint"`
	Receiver           string `bson:"receiver"`
	Amount             string `bson:"amount"`
	ItemFeeAmount      string `bson:"item_fee_amount"`
	DataSizeFeeAmount  string `bson:"data_size_fee_amount"`
	DataSizeUnit       int64  `bson:"data_size_unit"`
	ExecutionFeeAmount string `bson:"execution_fee_amount"`
}

func (fa *FixedItemDataSizeExecutionFeeer) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode bson of FixedItemDataSizeFeeer")

	var ufa FixedItemDataSizeExecutionFeeerBSONUnmarshaler
	if err := enc.Unmarshal(b, &ufa); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(ufa.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return fa.unpack(
		enc, ht, ufa.Receiver, ufa.Amount, ufa.ItemFeeAmount,
		ufa.DataSizeFeeAmount, ufa.DataSizeUnit, ufa.ExecutionFeeAmount,
	)
}
