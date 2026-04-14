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
	TotalFee   string     `json:"total_fee"`
}

func (r BaseFeeReceipt) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(BaseFeeReceiptJSONMarshaler{
		BaseHinter: r.BaseHinter,
		CurrencyID: r.currencyID,
		TotalFee:   r.totalFee,
	})
}

type BaseFeeReceiptJSONUnmarshaler struct {
	Hint       hint.Hint  `json:"_hint"`
	CurrencyID CurrencyID `json:"currency_id"`
	TotalFee   string     `json:"total_fee"`
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
	r.currencyID = u.CurrencyID
	r.totalFee = u.TotalFee

	return nil
}

type FixedFeeReceiptJSONMarshaler struct {
	hint.BaseHinter
	CurrencyID CurrencyID `json:"currency_id"`
	TotalFee   string     `json:"total_fee"`
	BaseFee    string     `json:"base_fee"`
}

func (r FixedFeeReceipt) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(FixedFeeReceiptJSONMarshaler{
		BaseHinter: r.BaseHinter,
		CurrencyID: r.currencyID,
		TotalFee:   r.totalFee,
		BaseFee:    r.baseFee,
	})
}

type FixedFeeReceiptJSONUnmarshaler struct {
	Hint       hint.Hint  `json:"_hint"`
	CurrencyID CurrencyID `json:"currency_id"`
	TotalFee   string     `json:"total_fee"`
	BaseFee    string     `json:"base_fee"`
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
	r.currencyID = u.CurrencyID
	r.totalFee = u.TotalFee
	r.baseFee = u.BaseFee

	return nil
}

type FixedItemDataSizeExecutionFeeReceiptJSONMarshaler struct {
	hint.BaseHinter
	CurrencyID       CurrencyID `json:"currency_id"`
	TotalFee         string     `json:"total_fee"`
	BaseFee          string     `json:"base_fee"`
	ItemUnitFee      string     `json:"item_unit_fee"`
	ItemCount        int        `json:"item_count"`
	ItemFee          string     `json:"item_fee"`
	DataSizeUnitFee  string     `json:"data_size_unit_fee"`
	DataSizeUnit     int64      `json:"data_size_unit"`
	DataSize         int        `json:"data_size"`
	DataSizeFee      string     `json:"data_size_fee"`
	ExecutionCount   int        `json:"execution_count,omitempty"`
	ExecutionUnitFee string     `json:"execution_unit_fee,omitempty"`
	ExecutionFee     string     `json:"execution_fee,omitempty"`
}

func (r FixedItemDataSizeExecutionFeeReceipt) MarshalJSON() ([]byte, error) {
	var executionUnitFee string
	var executionFee string
	if r.executionUnitFee == "0" {
		executionUnitFee = ""
	}
	if r.executionFee == "0" {
		executionFee = ""
	}
	return util.MarshalJSON(FixedItemDataSizeExecutionFeeReceiptJSONMarshaler{
		BaseHinter:       r.BaseHinter,
		CurrencyID:       r.currencyID,
		TotalFee:         r.totalFee,
		BaseFee:          r.baseFee,
		ItemUnitFee:      r.itemUnitFee,
		ItemCount:        r.itemCount,
		ItemFee:          r.itemFee,
		DataSizeUnitFee:  r.dataSizeUnitFee,
		DataSizeUnit:     r.dataSizeUnit,
		DataSize:         r.dataSize,
		DataSizeFee:      r.dataSizeFee,
		ExecutionCount:   r.executionCount,
		ExecutionUnitFee: executionUnitFee,
		ExecutionFee:     executionFee,
	})
}

type FixedItemDataSizeExecutionFeeReceiptJSONUnmarshaler struct {
	Hint             hint.Hint  `json:"_hint"`
	CurrencyID       CurrencyID `json:"currency_id"`
	TotalFee         string     `json:"total_fee"`
	BaseFee          string     `json:"base_fee"`
	ItemUnitFee      string     `json:"item_unit_fee"`
	ItemCount        int        `json:"item_count"`
	ItemFee          string     `json:"item_fee"`
	DataSizeUnitFee  string     `json:"data_size_unit_fee"`
	DataSizeUnit     int64      `json:"data_size_unit"`
	DataSize         int        `json:"data_size"`
	DataSizeFee      string     `json:"data_size_fee"`
	ExecutionCount   int        `json:"execution_count"`
	ExecutionUnitFee string     `json:"execution_unit_fee"`
	ExecutionFee     string     `json:"execution_fee"`
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

	var executionUnitFee string
	var executionFee string
	if u.ExecutionUnitFee == "" {
		executionUnitFee = "0"
	}
	if u.ExecutionFee == "" {
		executionFee = "0"
	}

	r.BaseHinter = hint.NewBaseHinter(ht)
	r.currencyID = u.CurrencyID
	r.totalFee = u.TotalFee
	r.baseFee = u.BaseFee
	r.itemUnitFee = u.ItemUnitFee
	r.itemCount = u.ItemCount
	r.itemFee = u.ItemFee
	r.dataSizeUnitFee = u.DataSizeUnitFee
	r.dataSizeUnit = u.DataSizeUnit
	r.dataSize = u.DataSize
	r.dataSizeFee = u.DataSizeFee
	r.executionCount = u.ExecutionCount
	r.executionUnitFee = executionUnitFee
	r.executionFee = executionFee

	return nil
}

type CurrencyOperationReceiptJSONMarshaler struct {
	hint.BaseHinter
	Feeer   string     `json:"feeer,omitempty"`
	Fee     FeeReceipt `json:"fee,omitempty"`
	GasUsed *uint64    `json:"gas_used,omitempty"`
}

func (r CurrencyOperationReceipt) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CurrencyOperationReceiptJSONMarshaler{
		BaseHinter: r.BaseHinter,
		Feeer:      r.feeer,
		Fee:        r.Fee,
		GasUsed:    r.GasUsed,
	})
}

type CurrencyOperationReceiptJSONUnmarshaler struct {
	Hint    hint.Hint       `json:"_hint"`
	Feeer   string          `json:"feeer"`
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
	r.feeer = u.Feeer
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
