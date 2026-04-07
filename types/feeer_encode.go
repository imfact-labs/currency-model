package types

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

func (fa *FixedFeeer) unpack(enc encoder.Encoder, ht hint.Hint, rc string, am string) error {
	switch ad, err := base.DecodeAddress(rc, enc); {
	case err != nil:
		return err
	default:
		fa.receiver = ad
	}

	if big, err := common.NewBigFromString(am); err != nil {
		return err
	} else {
		fa.amount = big
	}
	fa.BaseHinter = hint.NewBaseHinter(ht)

	return nil
}

func (fa *FixedItemDataSizeExecutionFeeer) unpack(
	enc encoder.Encoder, ht hint.Hint, rc string, am, ita, dsa string, dsu int64, ea string,
) error {
	switch ad, err := base.DecodeAddress(rc, enc); {
	case err != nil:
		return err
	default:
		fa.receiver = ad
	}

	if big, err := common.NewBigFromString(am); err != nil {
		return err
	} else {
		fa.amount = big
	}

	if big, err := common.NewBigFromString(ita); err != nil {
		return err
	} else {
		fa.itemFeeAmount = big
	}
	fa.BaseHinter = hint.NewBaseHinter(ht)

	if big, err := common.NewBigFromString(dsa); err != nil {
		return err
	} else {
		fa.dataSizeFeeAmount = big
	}

	fa.dataSizeUnit = dsu

	if big, err := common.NewBigFromString(ea); err != nil {
		return err
	} else {
		fa.executionFeeAmount = big
	}

	return nil
}
