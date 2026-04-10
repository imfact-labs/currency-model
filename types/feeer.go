package types

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
)

const (
	FeeerNil                        = "nil"
	FeeerFixed                      = "fixed"
	FeeerFixedItemDataSizeExecution = "fixed-item-data-size-execution"
)

var (
	NilFeeerHint                        = hint.MustNewHint("mitum-currency-nil-feeer-v0.0.1")
	FixedFeeerHint                      = hint.MustNewHint("mitum-currency-fixed-feeer-v0.0.1")
	FixedItemDataSizeExecutionFeeerHint = hint.MustNewHint("mitum-currency-fixed-item-data-size-execution-feeer-v0.0.1")
)

var UnlimitedMaxFeeAmount = common.NewBig(-1)

type Feeer interface {
	util.IsValider
	hint.Hinter
	Type() string
	Bytes() []byte
	Receiver() base.Address
	Min() common.Big
	Fee() common.Big
}

type ItemFeeer interface {
	ItemFee(int) common.Big
}

type DataSizeFeeer interface {
	DataSizeFee(int) common.Big
	DataSizeUnit() int64
}

type ExecutionFeeer interface {
	ExecutionFee() common.Big
}

type ExtFeeer interface {
	Feeer
	ItemFeeer
	DataSizeFeeer
	ExecutionFeeer
}

type NilFeeer struct {
	hint.BaseHinter
}

func NewNilFeeer() NilFeeer {
	return NilFeeer{BaseHinter: hint.NewBaseHinter(NilFeeerHint)}
}

func (NilFeeer) Type() string {
	return FeeerNil
}

func (NilFeeer) Bytes() []byte {
	return nil
}

func (NilFeeer) Receiver() base.Address {
	return nil
}

func (NilFeeer) Min() common.Big {
	return common.ZeroBig
}

func (NilFeeer) Fee() common.Big {
	return common.ZeroBig
}

func (fa NilFeeer) IsValid([]byte) error {
	return fa.BaseHinter.IsValid(nil)
}

type FixedFeeer struct {
	hint.BaseHinter
	receiver base.Address
	amount   common.Big
}

func NewFixedFeeer(receiver base.Address, amount common.Big) FixedFeeer {
	return FixedFeeer{
		BaseHinter: hint.NewBaseHinter(FixedFeeerHint),
		receiver:   receiver,
		amount:     amount,
	}
}

func (FixedFeeer) Type() string {
	return FeeerFixed
}

func (fa FixedFeeer) Bytes() []byte {
	return util.ConcatBytesSlice(fa.receiver.Bytes(), fa.amount.Bytes())
}

func (fa FixedFeeer) Receiver() base.Address {
	return fa.receiver
}

func (fa FixedFeeer) Min() common.Big {
	return fa.amount
}

func (fa FixedFeeer) Fee() common.Big {
	if fa.isZero() {
		return common.ZeroBig
	}

	return fa.amount
}

func (fa FixedFeeer) IsValid([]byte) error {
	if err := fa.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := util.CheckIsValiders(nil, false, fa.receiver); err != nil {
		return util.ErrInvalid.Errorf("invalid receiver for fixed feeer: %v", err)
	}

	if !fa.amount.OverNil() {
		return util.ErrInvalid.Errorf("fixed feeer amount under zero")
	}

	return nil
}

func (fa FixedFeeer) isZero() bool {
	return fa.amount.IsZero()
}

type FixedItemDataSizeExecutionFeeer struct {
	hint.BaseHinter
	receiver           base.Address
	amount             common.Big
	itemFeeAmount      common.Big
	dataSizeFeeAmount  common.Big
	dataSizeUnit       int64
	executionFeeAmount common.Big
}

func NewFixedItemDataSizeExecutionFeeer(
	receiver base.Address, amount, itemFeeAmount, dataSizeFeeAmount common.Big,
	dataSizeUnit int64, executionFeeAmount common.Big,
) FixedItemDataSizeExecutionFeeer {
	return FixedItemDataSizeExecutionFeeer{
		BaseHinter:         hint.NewBaseHinter(FixedItemDataSizeExecutionFeeerHint),
		receiver:           receiver,
		amount:             amount,
		itemFeeAmount:      itemFeeAmount,
		dataSizeFeeAmount:  dataSizeFeeAmount,
		dataSizeUnit:       dataSizeUnit,
		executionFeeAmount: executionFeeAmount,
	}
}

func (FixedItemDataSizeExecutionFeeer) Type() string {
	return FeeerFixedItemDataSizeExecution
}

func (fa FixedItemDataSizeExecutionFeeer) Bytes() []byte {
	return util.ConcatBytesSlice(fa.receiver.Bytes(), fa.amount.Bytes())
}

func (fa FixedItemDataSizeExecutionFeeer) Receiver() base.Address {
	return fa.receiver
}

func (fa FixedItemDataSizeExecutionFeeer) Min() common.Big {
	return fa.amount
}

func (fa FixedItemDataSizeExecutionFeeer) Fee() common.Big {
	if fa.isZero(fa.amount) {
		return common.ZeroBig
	}

	return fa.amount
}

func (fa FixedItemDataSizeExecutionFeeer) ItemFee(items int) common.Big {
	if fa.isZero(fa.itemFeeAmount) {
		return common.ZeroBig
	}

	return fa.itemFeeAmount.MulInt64(int64(items))
}

func (fa FixedItemDataSizeExecutionFeeer) DataSizeFee(size int) common.Big {
	if fa.isZero(fa.dataSizeFeeAmount) {
		return common.ZeroBig
	}

	unit := fa.dataSizeUnit
	if unit < 1 {
		return common.ZeroBig
	}

	bucket := (int64(size) + unit - 1) / unit

	return fa.dataSizeFeeAmount.MulInt64(bucket)
}

func (fa FixedItemDataSizeExecutionFeeer) DataSizeUnit() int64 {
	if fa.dataSizeUnit < 1 {
		return 0
	}

	return fa.dataSizeUnit
}

func (fa FixedItemDataSizeExecutionFeeer) ExecutionFee() common.Big {
	if fa.isZero(fa.executionFeeAmount) {
		return common.ZeroBig
	}

	return fa.executionFeeAmount
}

func (fa FixedItemDataSizeExecutionFeeer) IsValid([]byte) error {
	if err := fa.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := util.CheckIsValiders(nil, false, fa.receiver); err != nil {
		return util.ErrInvalid.Errorf("invalid receiver for fixed item feeer: %v", err)
	}

	if !fa.amount.OverNil() {
		return util.ErrInvalid.Errorf("fixed item feeer amount under zero")
	}

	if !fa.itemFeeAmount.OverNil() {
		return util.ErrInvalid.Errorf("fixed item feeer item amount under zero")
	}

	if !fa.dataSizeFeeAmount.OverNil() {
		return util.ErrInvalid.Errorf("fixed item feeer data size amount under zero")
	}

	if fa.dataSizeFeeAmount.OverZero() && fa.dataSizeUnit < 1 {
		return util.ErrInvalid.Errorf("fixed item feeer data size unit under one")
	}

	if !fa.executionFeeAmount.OverNil() {
		return util.ErrInvalid.Errorf("fixed item feeer execution amount under zero")
	}

	return nil
}

func (fa FixedItemDataSizeExecutionFeeer) isZero(am common.Big) bool {
	return am.IsZero()
}
