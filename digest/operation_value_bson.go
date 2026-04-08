package digest

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/operation/extras"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"time"

	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (va OperationValue) MarshalBSON() ([]byte, error) {
	var signs bson.A

	for i := range va.op.Signs() {
		signs = append(signs, bson.M{
			"signer":    va.op.Signs()[i].Signer().String(),
			"signature": va.op.Signs()[i].Signature(),
			"signed_at": va.op.Signs()[i].SignedAt(),
		})
	}

	var op = make(map[string]interface{})
	op = map[string]interface{}{
		"_hint": va.op.Hint().String(),
		"hash":  va.op.Hash().String(),
		"fact":  va.op.Fact(),
		"signs": signs,
	}

	var extension = make(map[string]interface{})
	eo, ok := va.op.(extras.OperationExtensions)
	if ok && len(eo.Extensions()) > 0 {
		for k, v := range eo.Extensions() {
			extension[k] = v
		}
		op["extension"] = extension
	}

	return bsonenc.Marshal(
		bson.M{
			"_hint":        va.Hint().String(),
			"op":           op,
			"height":       va.height,
			"confirmed_at": va.confirmedAt,
			"in_state":     va.inState,
			"reason":       va.reason,
			"index":        va.index,
			"receipt":      va.receipt,
		},
	)
}

type OperationValueBSONUnmarshaler struct {
	Hint        string        `bson:"_hint"`
	OP          bson.Raw      `bson:"op"`
	Height      base.Height   `bson:"height"`
	ConfirmedAt time.Time     `bson:"confirmed_at"`
	InState     bool          `bson:"in_state"`
	RS          string        `bson:"reason"`
	Index       uint64        `bson:"index"`
	Receipt     bson.RawValue `bson:"receipt"`
}

func (va *OperationValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode bson of OperationValue")

	var uva OperationValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &uva); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uva.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	va.BaseHinter = hint.NewBaseHinter(ht)

	var op common.BaseOperation
	if err := op.DecodeBSON(uva.OP, enc); err != nil {
		return e.Wrap(err)
	}

	var ueo extras.BaseOperationExtensions
	if err := ueo.DecodeBSON(uva.OP, enc); err != nil {
		return e.Wrap(err)
	}

	if ueo.Extensions() != nil {
		va.op = extras.ExtendedOperation{
			BaseOperation:           op,
			BaseOperationExtensions: &ueo,
		}
	} else {
		va.op = op
	}

	va.height = uva.Height
	va.confirmedAt = uva.ConfirmedAt
	va.inState = uva.InState
	va.index = uva.Index
	va.reason = uva.RS

	switch {
	case len(uva.Receipt.Value) < 1:
		va.receipt = nil
	case uva.Receipt.Type == bson.TypeNull:
		va.receipt = nil
	default:
		i, err := enc.Decode(uva.Receipt.Value)
		if err != nil {
			return e.Wrap(err)
		}

		if i == nil {
			va.receipt = nil
		} else if receipt, ok := i.(base.OperationReceipt); !ok {
			return e.Wrap(errors.Errorf("invalid operation receipt, %T", i))
		} else {
			va.receipt = receipt
		}
	}

	return nil
}
