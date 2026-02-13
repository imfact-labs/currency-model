package extension // nolint: dupl

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/ProtoconNet/mitum-currency/v3/utils/bsonenc"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

func (fact UpdateRecipientFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":      fact.Hint().String(),
			"sender":     fact.sender,
			"contract":   fact.contract,
			"recipients": fact.recipients,
			"currency":   fact.currency,
			"hash":       fact.BaseFact.Hash().String(),
			"token":      fact.BaseFact.Token(),
		},
	)
}

type UpdateRecipientsFactBSONUnmarshaler struct {
	Hint       string   `bson:"_hint"`
	Sender     string   `bson:"sender"`
	Contract   string   `bson:"contract"`
	Recipients []string `bson:"recipients"`
	Currency   string   `bson:"currency"`
}

func (fact *UpdateRecipientFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
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

	var uf UpdateRecipientsFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	if err := fact.unpack(enc, uf.Sender, uf.Contract, uf.Recipients, uf.Currency); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	return nil
}

func (op UpdateRecipient) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": op.Hint().String(),
			"hash":  op.Hash(),
			"fact":  op.Fact(),
			"signs": op.Signs(),
		})
}

func (op *UpdateRecipient) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
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
