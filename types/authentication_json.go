package types

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/pkg/errors"
)

type VerificationMethodOrRefJSONMarshaler struct {
	hint.BaseHinter
	REF    string              `json:"ref,omitempty"`
	METHOD IVerificationMethod `json:"method,omitempty"`
	POLICY AttestationPolicy   `json:"policy,omitempty"`
}

func (v VerificationMethodOrRef) MarshalJSON() ([]byte, error) {
	var vjm VerificationMethodOrRefJSONMarshaler
	if v.kind == VMRefKindReference {
		vjm = VerificationMethodOrRefJSONMarshaler{
			BaseHinter: v.BaseHinter,
			REF:        v.ref.String(),
			POLICY:     v.policy,
		}
		return util.MarshalJSON(vjm.REF)
	} else {
		vjm = VerificationMethodOrRefJSONMarshaler{
			BaseHinter: v.BaseHinter,
			METHOD:     v.method,
			POLICY:     v.policy,
		}
	}
	return util.MarshalJSON(vjm.METHOD)
}

type VerificationMethodOrRefJSONUnmarshaler struct {
	Hint   hint.Hint         `json:"_hint"`
	REF    string            `json:"ref"`
	METHOD json.RawMessage   `json:"method"`
	POLICY AttestationPolicy `json:"policy"`
}

func (v *VerificationMethodOrRef) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("failed to decode json of %T", VerificationMethodOrRef{})

	data := bytes.TrimSpace(b)
	if len(data) == 0 {
		return e.Wrap(errors.New("VerificationMethodOrRef: empty json"))
	}

	switch data[0] {
	case '"': // ref string
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return e.Wrap(err)
		}
		ref, err := NewDIDURLRefFromString(s)
		if err != nil {
			return e.Wrap(err)
		}
		v.ref = ref
		v.method = nil
		v.kind = VMRefKindReference
		return nil

	case '{': // inline verification method object
		var vm VerificationMethod
		if err := encoder.Decode(enc, data, &vm); err != nil {
			return e.Wrap(err)
		}
		v.method = vm
		v.ref = nil
		v.kind = VMRefKindEmbedded
		return nil

	default:
		return e.Wrap(fmt.Errorf("VerificationMethodOrRef: unsupported json token: %q", data[0]))
	}
}

type BaseVerificationMethodJSONMarshaler struct {
	ID         string `json:"id"`
	Controller string `json:"controller"`
	Type       string `json:"type"`
}

func (b BaseVerificationMethod) Marshaler() BaseVerificationMethodJSONMarshaler {
	return BaseVerificationMethodJSONMarshaler{
		ID:         b.id.String(),
		Controller: b.controller.String(),
		Type:       b.verificationType.String(),
	}
}

type BaseVerificationMethodJSONUnMarshaler struct {
	ID         string `json:"id"`
	Controller string `json:"controller"`
	Type       string `json:"type"`
}

func (b BaseVerificationMethod) UnmarshalJSON(v []byte) error {
	e := util.StringError("failed to decode json of %T", BaseVerificationMethod{})

	var u BaseVerificationMethodJSONUnMarshaler
	if err := json.Unmarshal(v, &u); err != nil {
		return e.Wrap(err)
	}

	b.SetType(VerificationMethodType(u.Type))
	id, err := NewDIDURLRefFromString(u.ID)
	if err != nil {
		return e.Wrap(err)
	}
	b.SetID(*id)
	controller, err := NewDIDRefFromString(u.Controller)
	if err != nil {
		return e.Wrap(err)
	}
	b.SetController(*controller)

	return nil
}

type VerificationMethodJSONMarshaler struct {
	hint.BaseHinter
	ID                 string             `json:"id"`
	Controller         string             `json:"controller"`
	Type               string             `json:"type"`
	PublicKeyJwk       *JWK               `json:"publicKeyJwk,omitempty"`
	PublicKeyMultibase string             `json:"publicKeyMultibase,omitempty"`
	PublicKey          string             `json:"publicKeyImFact,omitempty"`
	TargetId           string             `json:"targetId,omitempty"`
	Allowed            []AllowedOperation `json:"allowed,omitempty"`
}

func (v VerificationMethod) MarshalJSON() ([]byte, error) {
	var tid string
	if v.targetID != nil {
		tid = v.targetID.String()
	} else {
		tid = ""
	}
	var pk string
	if v.PublicKey() != nil {
		pk = v.PublicKey().String()
	} else {
		pk = ""
	}
	return util.MarshalJSON(VerificationMethodJSONMarshaler{
		BaseHinter:         v.BaseHinter,
		ID:                 v.ID().String(),
		Controller:         v.Controller().String(),
		Type:               v.Type().String(),
		PublicKeyJwk:       v.PublicKeyJwk(),
		PublicKeyMultibase: v.PublicKeyMultibase(),
		PublicKey:          pk,
		TargetId:           tid,
		Allowed:            v.allowed,
	})
}

type VerificationMethodJSONUnMarshaler struct {
	PublicKeyJwk       *JWK               `json:"publicKeyJwk"`
	PublicKeyMultibase string             `json:"publicKeyMultibase"`
	PublicKey          string             `json:"publicKeyImFact"`
	TargetId           string             `json:"targetId"`
	Allowed            []AllowedOperation `json:"allowed"`
}

func (v *VerificationMethod) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("failed to decode json of %T", VerificationMethod{})

	var u BaseVerificationMethodJSONUnMarshaler
	if err := json.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	err := v.BaseVerificationMethod.unpack(u.ID, u.Type, u.Controller)
	if err != nil {
		return e.Wrap(err)
	}

	var uk VerificationMethodJSONUnMarshaler
	if err := json.Unmarshal(b, &uk); err != nil {
		return e.Wrap(err)
	}

	return v.unpack(u.Type, uk.PublicKeyJwk, uk.PublicKeyMultibase, uk.PublicKey, uk.TargetId, uk.Allowed)
}

type AllowedOperationJSONMarshaler struct {
	Contract  base.Address `json:"contract,omitempty"`
	Operation string       `json:"operation"`
}

func (a AllowedOperation) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(AllowedOperationJSONMarshaler{
		Contract:  a.contract,
		Operation: a.operation.String(),
	})
}

type AllowedOperationJSONUnMarshaler struct {
	Contract  string `json:"contract"`
	Operation string `json:"operation"`
}

func (a *AllowedOperation) UnmarshalJSON(b []byte) error {
	e := util.StringError("failed to decode json of %T", AllowedOperation{})
	var u AllowedOperationJSONUnMarshaler
	if err := json.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	return a.unpack(u.Contract, u.Operation)
}
