package currency

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (it *BaseTransferItem) unpack(enc encoder.Encoder, ht hint.Hint, rc string, bam []byte) error {
	it.BaseHinter = hint.NewBaseHinter(ht)

	switch ad, err := base.DecodeAddress(rc, enc); {
	case err != nil:
		return err
	default:
		it.receiver = ad
	}

	ham, err := enc.DecodeSlice(bam)
	if err != nil {
		return err
	}

	amounts := make([]types.Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(types.Amount)
		if !ok {
			return common.ErrTypeMismatch.Wrap(errors.Errorf("expected Amount, not %T", ham[i]))
		}

		amounts[i] = j
	}

	it.amounts = amounts

	return nil
}
