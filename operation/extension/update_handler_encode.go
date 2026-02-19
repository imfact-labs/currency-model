package extension

import (
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
)

func (fact *UpdateHandlerFact) unpack(enc encoder.Encoder, sd, ct string, hds []string, cid string) error {
	switch ad, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return err
	default:
		fact.sender = ad
	}

	switch ad, err := base.DecodeAddress(ct, enc); {
	case err != nil:
		return err
	default:
		fact.contract = ad
	}

	handlers := make([]base.Address, len(hds))
	for i := range hds {
		switch ad, err := base.DecodeAddress(hds[i], enc); {
		case err != nil:
			return err
		default:
			handlers[i] = ad
		}
	}
	fact.handlers = handlers

	fact.currency = types.CurrencyID(cid)

	return nil
}
