package types

import (
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

type CurrencyOperationReceiptJSONMarshaler struct {
	hint.BaseHinter
	Fee     *FeeReceipt `json:"fee,omitempty"`
	GasUsed *uint64     `json:"gas_used,omitempty"`
}

func (r CurrencyOperationReceipt) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CurrencyOperationReceiptJSONMarshaler{
		BaseHinter: r.BaseHinter,
		Fee:        r.Fee,
		GasUsed:    r.GasUsed,
	})
}

type CurrencyOperationReceiptJSONUnmarshaler struct {
	Hint    hint.Hint   `json:"_hint"`
	Fee     *FeeReceipt `json:"fee"`
	GasUsed *uint64     `json:"gas_used"`
}

func (r *CurrencyOperationReceipt) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u CurrencyOperationReceiptJSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	ht := u.Hint
	if ht.String() == "" {
		ht = CurrencyOperationReceiptHint
	}

	r.BaseHinter = hint.NewBaseHinter(ht)
	r.Fee = u.Fee
	r.GasUsed = u.GasUsed

	return nil
}
