package currency

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type MintFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Receiver base.Address `json:"receiver"`
	Amount   types.Amount `json:"amount"`
}

func (fact MintFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(MintFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Receiver:              fact.receiver,
		Amount:                fact.amount,
	})
}

type MintFactJSONUnmarshaler struct {
	base.BaseFactJSONUnmarshaler
	Receiver string          `json:"receiver"`
	Amount   json.RawMessage `json:"amount"`
}

func (fact *MintFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uf MintFactJSONUnmarshaler

	if err := enc.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return nil
}

func (op *Mint) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var ubo common.BaseNodeOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *op)
	}

	op.BaseNodeOperation = ubo

	return nil
}
