package types

import (
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (d Data) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bson.M{
		"_hint":   d.Hint().String(),
		"address": d.address,
		"did":     d.did.String(),
	})
}

type DataBSONUnmarshaler struct {
	Hint    string `bson:"_hint"`
	Address string `bson:"address"`
	DID     string `bson:"did"`
}

func (d *Data) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of Data")

	var u DataBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	did, err := NewDIDRefFromString(u.DID)
	if err != nil {
		return e.Wrap(err)
	}

	return d.unpack(enc, ht, u.Address, *did)
}
