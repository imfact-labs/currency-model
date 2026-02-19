package types

import (
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

func (d *Data) unpack(
	enc encoder.Encoder,
	ht hint.Hint,
	pubKey string, did DIDRef,
) error {
	d.BaseHinter = hint.NewBaseHinter(ht)
	a, err := base.DecodeAddress(pubKey, enc)
	if err != nil {
		return err
	}
	d.address = a

	d.did = did

	return nil
}
