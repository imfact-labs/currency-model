package extension

import (
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (c ContractAccountStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":           c.Hint().String(),
			"contractaccount": c.status,
		},
	)

}

type ContractAccountStateValueBSONUnmarshaler struct {
	Hint            string   `bson:"_hint"`
	ContractAccount bson.Raw `bson:"contractaccount"`
}

func (c *ContractAccountStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode bson of ContractAccountStateValue")

	var u ContractAccountStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	c.BaseHinter = hint.NewBaseHinter(ht)

	var ca types.ContractAccountStatus
	if err := ca.DecodeBSON(u.ContractAccount, enc); err != nil {
		return e.Wrap(err)
	}

	c.status = ca

	return nil
}
