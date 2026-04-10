package types

import (
	"encoding/json"

	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

type BaseFeeReceiptJSONMarshaler struct {
	hint.BaseHinter
	CurrencyID CurrencyID `json:"currency_id"`
	Amount     string     `json:"amount"`
}

func (r BaseFeeReceipt) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(BaseFeeReceiptJSONMarshaler{
		BaseHinter: r.BaseHinter,
		CurrencyID: r.CurrencyID,
		Amount:     r.Amount,
	})
}

type BaseFeeReceiptJSONUnmarshaler struct {
	Hint       hint.Hint  `json:"_hint"`
	CurrencyID CurrencyID `json:"currency_id"`
	Amount     string     `json:"amount"`
}

func (r *BaseFeeReceipt) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u BaseFeeReceiptJSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	ht := u.Hint
	if ht.String() == "" {
		ht = BaseFeeReceiptHint
	}

	r.BaseHinter = hint.NewBaseHinter(ht)
	r.CurrencyID = u.CurrencyID
	r.Amount = u.Amount

	return nil
}

type FixedFeeReceiptJSONMarshaler struct {
	hint.BaseHinter
	CurrencyID CurrencyID `json:"currency_id"`
	Amount     string     `json:"amount"`
	BaseAmount string     `json:"base_amount"`
}

func (r FixedFeeReceipt) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(FixedFeeReceiptJSONMarshaler{
		BaseHinter: r.BaseHinter,
		CurrencyID: r.CurrencyID,
		Amount:     r.Amount,
		BaseAmount: r.BaseAmount,
	})
}

type FixedFeeReceiptJSONUnmarshaler struct {
	Hint       hint.Hint  `json:"_hint"`
	CurrencyID CurrencyID `json:"currency_id"`
	Amount     string     `json:"amount"`
	BaseAmount string     `json:"base_amount"`
}

func (r *FixedFeeReceipt) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u FixedFeeReceiptJSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	ht := u.Hint
	if ht.String() == "" {
		ht = FixedFeeReceiptHint
	}

	r.BaseHinter = hint.NewBaseHinter(ht)
	r.CurrencyID = u.CurrencyID
	r.Amount = u.Amount
	r.BaseAmount = u.BaseAmount

	return nil
}

type FixedItemDataSizeExecutionFeeReceiptJSONMarshaler struct {
	hint.BaseHinter
	CurrencyID         CurrencyID `json:"currency_id"`
	TotalAmount        string     `json:"total_amount"`
	BaseAmount         string     `json:"base_amount"`
	ItemCount          int        `json:"item_count"`
	ItemFeeAmount      string     `json:"item_fee_amount"`
	ItemFee            string     `json:"item_fee"`
	DataSize           int        `json:"data_size"`
	DataSizeUnit       int64      `json:"data_size_unit"`
	DataSizeFeeAmount  string     `json:"data_size_fee_amount"`
	DataSizeFee        string     `json:"data_size_fee"`
	ExecutionFeeAmount string     `json:"execution_fee_amount"`
	ExecutionFee       string     `json:"execution_fee"`
}

func (r FixedItemDataSizeExecutionFeeReceipt) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(FixedItemDataSizeExecutionFeeReceiptJSONMarshaler{
		BaseHinter:         r.BaseHinter,
		CurrencyID:         r.CurrencyID,
		TotalAmount:        r.TotalAmount,
		BaseAmount:         r.BaseAmount,
		ItemCount:          r.ItemCount,
		ItemFeeAmount:      r.ItemFeeAmount,
		ItemFee:            r.ItemFee,
		DataSize:           r.DataSize,
		DataSizeUnit:       r.DataSizeUnit,
		DataSizeFeeAmount:  r.DataSizeFeeAmount,
		DataSizeFee:        r.DataSizeFee,
		ExecutionFeeAmount: r.ExecutionFeeAmount,
		ExecutionFee:       r.ExecutionFee,
	})
}

type FixedItemDataSizeExecutionFeeReceiptJSONUnmarshaler struct {
	Hint               hint.Hint  `json:"_hint"`
	CurrencyID         CurrencyID `json:"currency_id"`
	TotalAmount        string     `json:"total_amount"`
	BaseAmount         string     `json:"base_amount"`
	ItemCount          int        `json:"item_count"`
	ItemFeeAmount      string     `json:"item_fee_amount"`
	ItemFee            string     `json:"item_fee"`
	DataSize           int        `json:"data_size"`
	DataSizeUnit       int64      `json:"data_size_unit"`
	DataSizeFeeAmount  string     `json:"data_size_fee_amount"`
	DataSizeFee        string     `json:"data_size_fee"`
	ExecutionFeeAmount string     `json:"execution_fee_amount"`
	ExecutionFee       string     `json:"execution_fee"`
}

func (r *FixedItemDataSizeExecutionFeeReceipt) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u FixedItemDataSizeExecutionFeeReceiptJSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	ht := u.Hint
	if ht.String() == "" {
		ht = FixedItemDataSizeExecutionFeeReceiptHint
	}

	r.BaseHinter = hint.NewBaseHinter(ht)
	r.CurrencyID = u.CurrencyID
	r.TotalAmount = u.TotalAmount
	r.BaseAmount = u.BaseAmount
	r.ItemCount = u.ItemCount
	r.ItemFeeAmount = u.ItemFeeAmount
	r.ItemFee = u.ItemFee
	r.DataSize = u.DataSize
	r.DataSizeUnit = u.DataSizeUnit
	r.DataSizeFeeAmount = u.DataSizeFeeAmount
	r.DataSizeFee = u.DataSizeFee
	r.ExecutionFeeAmount = u.ExecutionFeeAmount
	r.ExecutionFee = u.ExecutionFee

	return nil
}

type CurrencyOperationReceiptJSONMarshaler struct {
	hint.BaseHinter
	Fee     FeeReceipt `json:"fee,omitempty"`
	GasUsed *uint64    `json:"gas_used,omitempty"`
}

func (r CurrencyOperationReceipt) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CurrencyOperationReceiptJSONMarshaler{
		BaseHinter: r.BaseHinter,
		Fee:        r.Fee,
		GasUsed:    r.GasUsed,
	})
}

type CurrencyOperationReceiptJSONUnmarshaler struct {
	Hint    hint.Hint       `json:"_hint"`
	Fee     json.RawMessage `json:"fee"`
	GasUsed *uint64         `json:"gas_used"`
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
	r.GasUsed = u.GasUsed

	switch string(u.Fee) {
	case "", "null":
		r.Fee = nil
	default:
		var fee FeeReceipt
		if err := encoder.Decode(enc, u.Fee, &fee); err != nil {
			return err
		}

		r.Fee = fee
	}

	return nil
}
