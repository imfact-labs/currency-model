package types

import (
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

type DataJSONMarshaler struct {
	hint.BaseHinter
	Address string `json:"address"`
	DID     string `json:"did"`
}

func (d Data) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(DataJSONMarshaler{
		BaseHinter: d.BaseHinter,
		Address:    d.address.String(),
		DID:        d.did.String(),
	})
}

type DataJSONUnmarshaler struct {
	Hint    hint.Hint `json:"_hint"`
	Address string    `json:"address"`
	DID     string    `json:"did"`
}

func (d *Data) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("failed to decode json of Data")

	var u DataJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	did, err := NewDIDRefFromString(u.DID)
	if err != nil {
		return e.Wrap(err)
	}

	return d.unpack(enc, u.Hint, u.Address, *did)
}
