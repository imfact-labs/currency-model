package currency

import (
	"github.com/imfact-labs/imfact-currency/common"
	"github.com/imfact-labs/imfact-currency/types"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

var maxCurrenciesCreateAccountItemMultiAmounts = 10

var (
	CreateAccountItemMultiAmountsHint = hint.MustNewHint("mitum-currency-create-account-multiple-amounts-v0.0.1")
)

type CreateAccountItemMultiAmounts struct {
	BaseCreateAccountItem
}

func NewCreateAccountItemMultiAmounts(keys types.AccountKeys, amounts []types.Amount) CreateAccountItemMultiAmounts {
	return CreateAccountItemMultiAmounts{
		BaseCreateAccountItem: NewBaseCreateAccountItem(CreateAccountItemMultiAmountsHint, keys, amounts),
	}
}

func (it CreateAccountItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseCreateAccountItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxCurrenciesCreateAccountItemMultiAmounts {
		return common.ErrValOOR.Wrap(errors.Errorf("Amounts over allowed, %d > %d", n, maxCurrenciesCreateAccountItemMultiAmounts))
	}

	return nil
}

func (it CreateAccountItemMultiAmounts) Rebuild() CreateAccountItem {
	it.BaseCreateAccountItem = it.BaseCreateAccountItem.Rebuild().(BaseCreateAccountItem)

	return it
}
