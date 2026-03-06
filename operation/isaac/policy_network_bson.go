package isaacoperation

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (fact GenesisNetworkPolicyFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":  fact.Hint().String(),
			"policy": fact.policy,
			"hash":   fact.Hash(),
			"token":  fact.Token(),
		},
	)
}

type GenesisNetworkPolicyFactBSONUnMarshaler struct {
	Hint   string   `bson:"_hint"`
	Policy bson.Raw `bson:"policy"`
}

func (fact *GenesisNetworkPolicyFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode bson of GenesisNetworkPolicyFact")

	var u common.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e.Wrap(err)
	}

	fact.SetHash(u.Hash)
	fact.SetToken(u.Token)

	var uf GenesisNetworkPolicyFactBSONUnMarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	if err := encoder.Decode(enc, uf.Policy, &fact.policy); err != nil {
		return e.Wrap(err)
	}

	return nil
}

func (fact NetworkPolicyFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":  fact.Hint().String(),
			"policy": fact.policy,
			"hash":   fact.Hash(),
			"token":  fact.Token(),
		},
	)
}

type NetworkPolicyFactBSONUnMarshaler struct {
	Hint   string   `bson:"_hint"`
	Policy bson.Raw `bson:"policy"`
}

func (fact *NetworkPolicyFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode bson of NetworkPolicyFact")

	var u common.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetHash(u.Hash)
	fact.BaseFact.SetToken(u.Token)

	var uf NetworkPolicyFactBSONUnMarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	if err := encoder.Decode(enc, uf.Policy, &fact.policy); err != nil {
		return e.Wrap(err)
	}

	return nil
}
