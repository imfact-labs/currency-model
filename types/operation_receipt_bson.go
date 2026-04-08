package types

import (
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (r CurrencyOperationReceipt) MarshalBSON() ([]byte, error) {
	m := bson.M{
		"_hint": r.Hint().String(),
	}

	if r.Fee != nil {
		m["fee"] = r.Fee
	}

	if r.GasUsed != nil {
		m["gas_used"] = r.GasUsed
	}

	return bsonenc.Marshal(m)
}

type CurrencyOperationReceiptBSONUnmarshaler struct {
	Hint    string      `bson:"_hint"`
	Fee     *FeeReceipt `bson:"fee"`
	GasUsed *uint64     `bson:"gas_used"`
}

func (r *CurrencyOperationReceipt) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u CurrencyOperationReceiptBSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	hts := u.Hint
	if hts == "" {
		hts = CurrencyOperationReceiptHint.String()
	}

	ht, err := hint.ParseHint(hts)
	if err != nil {
		return err
	}

	r.BaseHinter = hint.NewBaseHinter(ht)
	r.Fee = u.Fee
	r.GasUsed = u.GasUsed

	return nil
}
