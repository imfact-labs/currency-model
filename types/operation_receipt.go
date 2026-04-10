package types

import (
	"fmt"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
)

var CurrencyOperationReceiptHint = hint.MustNewHint("currency-operation-receipt-v0.0.1")

var BaseFeeReceiptHint = hint.MustNewHint("currency-base-fee-receipt-v0.0.1")

var FixedFeeReceiptHint = hint.MustNewHint("currency-fixed-fee-receipt-v0.0.1")

var FixedItemDataSizeExecutionFeeReceiptHint = hint.MustNewHint(
	"currency-fixed-item-data-size-execution-fee-receipt-v0.0.1",
)

type FeeReceipt interface {
	hint.Hinter
	util.IsValider
	Currency() CurrencyID
	FeeAmount() string
}

type BaseFeeReceipt struct {
	hint.BaseHinter
	CurrencyID CurrencyID `json:"currency_id" bson:"currency_id"`
	Amount     string     `json:"amount" bson:"amount"`
}

func NewBaseFeeReceipt(currencyID CurrencyID, amount common.Big) BaseFeeReceipt {
	return BaseFeeReceipt{
		BaseHinter: hint.NewBaseHinter(BaseFeeReceiptHint),
		CurrencyID: currencyID,
		Amount:     amount.String(),
	}
}

func (r BaseFeeReceipt) Currency() CurrencyID {
	return r.CurrencyID
}

func (r BaseFeeReceipt) FeeAmount() string {
	return r.Amount
}

func (r BaseFeeReceipt) IsValid([]byte) error {
	if err := r.BaseHinter.IsValid(BaseFeeReceiptHint.Type().Bytes()); err != nil {
		return err
	}

	return validateReceiptAmounts(r.CurrencyID, r.Amount)
}

type FixedFeeReceipt struct {
	hint.BaseHinter
	CurrencyID CurrencyID `json:"currency_id" bson:"currency_id"`
	Amount     string     `json:"amount" bson:"amount"`
	BaseAmount string     `json:"base_amount" bson:"base_amount"`
}

func NewFixedFeeReceipt(currencyID CurrencyID, amount common.Big) FixedFeeReceipt {
	return FixedFeeReceipt{
		BaseHinter: hint.NewBaseHinter(FixedFeeReceiptHint),
		CurrencyID: currencyID,
		Amount:     amount.String(),
		BaseAmount: amount.String(),
	}
}

func (r FixedFeeReceipt) Currency() CurrencyID {
	return r.CurrencyID
}

func (r FixedFeeReceipt) FeeAmount() string {
	return r.Amount
}

func (r FixedFeeReceipt) IsValid([]byte) error {
	if err := r.BaseHinter.IsValid(FixedFeeReceiptHint.Type().Bytes()); err != nil {
		return err
	}

	amounts, err := receiptAmountsByField(
		r.CurrencyID,
		map[string]string{
			"amount":      r.Amount,
			"base_amount": r.BaseAmount,
		},
	)
	if err != nil {
		return err
	}

	if !amounts["amount"].Equal(amounts["base_amount"]) {
		return util.ErrInvalid.Errorf("amount and base_amount do not match")
	}

	return nil
}

type FixedItemDataSizeExecutionFeeReceipt struct {
	hint.BaseHinter
	CurrencyID         CurrencyID `json:"currency_id" bson:"currency_id"`
	TotalAmount        string     `json:"total_amount" bson:"total_amount"`
	BaseAmount         string     `json:"base_amount" bson:"base_amount"`
	ItemCount          int        `json:"item_count" bson:"item_count"`
	ItemFeeAmount      string     `json:"item_fee_amount" bson:"item_fee_amount"`
	ItemFee            string     `json:"item_fee" bson:"item_fee"`
	DataSize           int        `json:"data_size" bson:"data_size"`
	DataSizeUnit       int64      `json:"data_size_unit" bson:"data_size_unit"`
	DataSizeFeeAmount  string     `json:"data_size_fee_amount" bson:"data_size_fee_amount"`
	DataSizeFee        string     `json:"data_size_fee" bson:"data_size_fee"`
	ExecutionFeeAmount string     `json:"execution_fee_amount" bson:"execution_fee_amount"`
	ExecutionFee       string     `json:"execution_fee" bson:"execution_fee"`
}

func NewFixedItemDataSizeExecutionFeeReceipt(
	currencyID CurrencyID,
	totalAmount common.Big,
	baseAmount common.Big,
	itemCount int,
	itemFeeAmount common.Big,
	itemFee common.Big,
	dataSize int,
	dataSizeUnit int64,
	dataSizeFeeAmount common.Big,
	dataSizeFee common.Big,
	executionFeeAmount common.Big,
	executionFee common.Big,
) FixedItemDataSizeExecutionFeeReceipt {
	return FixedItemDataSizeExecutionFeeReceipt{
		BaseHinter:         hint.NewBaseHinter(FixedItemDataSizeExecutionFeeReceiptHint),
		CurrencyID:         currencyID,
		TotalAmount:        totalAmount.String(),
		BaseAmount:         baseAmount.String(),
		ItemCount:          itemCount,
		ItemFeeAmount:      itemFeeAmount.String(),
		ItemFee:            itemFee.String(),
		DataSize:           dataSize,
		DataSizeUnit:       dataSizeUnit,
		DataSizeFeeAmount:  dataSizeFeeAmount.String(),
		DataSizeFee:        dataSizeFee.String(),
		ExecutionFeeAmount: executionFeeAmount.String(),
		ExecutionFee:       executionFee.String(),
	}
}

func (r FixedItemDataSizeExecutionFeeReceipt) Currency() CurrencyID {
	return r.CurrencyID
}

func (r FixedItemDataSizeExecutionFeeReceipt) FeeAmount() string {
	return r.TotalAmount
}

func (r FixedItemDataSizeExecutionFeeReceipt) IsValid([]byte) error {
	if err := r.BaseHinter.IsValid(FixedItemDataSizeExecutionFeeReceiptHint.Type().Bytes()); err != nil {
		return err
	}

	if r.ItemCount < 0 {
		return util.ErrInvalid.Errorf("item count under zero")
	}

	if r.DataSize < 0 {
		return util.ErrInvalid.Errorf("data size under zero")
	}

	if r.DataSizeUnit < 0 {
		return util.ErrInvalid.Errorf("data size unit under zero")
	}

	amounts, err := receiptAmountsByField(
		r.CurrencyID,
		map[string]string{
			"total_amount":          r.TotalAmount,
			"base_amount":           r.BaseAmount,
			"item_fee_amount":       r.ItemFeeAmount,
			"item_fee":              r.ItemFee,
			"data_size_fee_amount":  r.DataSizeFeeAmount,
			"data_size_fee":         r.DataSizeFee,
			"execution_fee_amount":  r.ExecutionFeeAmount,
			"execution_fee":         r.ExecutionFee,
		},
	)
	if err != nil {
		return err
	}

	expectedItemFee := amounts["item_fee_amount"].MulInt64(int64(r.ItemCount))
	if !expectedItemFee.Equal(amounts["item_fee"]) {
		return util.ErrInvalid.Errorf("item_fee does not match item_count * item_fee_amount")
	}

	expectedDataSizeFee, err := validateDataSizeFee(r.DataSize, r.DataSizeUnit, amounts["data_size_fee_amount"])
	if err != nil {
		return err
	}

	if !expectedDataSizeFee.Equal(amounts["data_size_fee"]) {
		return util.ErrInvalid.Errorf("data_size_fee does not match data_size fee formula")
	}

	if !amounts["execution_fee_amount"].Equal(amounts["execution_fee"]) {
		return util.ErrInvalid.Errorf("execution_fee does not match execution_fee_amount")
	}

	expectedTotal := amounts["base_amount"].
		Add(amounts["item_fee"]).
		Add(amounts["data_size_fee"]).
		Add(amounts["execution_fee"])
	if !expectedTotal.Equal(amounts["total_amount"]) {
		return util.ErrInvalid.Errorf("total_amount does not match fee breakdown")
	}

	return nil
}

func NewFeeReceiptFromFeeer(
	currencyID CurrencyID,
	feeer Feeer,
	itemCount int,
	dataSize int,
) (FeeReceipt, common.Big) {
	if feeer == nil {
		return nil, common.ZeroBig
	}

	switch fa := feeer.(type) {
	case *FixedItemDataSizeExecutionFeeer:
		if fa == nil {
			return nil, common.ZeroBig
		}

		baseAmount := fa.Fee()
		itemFee := fa.ItemFee(itemCount)
		dataSizeFee := fa.DataSizeFee(dataSize)
		executionFee := fa.ExecutionFee()
		totalAmount := baseAmount.Add(itemFee).Add(dataSizeFee).Add(executionFee)

		return NewFixedItemDataSizeExecutionFeeReceipt(
			currencyID,
			totalAmount,
			baseAmount,
			itemCount,
			fa.itemFeeAmount,
			itemFee,
			dataSize,
			fa.DataSizeUnit(),
			fa.dataSizeFeeAmount,
			dataSizeFee,
			fa.executionFeeAmount,
			executionFee,
		), totalAmount
	case FixedItemDataSizeExecutionFeeer:
		baseAmount := fa.Fee()
		itemFee := fa.ItemFee(itemCount)
		dataSizeFee := fa.DataSizeFee(dataSize)
		executionFee := fa.ExecutionFee()
		totalAmount := baseAmount.Add(itemFee).Add(dataSizeFee).Add(executionFee)

		return NewFixedItemDataSizeExecutionFeeReceipt(
			currencyID,
			totalAmount,
			baseAmount,
			itemCount,
			fa.itemFeeAmount,
			itemFee,
			dataSize,
			fa.DataSizeUnit(),
			fa.dataSizeFeeAmount,
			dataSizeFee,
			fa.executionFeeAmount,
			executionFee,
		), totalAmount
	case *FixedFeeer:
		if fa == nil {
			return nil, common.ZeroBig
		}

		totalAmount := fa.Fee()

		return NewFixedFeeReceipt(currencyID, totalAmount), totalAmount
	case FixedFeeer:
		totalAmount := fa.Fee()

		return NewFixedFeeReceipt(currencyID, totalAmount), totalAmount
	default:
		totalAmount := feeer.Fee()

		return NewBaseFeeReceipt(currencyID, totalAmount), totalAmount
	}
}

type CurrencyOperationReceipt struct {
	hint.BaseHinter
	Fee     FeeReceipt `json:"fee,omitempty" bson:"fee,omitempty"`
	GasUsed *uint64    `json:"gas_used,omitempty" bson:"gas_used,omitempty"`
}

func NewCurrencyOperationReceipt(fee FeeReceipt, gasUsed *uint64) CurrencyOperationReceipt {
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
		if err := r.Fee.IsValid(nil); err != nil {
			return err
		}
	}

	return nil
}

func validateReceiptAmounts(currencyID CurrencyID, amounts ...string) error {
	_, err := receiptAmountsByField(currencyID, toUnnamedReceiptAmountMap(amounts))

	return err
}

func receiptAmountsByField(currencyID CurrencyID, amounts map[string]string) (map[string]common.Big, error) {
	if err := currencyID.IsValid(nil); err != nil {
		return nil, err
	}

	parsed := make(map[string]common.Big, len(amounts))
	for field, amountString := range amounts {
		amount, err := common.NewBigFromString(amountString)
		if err != nil {
			return nil, util.ErrInvalid.Errorf("invalid %s: %v", field, err)
		}

		if !amount.OverNil() {
			return nil, util.ErrInvalid.Errorf("%s under zero", field)
		}

		parsed[field] = amount
	}

	return parsed, nil
}

func toUnnamedReceiptAmountMap(amounts []string) map[string]string {
	m := make(map[string]string, len(amounts))
	for i := range amounts {
		m[fmt.Sprintf("amount[%d]", i)] = amounts[i]
	}

	return m
}

func validateDataSizeFee(dataSize int, dataSizeUnit int64, dataSizeFeeAmount common.Big) (common.Big, error) {
	if dataSizeFeeAmount.IsZero() {
		return common.ZeroBig, nil
	}

	if dataSizeUnit < 1 {
		return common.ZeroBig, util.ErrInvalid.Errorf("data_size_unit under one for non-zero data_size_fee_amount")
	}

	bucket := (int64(dataSize) + dataSizeUnit - 1) / dataSizeUnit

	return dataSizeFeeAmount.MulInt64(bucket), nil
}
