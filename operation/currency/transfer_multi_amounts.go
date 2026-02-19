package currency

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/pkg/errors"
)

var (
	TransferItemMultiAmountsHint = hint.MustNewHint("mitum-currency-transfer-item-multi-amounts-v0.0.1")
)

var maxCurenciesTransferItemMultiAmounts = 10

type TransferItemMultiAmounts struct {
	BaseTransferItem
}

func NewTransferItemMultiAmounts(receiver base.Address, amounts []types.Amount) TransferItemMultiAmounts {
	return TransferItemMultiAmounts{
		BaseTransferItem: NewBaseTransferItem(TransferItemMultiAmountsHint, receiver, amounts),
	}
}

func (it TransferItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseTransferItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxCurenciesTransferItemMultiAmounts {
		return common.ErrValOOR.Wrap(errors.Errorf("amounts over allowed; %d > %d", n, maxCurenciesTransferItemMultiAmounts))
	}

	return nil
}

func (it TransferItemMultiAmounts) Rebuild() TransferItem {
	it.BaseTransferItem = it.BaseTransferItem.Rebuild().(BaseTransferItem)

	return it
}
