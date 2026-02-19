package extension

import (
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
)

func (fact *UpdateRecipientFact) unpack(enc encoder.Encoder, sd, ct string, rps []string, cid string) error {
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

	recipients := make([]base.Address, len(rps))
	for i := range rps {
		switch ad, err := base.DecodeAddress(rps[i], enc); {
		case err != nil:
			return err
		default:
			recipients[i] = ad
		}
	}
	fact.recipients = recipients

	fact.currency = types.CurrencyID(cid)

	return nil
}
