package types

import (
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
	currencyID CurrencyID
	totalFee   string
}

func NewBaseFeeReceipt(currencyID CurrencyID, amount common.Big) BaseFeeReceipt {
	return BaseFeeReceipt{
		BaseHinter: hint.NewBaseHinter(BaseFeeReceiptHint),
		currencyID: currencyID,
		totalFee:   amount.String(),
	}
}

func (r BaseFeeReceipt) Currency() CurrencyID {
	return r.currencyID
}

func (r BaseFeeReceipt) FeeAmount() string {
	return r.totalFee
}

func (r BaseFeeReceipt) IsValid([]byte) error {
	if err := r.BaseHinter.IsValid(BaseFeeReceiptHint.Type().Bytes()); err != nil {
		return err
	}

	if err := r.currencyID.IsValid(nil); err != nil {
		return err
	}

	_, err := parseReceiptAmount("total_fee", r.totalFee)

	return err
}

type FixedFeeReceipt struct {
	hint.BaseHinter
	currencyID CurrencyID
	totalFee   string
	baseFee    string
}

func NewFixedFeeReceipt(currencyID CurrencyID, amount common.Big) FixedFeeReceipt {
	return FixedFeeReceipt{
		BaseHinter: hint.NewBaseHinter(FixedFeeReceiptHint),
		currencyID: currencyID,
		totalFee:   amount.String(),
		baseFee:    amount.String(),
	}
}

func (r FixedFeeReceipt) Currency() CurrencyID {
	return r.currencyID
}

func (r FixedFeeReceipt) BaseFee() string { return r.baseFee }

func (r FixedFeeReceipt) FeeAmount() string {
	return r.totalFee
}

func (r FixedFeeReceipt) IsValid([]byte) error {
	if err := r.BaseHinter.IsValid(FixedFeeReceiptHint.Type().Bytes()); err != nil {
		return err
	}

	if err := r.currencyID.IsValid(nil); err != nil {
		return err
	}

	totalFee, err := parseReceiptAmount("total_fee", r.totalFee)
	if err != nil {
		return err
	}

	baseFee, err := parseReceiptAmount("base_fee", r.baseFee)
	if err != nil {
		return err
	}

	if !totalFee.Equal(baseFee) {
		return util.ErrInvalid.Errorf("total_fee and base_fee do not match")
	}

	return nil
}

type FixedItemDataSizeExecutionFeeReceipt struct {
	hint.BaseHinter
	currencyID       CurrencyID
	totalFee         string
	baseFee          string
	itemUnitFee      string
	itemCount        int
	itemFee          string
	dataSizeUnitFee  string
	dataSizeUnit     int64
	dataSize         int
	dataSizeFee      string
	executionUnitFee string
	executionCount   int
	executionFee     string
}

func NewFixedItemDataSizeExecutionFeeReceipt(
	currencyID CurrencyID,
	totalFee common.Big,
	baseFee common.Big,
	itemUnitFee common.Big,
	itemCount int,
	itemFee common.Big,
	dataSizeUnitFee common.Big,
	dataSizeUnit int64,
	dataSize int,
	dataSizeFee common.Big,
	executionUnitFee common.Big,
	executionCount int,
	executionFee common.Big,
) FixedItemDataSizeExecutionFeeReceipt {
	return FixedItemDataSizeExecutionFeeReceipt{
		BaseHinter:       hint.NewBaseHinter(FixedItemDataSizeExecutionFeeReceiptHint),
		currencyID:       currencyID,
		totalFee:         totalFee.String(),
		baseFee:          baseFee.String(),
		itemUnitFee:      itemUnitFee.String(),
		itemCount:        itemCount,
		itemFee:          itemFee.String(),
		dataSizeUnitFee:  dataSizeUnitFee.String(),
		dataSizeUnit:     dataSizeUnit,
		dataSize:         dataSize,
		dataSizeFee:      dataSizeFee.String(),
		executionUnitFee: executionUnitFee.String(),
		executionCount:   executionCount,
		executionFee:     executionFee.String(),
	}
}

func (r FixedItemDataSizeExecutionFeeReceipt) Currency() CurrencyID {
	return r.currencyID
}

func (r FixedItemDataSizeExecutionFeeReceipt) CurrencyID() CurrencyID {
	return r.currencyID
}

func (r FixedItemDataSizeExecutionFeeReceipt) FeeAmount() string {
	return r.totalFee
}

func (r FixedItemDataSizeExecutionFeeReceipt) TotalFee() string {
	return r.totalFee
}

func (r FixedItemDataSizeExecutionFeeReceipt) BaseFee() string {
	return r.baseFee
}

func (r FixedItemDataSizeExecutionFeeReceipt) ItemUnitFee() string {
	return r.itemUnitFee
}

func (r FixedItemDataSizeExecutionFeeReceipt) ItemCount() int {
	return r.itemCount
}

func (r FixedItemDataSizeExecutionFeeReceipt) ItemFee() string {
	return r.itemFee
}

func (r FixedItemDataSizeExecutionFeeReceipt) DataSizeUnitFee() string {
	return r.dataSizeUnitFee
}

func (r FixedItemDataSizeExecutionFeeReceipt) DataSizeUnit() int64 {
	return r.dataSizeUnit
}

func (r FixedItemDataSizeExecutionFeeReceipt) DataSize() int {
	return r.dataSize
}

func (r FixedItemDataSizeExecutionFeeReceipt) DataSizeFee() string {
	return r.dataSizeFee
}

func (r FixedItemDataSizeExecutionFeeReceipt) ExecutionCount() int {
	return r.executionCount
}

func (r FixedItemDataSizeExecutionFeeReceipt) ExecutionUnitFee() string {
	return r.executionUnitFee
}

func (r FixedItemDataSizeExecutionFeeReceipt) ExecutionFee() string {
	return r.executionFee
}

func (r FixedItemDataSizeExecutionFeeReceipt) IsValid([]byte) error {
	if err := r.BaseHinter.IsValid(FixedItemDataSizeExecutionFeeReceiptHint.Type().Bytes()); err != nil {
		return err
	}

	if err := r.currencyID.IsValid(nil); err != nil {
		return err
	}

	if r.itemCount < 0 {
		return util.ErrInvalid.Errorf("item count under zero")
	}

	if r.dataSize < 0 {
		return util.ErrInvalid.Errorf("data size under zero")
	}

	if r.dataSizeUnit < 0 {
		return util.ErrInvalid.Errorf("data size unit under zero")
	}

	if r.executionCount < 0 {
		return util.ErrInvalid.Errorf("execution count under zero")
	}

	totalFee, err := parseReceiptAmount("total_fee", r.totalFee)
	if err != nil {
		return err
	}

	baseFee, err := parseReceiptAmount("base_fee", r.baseFee)
	if err != nil {
		return err
	}

	itemUnitFee, err := parseReceiptAmount("item_unit_fee", r.itemUnitFee)
	if err != nil {
		return err
	}

	itemFee, err := parseReceiptAmount("item_fee", r.itemFee)
	if err != nil {
		return err
	}

	dataSizeUnitFee, err := parseReceiptAmount("data_size_unit_fee", r.dataSizeUnitFee)
	if err != nil {
		return err
	}

	dataSizeFee, err := parseReceiptAmount("data_size_fee", r.dataSizeFee)
	if err != nil {
		return err
	}

	executionUnitFee, executionUnitFeeFound, err := parseOptionalReceiptAmount("execution_unit_fee", r.executionUnitFee)
	if err != nil {
		return err
	}

	executionFee, executionFeeFound, err := parseOptionalReceiptAmount("execution_fee", r.executionFee)
	if err != nil {
		return err
	}

	if executionUnitFeeFound != executionFeeFound {
		return util.ErrInvalid.Errorf("execution_fee and execution_unit_fee must both be set or both be empty")
	}

	expectedItemFee := itemUnitFee.MulInt64(int64(r.itemCount))
	if !expectedItemFee.Equal(itemFee) {
		return util.ErrInvalid.Errorf("item_fee does not match item_count * item_unit_fee")
	}

	expectedDataSizeFee, err := validateDataSizeFee(r.dataSize, r.dataSizeUnit, dataSizeUnitFee)
	if err != nil {
		return err
	}

	if !expectedDataSizeFee.Equal(dataSizeFee) {
		return util.ErrInvalid.Errorf("data_size_fee does not match data_size fee formula")
	}

	if executionUnitFeeFound && !executionUnitFee.Equal(executionFee) {
		return util.ErrInvalid.Errorf("execution_fee does not match execution_unit_fee")
	}

	expectedTotal := baseFee.
		Add(itemFee).
		Add(dataSizeFee)
	if executionFeeFound {
		expectedTotal = expectedTotal.Add(executionFee)
	}
	if !expectedTotal.Equal(totalFee) {
		return util.ErrInvalid.Errorf("total_fee does not match fee breakdown")
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

		baseFee := fa.Fee()
		itemFee := fa.ItemFee(itemCount)
		dataSizeFee := fa.DataSizeFee(dataSize)
		executionFee := fa.ExecutionFee()
		totalFee := baseFee.Add(itemFee).Add(dataSizeFee).Add(executionFee)

		return NewFixedItemDataSizeExecutionFeeReceipt(
			currencyID,
			totalFee,
			baseFee,
			fa.itemFeeAmount,
			itemCount,
			itemFee,
			fa.dataSizeFeeAmount,
			fa.DataSizeUnit(),
			dataSize,
			dataSizeFee,
			fa.executionFeeAmount,
			0,
			executionFee,
		), totalFee
	case FixedItemDataSizeExecutionFeeer:
		baseFee := fa.Fee()
		itemFee := fa.ItemFee(itemCount)
		dataSizeFee := fa.DataSizeFee(dataSize)
		executionFee := fa.ExecutionFee()
		totalFee := baseFee.Add(itemFee).Add(dataSizeFee).Add(executionFee)

		return NewFixedItemDataSizeExecutionFeeReceipt(
			currencyID,
			totalFee,
			baseFee,
			fa.itemFeeAmount,
			itemCount,
			itemFee,
			fa.dataSizeFeeAmount,
			fa.DataSizeUnit(),
			dataSize,
			dataSizeFee,
			fa.executionFeeAmount,
			0,
			executionFee,
		), totalFee
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
	feeer   string
	Fee     FeeReceipt `json:"fee,omitempty" bson:"fee,omitempty"`
	GasUsed *uint64    `json:"gas_used,omitempty" bson:"gas_used,omitempty"`
}

func NewCurrencyOperationReceipt(feeer string, fee FeeReceipt, gasUsed *uint64) CurrencyOperationReceipt {
	return CurrencyOperationReceipt{
		BaseHinter: hint.NewBaseHinter(CurrencyOperationReceiptHint),
		feeer:      feeer,
		Fee:        fee,
		GasUsed:    gasUsed,
	}
}

func (r CurrencyOperationReceipt) Feeer() string {
	return r.feeer
}

func (r CurrencyOperationReceipt) IsValid([]byte) error {
	if err := r.BaseHinter.IsValid(CurrencyOperationReceiptHint.Type().Bytes()); err != nil {
		return err
	}

	if r.Fee != nil {
		if len(r.feeer) < 1 {
			return util.ErrInvalid.Errorf("empty feeer")
		}

		if err := r.Fee.IsValid(nil); err != nil {
			return err
		}
	}

	return nil
}

func parseReceiptAmount(field, amountString string) (common.Big, error) {
	amount, err := common.NewBigFromString(amountString)
	if err != nil {
		return common.ZeroBig, util.ErrInvalid.Errorf("invalid %s: %v", field, err)
	}

	if !amount.OverNil() {
		return common.ZeroBig, util.ErrInvalid.Errorf("%s under zero", field)
	}

	return amount, nil
}

func parseOptionalReceiptAmount(field, amountString string) (common.Big, bool, error) {
	if amountString == "" {
		return common.ZeroBig, false, nil
	}

	amount, err := parseReceiptAmount(field, amountString)
	if err != nil {
		return common.ZeroBig, false, err
	}

	return amount, true, nil
}

func validateDataSizeFee(dataSize int, dataSizeUnit int64, dataSizeUnitFee common.Big) (common.Big, error) {
	if dataSizeUnitFee.IsZero() {
		return common.ZeroBig, nil
	}

	if dataSizeUnit < 1 {
		return common.ZeroBig, util.ErrInvalid.Errorf("data_size_unit under one for non-zero data_size_unit_fee")
	}

	bucket := (int64(dataSize) + dataSizeUnit - 1) / dataSizeUnit

	return dataSizeUnitFee.MulInt64(bucket), nil
}
