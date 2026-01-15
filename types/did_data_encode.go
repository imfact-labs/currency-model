package types

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
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
