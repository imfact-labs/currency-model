package types

import (
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

type DesignJSONMarshaler struct {
	hint.BaseHinter
	DIDMethod string `json:"didMethod"`
}

func (de Design) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(DesignJSONMarshaler{
		BaseHinter: de.BaseHinter,
		DIDMethod:  de.didMethod,
	})
}

type DesignJSONUnmarshaler struct {
	Hint      hint.Hint `json:"_hint"`
	DIDMethod string    `json:"didMethod"`
}

func (de *Design) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("failed to decode json of Design")

	var u DesignJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	return de.unmarshal(enc, u.Hint, u.DIDMethod)
}
