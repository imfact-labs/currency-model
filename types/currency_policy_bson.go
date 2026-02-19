package types

import (
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
)

func (po CurrencyPolicy) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":       po.Hint().String(),
			"min_balance": po.minBalance.String(),
			"feeer":       po.feeer,
		},
	)
}

type CurrencyPolicyBSONUnmarshaler struct {
	Hint       string   `bson:"_hint"`
	MinBalance string   `bson:"min_balance"`
	Feeer      bson.Raw `bson:"feeer"`
}

func (po *CurrencyPolicy) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode bson of CurrencyPolicy")

	var upo CurrencyPolicyBSONUnmarshaler
	if err := enc.Unmarshal(b, &upo); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(upo.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return po.unpack(enc, ht, upo.MinBalance, upo.Feeer)
}
