package did_registry

import (
	"github.com/imfact-labs/imfact-currency/common"
	"github.com/imfact-labs/imfact-currency/operation/extras"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/imfact-labs/imfact-currency/utils/bsonenc"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

func (fact RegisterModelFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":     fact.Hint().String(),
			"hash":      fact.BaseFact.Hash().String(),
			"token":     fact.BaseFact.Token(),
			"sender":    fact.sender,
			"contract":  fact.contract,
			"didMethod": fact.didMethod,
			"currency":  fact.currency,
		},
	)
}

type RegisterModelFactBSONUnmarshaler struct {
	Hint      string `bson:"_hint"`
	Sender    string `bson:"sender"`
	Contract  string `bson:"contract"`
	DIDMethod string `bson:"didMethod"`
	Currency  string `bson:"currency"`
}

func (fact *RegisterModelFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u common.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(u.Hash))
	fact.BaseFact.SetToken(u.Token)

	var uf RegisterModelFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	if err := fact.unpack(
		enc, uf.Sender, uf.Contract, uf.DIDMethod, uf.Currency,
	); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	return nil
}

func (op RegisterModel) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": op.Hint().String(),
			"hash":  op.Hash().String(),
			"fact":  op.Fact(),
			"signs": op.Signs(),
		},
	)
}

func (op *RegisterModel) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo common.BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *op)
	}

	op.BaseOperation = ubo

	var ueo extras.BaseOperationExtensions
	if err := ueo.DecodeBSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *op)
	}

	op.BaseOperationExtensions = &ueo

	return nil
}
