package extension // nolint: dupl

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/operation/extras"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util/hint"
)

func (fact WithdrawFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":  fact.Hint().String(),
			"sender": fact.sender,
			"items":  fact.items,
			"hash":   fact.Hash(),
			"token":  fact.Token(),
		},
	)
}

type WithdrawFactBSONUnmarshaler struct {
	Hint   string   `bson:"_hint"`
	Sender string   `bson:"sender"`
	Items  bson.Raw `bson:"items"`
}

func (fact *WithdrawFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {

	var ubf common.BaseFactBSONUnmarshaler
	err := enc.Unmarshal(b, &ubf)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	fact.SetHash(ubf.Hash)
	fact.SetToken(ubf.Token)

	var uf WithdrawFactBSONUnmarshaler
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

func (op Withdraw) MarshalBSON() ([]byte, error) {
	bm := bson.M{}
	for k, v := range op.Extensions() {
		bm[k] = v
	}
	return bsonenc.Marshal(
		bson.M{
			"_hint":     op.Hint().String(),
			"hash":      op.Hash(),
			"fact":      op.Fact(),
			"signs":     op.Signs(),
			"extension": bm,
		})
}

func (op *Withdraw) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {

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
