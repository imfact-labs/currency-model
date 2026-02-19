package types // nolint: dupl, revive

import (
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (cs ContractAccountStatus) MarshalBSON() ([]byte, error) {
	var rs string
	if cs.registerOperation != nil {
		rs = cs.registerOperation.String()
	}
	return bsonenc.Marshal(
		bson.M{
			"_hint":              cs.Hint().String(),
			"owner":              cs.owner,
			"is_active":          cs.isActive,
			"balance_status":     cs.balanceStatus,
			"register_operation": rs,
			"handlers":           cs.handlers,
			"recipients":         cs.recipients,
		},
	)
}

type ContractAccountBSONUnmarshaler struct {
	Hint              string   `bson:"_hint"`
	Owner             string   `bson:"owner"`
	IsActive          bool     `bson:"is_active"`
	BalanceStatus     uint8    `bson:"balance_status"`
	RegisterOperation string   `bson:"register_operation"`
	Handlers          []string `bson:"handlers"`
	Recipients        []string `bson:"recipients"`
}

func (cs *ContractAccountStatus) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode bson of ContractAccountStatus")

	var ucs ContractAccountBSONUnmarshaler
	if err := bsonenc.Unmarshal(b, &ucs); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(ucs.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	var rht *hint.Hint
	if ucs.RegisterOperation != "" {
		h, err := hint.ParseHint(ucs.RegisterOperation)
		if err != nil {
			return e.Wrap(err)
		}
		rht = &h
	}

	return cs.unpack(enc, ht, ucs.Owner, ucs.IsActive, ucs.BalanceStatus, rht, ucs.Handlers, ucs.Recipients)
}
