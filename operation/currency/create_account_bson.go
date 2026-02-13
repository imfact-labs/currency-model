package currency // nolint: dupl

import (
	"github.com/imfact-labs/imfact-currency/common"
	"github.com/imfact-labs/imfact-currency/operation/extras"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/imfact-labs/imfact-currency/utils/bsonenc"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

func (fact CreateAccountFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":  fact.Hint().String(),
			"sender": fact.sender,
			"items":  fact.items,
			"hash":   fact.BaseFact.Hash().String(),
			"token":  fact.BaseFact.Token(),
		},
	)
}

type CreateAccountFactBSONUnmarshaler struct {
	Hint   string   `bson:"_hint"`
	Sender string   `bson:"sender"`
	Items  bson.Raw `bson:"items"`
}

func (fact *CreateAccountFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u common.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	h := valuehash.NewBytesFromString(u.Hash)

	fact.BaseFact.SetHash(h)
	err = fact.BaseFact.SetToken(u.Token)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	var uf CreateAccountFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	if err := fact.unpack(enc, uf.Sender, uf.Items); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	return nil
}

func (op CreateAccount) MarshalBSON() ([]byte, error) {
	bm := bson.M{}
	for k, v := range op.Extensions() {
		bm[k] = v
	}

	return bsonenc.Marshal(
		bson.M{
			"_hint":     op.Hint().String(),
			"hash":      op.Hash().String(),
			"fact":      op.Fact(),
			"signs":     op.Signs(),
			"extension": bm,
		})
}

func (op *CreateAccount) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
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
