package currency

import (
	"encoding/json"
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
)

type RegisterCurrencyFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Currency types.CurrencyDesign `json:"currency"`
}

func (fact RegisterCurrencyFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(RegisterCurrencyFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Currency:              fact.currency,
	})
}

type RegisterCurrencyFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Currency json.RawMessage `json:"currency"`
}

func (fact *RegisterCurrencyFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uf RegisterCurrencyFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	if err := fact.unpack(enc, uf.Currency); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	return nil
}

//func (op RegisterCurrency) MarshalJSON() ([]byte, error) {
//	return util.MarshalJSON(BaseOperationMarshaler{
//		MBaseOperationJSONMarshaler: op.MBaseOperation.JSONMarshaler(),
//	})
//}

func (op *RegisterCurrency) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var ubo common.BaseNodeOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *op)
	}

	op.BaseNodeOperation = ubo

	return nil
}
