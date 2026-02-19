package isaacoperation

import (
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
)

func (fact *SuffrageJoinFact) unpack(
	enc encoder.Encoder,
	candidate string,
	height base.Height,
) error {
	switch i, err := base.DecodeAddress(candidate, enc); {
	case err != nil:
		return err
	default:
		fact.candidate = i
	}

	fact.start = height

	return nil
}
