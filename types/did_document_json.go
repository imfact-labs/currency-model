package types

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

type DIDDocumentJSONMarshaler struct {
	hint.BaseHinter
	Context_  []string                        `json:"@context"`
	ID        string                          `json:"id"`
	Auth      []VerificationRelationshipEntry `json:"authentication,omitempty"`
	VRFMethod []IVerificationMethod           `json:"verificationMethod,omitempty"`
}

func (d DIDDocument) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(DIDDocumentJSONMarshaler{
		BaseHinter: d.BaseHinter,
		Context_:   d.context_,
		ID:         d.id.String(),
		Auth:       d.authentication,
		VRFMethod:  d.verificationMethod,
	})
}

type DIDDocumentJSONUnmarshaler struct {
	Hint      hint.Hint       `json:"_hint"`
	Context_  []string        `json:"@context"`
	ID        string          `json:"id"`
	Auth      json.RawMessage `json:"authentication"`
	VRFMethod json.RawMessage `json:"verificationMethod"`
}

func (d *DIDDocument) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("failed to decode json of %T", DIDDocument{})

	var u DIDDocumentJSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	d.BaseHinter = hint.NewBaseHinter(u.Hint)

	if u.Auth != nil {
		var bAuth []json.RawMessage
		err := json.Unmarshal(u.Auth, &bAuth)
		if err != nil {
			return e.Wrap(err)
		}

		auths := make([]VerificationRelationshipEntry, len(bAuth))
		for i, hinter := range bAuth {
			var vrfR VerificationMethodOrRef
			err := vrfR.DecodeJSON(hinter, enc)
			if err != nil {
				return e.Wrap(err)
			}

			if err := vrfR.IsValid(nil); err != nil {
				return e.Wrap(err)
			} else {
				auths[i] = &vrfR
			}
		}
		d.authentication = auths
	}

	hr, err := enc.DecodeSlice(u.VRFMethod)
	if err != nil {
		return e.Wrap(err)
	}

	vrfs := make([]IVerificationMethod, len(hr))
	for i, hinter := range hr {
		if v, ok := hinter.(IVerificationMethod); !ok {
			return e.Wrap(errors.Errorf("expected DIDVerificationMethod, not %T", hinter))
		} else {
			if err := v.IsValid(nil); err != nil {
				return e.Wrap(err)
			} else {
				vrfs[i] = v
			}
		}

	}
	d.verificationMethod = vrfs
	err = d.unpack(u.Context_, u.ID)
	if err != nil {
		return e.Wrap(err)
	}

	return nil
}

type ServiceJSONMarshaler struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndPoint string `json:"service_end_point"`
}

func (d Service) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(ServiceJSONMarshaler{
		ID:              d.id,
		Type:            d.serviceType,
		ServiceEndPoint: d.serviceEndPoint,
	})
}

type ServiceJSONUnmarshaler struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndPoint string `json:"service_end_point"`
}

func (d *Service) UnmarshalJSON(b []byte) error {
	e := util.StringError("failed to decode json of Service")

	var u ServiceJSONUnmarshaler
	if err := json.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	return d.unpack(u.ID, u.Type, u.ServiceEndPoint)
}
