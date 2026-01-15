package types

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

func (v VerificationMethodOrRef) MarshalBSONValue() (bsontype.Type, []byte, error) {
	switch v.kind {
	case VMRefKindReference:
		if v.ref == nil {
			return 0, nil, errors.New("reference not set")
		}
		return bson.MarshalValue(v.ref.String())
	case VMRefKindEmbedded:
		if v.method == nil {
			return 0, nil, errors.New("method not set")
		}
		return bson.MarshalValue(v.method)
	default:
		return 0, nil, errors.Errorf("unknown kind: %v", v.kind)
	}
}

type VerificationMethodOrRefBSONUnmarshaler struct {
	Hint   string            `bson:"_hint"`
	REF    string            `bson:"ref"`
	METHOD bson.Raw          `bson:"method"`
	KIND   VMRefKind         `bson:"kind"`
	POLICY AttestationPolicy `bson:"policy"`
}

func (v *VerificationMethodOrRef) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	switch t {
	case bson.TypeString:
		var s string
		if err := bson.UnmarshalValue(t, data, &s); err != nil {
			return err
		}
		ref, err := NewDIDURLRefFromString(s)
		if err != nil {
			return err
		}
		v.ref = ref
		v.method = nil
		v.kind = VMRefKindReference
		return nil

	case bson.TypeEmbeddedDocument:
		var vm VerificationMethod
		if err := bson.UnmarshalValue(t, data, &vm); err != nil {
			return err
		}
		v.method = vm
		v.ref = nil
		v.kind = VMRefKindEmbedded
		return nil

	default:
		return errors.New("unsupported bson type")
	}
}

type BaseVerificationMethodBSONMarshaler struct {
	ID         string `bson:"id"`
	Controller string `bson:"controller"`
	Type       string `bson:"type"`
}

func (b BaseVerificationMethod) BSONMarshaler() BaseVerificationMethodBSONMarshaler {
	return BaseVerificationMethodBSONMarshaler{
		ID:         b.id.String(),
		Controller: b.controller.String(),
		Type:       b.verificationType.String(),
	}
}

type BaseVerificationMethodBSONUnMarshaler struct {
	ID         string `bson:"id"`
	Controller string `bson:"controller"`
	Type       string `bson:"type"`
}

func (b BaseVerificationMethod) UnmarshalBSON(v []byte) error {
	e := util.StringError("failed to decode bson of %T", BaseVerificationMethod{})

	var u BaseVerificationMethodBSONUnMarshaler
	if err := bson.Unmarshal(v, &u); err != nil {
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

type VerificationMethodBSONMarshaler struct {
	Hint               string             `bson:"_hint"`
	ID                 string             `bson:"id"`
	Controller         string             `bson:"controller"`
	Type               string             `bson:"type"`
	PublicKeyJwk       *JWK               `bson:"publicKeyJwk,omitempty"`
	PublicKeyMultibase string             `bson:"publicKeyMultibase,omitempty"`
	PublicKey          string             `bson:"publicKeyImFact,omitempty"`
	TargetId           string             `bson:"targetId,omitempty"`
	Allowed            []AllowedOperation `bson:"allowed,omitempty"`
}

func (v VerificationMethod) MarshalBSON() ([]byte, error) {
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

	return bsonenc.Marshal(VerificationMethodBSONMarshaler{
		Hint:               v.Hint().String(),
		ID:                 v.ID().String(),
		Controller:         v.Controller().String(),
		Type:               v.Type().String(),
		PublicKeyJwk:       v.PublicKeyJwk(),
		PublicKeyMultibase: v.PublicKeyMultibase(),
		PublicKey:          pk,
		TargetId:           tid,
		Allowed:            v.Allowed(),
	})
}

type VerificationMethodBSONUnMarshaler struct {
	Hint               string             `bson:"_hint"`
	PublicKeyJwk       *JWK               `bson:"publicKeyJwk"`
	PublicKeyMultibase string             `bson:"publicKeyMultibase"`
	PublicKey          string             `bson:"publicKeyImFact"`
	TargetId           string             `bson:"targetId"`
	Allowed            []AllowedOperation `bson:"allowed"`
}

func (v *VerificationMethod) UnmarshalBSON(b []byte) error {
	e := util.StringError("failed to decode bson of %T", VerificationMethod{})

	var u BaseVerificationMethodBSONUnMarshaler
	if err := bson.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	err := v.BaseVerificationMethod.unpack(u.ID, u.Type, u.Controller)
	if err != nil {
		return e.Wrap(err)
	}

	var uk VerificationMethodBSONUnMarshaler
	if err := bson.Unmarshal(b, &uk); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uk.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	v.BaseHinter = hint.NewBaseHinter(ht)

	return v.unpack(uk.PublicKeyJwk, uk.PublicKeyMultibase, uk.PublicKey, uk.TargetId, uk.Allowed)
}

type AllowedOperationBSONMarshaler struct {
	Contract  string `bson:"contract,omitempty"`
	Operation string `bson:"operation"`
}

func (a AllowedOperation) MarshalBSON() ([]byte, error) {
	var contract string
	if a.contract != nil {
		contract = a.contract.String()
	} else {
		contract = ""
	}
	return bsonenc.Marshal(AllowedOperationBSONMarshaler{
		Contract:  contract,
		Operation: a.operation.String(),
	})
}

type AllowedOperationBSONUnMarshaler struct {
	Contract  string `bson:"contract"`
	Operation string `bson:"operation"`
}

func (a *AllowedOperation) UnmarshalBSON(b []byte) error {
	e := util.StringError("failed to decode bson of %T", AllowedOperation{})
	var u AllowedOperationBSONUnMarshaler
	if err := bson.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	return a.unpack(u.Contract, u.Operation)
}
