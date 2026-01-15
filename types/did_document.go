package types

import (
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

var DIDDocumentHint = hint.MustNewHint("mitum-did-document-v0.0.1")

type DIDDocument struct {
	hint.BaseHinter
	context_             []string
	id                   DIDRef
	controller           DIDRef
	verificationMethod   []IVerificationMethod
	authentication       []VerificationRelationshipEntry
	assertionMethod      []VerificationRelationshipEntry
	keyAgreement         []VerificationRelationshipEntry
	capabilityInvocation []VerificationRelationshipEntry
	capabilityDelegation []VerificationRelationshipEntry
}

func NewDIDDocument(did DIDRef) DIDDocument {
	return DIDDocument{
		BaseHinter: hint.NewBaseHinter(DIDDocumentHint),
		context_:   []string{"https://www.w3.org/ns/did/v1", "https://imfact.im/did/contexts/v1.jsonld"},
		id:         did,
	}
}

func (d DIDDocument) IsValid([]byte) error {
	foundMap := map[string]struct{}{}
	for _, v := range d.authentication {
		var id string
		switch kind := v.Kind(); {
		case kind == VMRefKindReference:
			id = v.Ref().String()
		default:
			id = v.Method().ID().String()
		}
		if _, found := foundMap[id]; found {
			return errors.Errorf("duplicated authentication id found")
		}
		foundMap[id] = struct{}{}
	}

	foundMap = map[string]struct{}{}
	for _, v := range d.verificationMethod {
		if _, found := foundMap[v.ID().String()]; found {
			return errors.Errorf("duplicated verificationMethod id found")
		}
		foundMap[v.ID().String()] = struct{}{}
	}
	return nil
}

func (d DIDDocument) Bytes() []byte {
	var ctx []byte
	for _, v := range d.context_ {
		ctx = util.ConcatBytesSlice(ctx, []byte(v))
	}

	var bAuth [][]byte
	for _, v := range d.authentication {
		bAuth = append(bAuth, v.Bytes())
	}
	byteAuth := util.ConcatBytesSlice(bAuth...)

	var bVrf [][]byte
	for _, v := range d.verificationMethod {
		bVrf = append(bVrf, v.Bytes())
	}
	byteVrf := util.ConcatBytesSlice(bVrf...)

	return util.ConcatBytesSlice(
		ctx,
		d.id.Bytes(),
		byteAuth,
		byteVrf,
	)
}

func (d DIDDocument) DID() DIDRef {
	return d.id
}

func (d *DIDDocument) SetAuthentication(auth VerificationRelationshipEntry) {
	d.authentication = append(d.authentication, auth)
}

func (d DIDDocument) Authentication(id string) (VerificationRelationshipEntry, error) {
	for _, v := range d.authentication {
		var authID string
		switch kind := v.Kind(); {
		case kind == VMRefKindReference:
			authID = v.Ref().String()
		default:
			authID = v.Method().ID().String()
		}
		if authID == id {
			return v, nil
		}
	}

	return nil, errors.Errorf("Authentication not found by id %v", id)
}

func (d DIDDocument) SetVerificationMethod() {
	d.verificationMethod = append(d.verificationMethod)
}

func (d DIDDocument) VerificationMethod(id string) (IVerificationMethod, error) {
	for _, v := range d.verificationMethod {
		if v.ID().String() == id {
			return v, nil
		}
	}

	return nil, errors.Errorf("VerificationMethod not found by id %v", id)
}

type Service struct {
	id              string
	serviceType     string
	serviceEndPoint string
}

func NewService(
	id, serviceType, serviceEndPoint string,
) Service {
	return Service{
		id:              id,
		serviceType:     serviceType,
		serviceEndPoint: serviceEndPoint,
	}
}

func (d Service) IsValid([]byte) error {
	return nil
}

func (d Service) Bytes() []byte {
	return util.ConcatBytesSlice(
		[]byte(d.id),
		[]byte(d.serviceType),
		[]byte(d.serviceEndPoint),
	)
}
