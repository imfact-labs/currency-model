package currency

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *UpdateCurrencyFact) unpack(enc encoder.Encoder, cid string, bpo []byte) error {
	if hinter, err := enc.Decode(bpo); err != nil {
		return err
	} else if po, ok := hinter.(types.CurrencyPolicy); !ok {
		return common.ErrTypeMismatch.Wrap(errors.Errorf("expected CurrencyPolicy, not %T", hinter))
	} else {
		fact.policy = po
	}

	fact.currency = types.CurrencyID(cid)

	return nil
}
