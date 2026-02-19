package state

import (
	"encoding/json"
	"github.com/imfact-labs/currency-model/types"
	"github.com/pkg/errors"

	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

type DesignStateValueJSONMarshaler struct {
	hint.BaseHinter
	Design types.Design `json:"design"`
}

func (sv DesignStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(
		DesignStateValueJSONMarshaler(sv),
	)
}

type DesignStateValueJSONUnmarshaler struct {
	Hint   hint.Hint       `json:"_hint"`
	Design json.RawMessage `json:"design"`
}

func (sv *DesignStateValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("failed to decode json of DesignStateValue")

	var u DesignStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	sv.BaseHinter = hint.NewBaseHinter(u.Hint)

	var sd types.Design
	if err := sd.DecodeJSON(u.Design, enc); err != nil {
		return e.Wrap(err)
	}
	sv.Design = sd

	return nil
}

type DIDDataStateValueJSONMarshaler struct {
	hint.BaseHinter
	Data types.Data `json:"did_data"`
}

func (sv DataStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(
		DIDDataStateValueJSONMarshaler(sv),
	)
}

type DataStateValueJSONUnmarshaler struct {
	Hint    hint.Hint       `json:"_hint"`
	DIDData json.RawMessage `json:"did_data"`
}

func (sv *DataStateValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode json of DataStateValue")

	var u DataStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	sv.BaseHinter = hint.NewBaseHinter(u.Hint)

	var t types.Data
	if err := t.DecodeJSON(u.DIDData, enc); err != nil {
		return e.Wrap(err)
	}
	sv.Data = t

	return nil
}

type DIDDocumentStateValueJSONMarshaler struct {
	hint.BaseHinter
	Document types.DIDDocument `json:"did_document"`
}

func (sv DocumentStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(
		DIDDocumentStateValueJSONMarshaler(sv),
	)
}

type DocumentStateValueJSONUnmarshaler struct {
	Hint        hint.Hint       `json:"_hint"`
	DIDDocument json.RawMessage `json:"did_document"`
}

func (sv *DocumentStateValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode json of DocumentStateValue")

	var u DocumentStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	sv.BaseHinter = hint.NewBaseHinter(u.Hint)

	if t, err := enc.Decode(u.DIDDocument); err != nil {
		return e.Wrap(err)
	} else if v, ok := t.(types.DIDDocument); !ok {
		return e.Wrap(errors.Errorf("expected DIDDocument, but %T", t))
	} else {
		sv.Document = v
	}

	return nil
}
