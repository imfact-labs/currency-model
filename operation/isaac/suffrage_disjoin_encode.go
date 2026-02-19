package isaacoperation

import (
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
)

func (fact *SuffrageDisjoinFact) unpack(
	enc encoder.Encoder,
	nd string,
	height base.Height,
) error {
	switch i, err := base.DecodeAddress(nd, enc); {
	case err != nil:
		return err
	default:
		fact.node = i
	}

	fact.start = height

	return nil
}
