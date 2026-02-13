package currency

import (
	"github.com/imfact-labs/imfact-currency/common"
	"github.com/imfact-labs/imfact-currency/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *MintFact) unpack(enc encoder.Encoder, rc string, bam []byte) error {
	switch ad, err := base.DecodeAddress(rc, enc); {
	case err != nil:
		return err
	default:
		fact.receiver = ad
	}

	if hinter, err := enc.Decode(bam); err != nil {
		return err
	} else if am, ok := hinter.(types.Amount); !ok {
		return common.ErrTypeMismatch.Wrap(errors.Errorf("expected InitialSupply, not %T", hinter))
	} else {
		fact.amount = am
	}

	return nil
}
