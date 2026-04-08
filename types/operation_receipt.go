package types

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
)

var CurrencyOperationReceiptHint = hint.MustNewHint("currency-operation-receipt-v0.0.1")

type FeeReceipt struct {
	CurrencyID CurrencyID `json:"currency_id" bson:"currency_id"`
	Amount     string     `json:"amount" bson:"amount"`
}

func (r FeeReceipt) IsValid() error {
	if err := r.CurrencyID.IsValid(nil); err != nil {
		return err
	}

	amount, err := common.NewBigFromString(r.Amount)
	if err != nil {
		return err
	}

	if !amount.OverNil() {
		return util.ErrInvalid.Errorf("fee amount under zero")
	}

	return nil
}

type CurrencyOperationReceipt struct {
	hint.BaseHinter
	Fee     *FeeReceipt `json:"fee,omitempty" bson:"fee,omitempty"`
	GasUsed *uint64     `json:"gas_used,omitempty" bson:"gas_used,omitempty"`
}

func NewCurrencyOperationReceipt(fee *FeeReceipt, gasUsed *uint64) CurrencyOperationReceipt {
	return CurrencyOperationReceipt{
		BaseHinter: hint.NewBaseHinter(CurrencyOperationReceiptHint),
		Fee:        fee,
		GasUsed:    gasUsed,
	}
}

func (r CurrencyOperationReceipt) IsValid([]byte) error {
	if err := r.BaseHinter.IsValid(CurrencyOperationReceiptHint.Type().Bytes()); err != nil {
		return err
	}

	if r.Fee != nil {
		if err := r.Fee.IsValid(); err != nil {
			return err
		}
	}

	return nil
}
