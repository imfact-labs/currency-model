package types // nolint: dupl, revive

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (cs ContractAccountStatus) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":          cs.Hint().String(),
			"owner":          cs.owner,
			"is_active":      cs.isActive,
			"balance_status": cs.balanceStatus,
			"handlers":       cs.handlers,
			"recipients":     cs.recipients,
		},
	)
}

type ContractAccountBSONUnmarshaler struct {
	Hint          string   `bson:"_hint"`
	Owner         string   `bson:"owner"`
	IsActive      bool     `bson:"is_active"`
	BalanceStatus uint8    `bson:"balance_status"`
	Handlers      []string `bson:"handlers"`
	Recipients    []string `bson:"recipients"`
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

	return cs.unpack(enc, ht, ucs.Owner, ucs.IsActive, ucs.BalanceStatus, ucs.Handlers, ucs.Recipients)
}
