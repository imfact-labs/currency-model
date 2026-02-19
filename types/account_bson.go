package types

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/valuehash"
)

func (ac Account) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":   ac.Hint().String(),
			"hash":    ac.h,
			"address": ac.address,
			"keys":    ac.keys,
		},
	)
}

type AccountBSONUnmarshaler struct {
	Hint    string          `bson:"_hint"`
	Hash    valuehash.Bytes `bson:"hash"`
	Address string          `bson:"address"`
	Keys    bson.Raw        `bson:"keys"`
}

func (ac *Account) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode bson of Account")

	var uac AccountBSONUnmarshaler
	if err := enc.Unmarshal(b, &uac); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uac.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	ac.h = valuehash.NewHashFromBytes(uac.Hash)

	ac.BaseHinter = hint.NewBaseHinter(ht)
	switch ad, err := base.DecodeAddress(uac.Address, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		ac.address = ad
	}

	k, err := enc.Decode(uac.Keys)
	if err != nil {
		return e.Wrap(err)
	} else if k != nil {
		v, ok := k.(AccountKeys)
		if !ok {
			return errors.Errorf("expected BaseAccountKeys, not %T", k)
		}
		ac.keys = v
	}

	return nil
}
