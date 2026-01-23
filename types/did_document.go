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
	service              []Service
}

func NewDIDDocument(did DIDRef) DIDDocument {
	return DIDDocument{
		BaseHinter: hint.NewBaseHinter(DIDDocumentHint),
		context_:   []string{"https://www.w3.org/ns/did/v1", "https://imfact.im/did/contexts/v1.jsonld"},
		id:         did,
	}
}

func (d DIDDocument) IsValid([]byte) error {
	validationTarget := map[string][]VerificationRelationshipEntry{
		"authentication":       d.authentication,
		"assertionMethod":      d.assertionMethod,
		"keyAgreement":         d.keyAgreement,
		"capabilityInvocation": d.capabilityInvocation,
		"capabilityDelegation": d.capabilityDelegation,
	}

	foundMap := map[string]struct{}{}
	for tName, t := range validationTarget {
		if t != nil {
			for _, v := range t {
				var id string
				switch kind := v.Kind(); {
				case kind == VMRefKindReference:
					id = v.Ref().String()
				default:
					id = v.Method().ID().String()
				}
				if _, found := foundMap[id]; found {
					return errors.Errorf("duplicated %s id found", tName)
				}
				foundMap[id] = struct{}{}
				if err := v.IsValid(nil); err != nil {
					return err
				}
			}
		}
	}

	foundMap = map[string]struct{}{}
	if d.verificationMethod != nil {
		for _, v := range d.verificationMethod {
			if _, found := foundMap[v.ID().String()]; found {
				return errors.Errorf("duplicated verificationMethod id found")
			}
			foundMap[v.ID().String()] = struct{}{}
			if err := v.IsValid(nil); err != nil {
				return err
			}
		}
	}

	foundMap = map[string]struct{}{}
	if d.service != nil {
		for _, v := range d.service {
			if _, found := foundMap[v.ID().String()]; found {
				return errors.Errorf("duplicated service id found")
			}
			foundMap[v.ID().String()] = struct{}{}
			if err := v.IsValid(nil); err != nil {
				return err
			}
		}
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

	var bAsrt [][]byte
	for _, v := range d.assertionMethod {
		bAsrt = append(bAsrt, v.Bytes())
	}
	byteAsrt := util.ConcatBytesSlice(bAsrt...)

	var bKagr [][]byte
	for _, v := range d.keyAgreement {
		bKagr = append(bKagr, v.Bytes())
	}
	byteKagr := util.ConcatBytesSlice(bKagr...)

	var bCapInv [][]byte
	for _, v := range d.capabilityInvocation {
		bCapInv = append(bCapInv, v.Bytes())
	}
	byteCapInv := util.ConcatBytesSlice(bCapInv...)

	var bCapDlg [][]byte
	for _, v := range d.capabilityDelegation {
		bCapDlg = append(bCapDlg, v.Bytes())
	}
	byteCapDlg := util.ConcatBytesSlice(bCapDlg...)

	var bSvc [][]byte
	for _, v := range d.service {
		bSvc = append(bSvc, v.Bytes())
	}
	byteSvc := util.ConcatBytesSlice(bSvc...)

	return util.ConcatBytesSlice(
		ctx,
		d.id.Bytes(),
		byteAuth,
		byteVrf,
		byteAsrt,
		byteKagr,
		byteCapInv,
		byteCapDlg,
		byteSvc,
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
	id              DIDURLRef
	serviceType     string
	serviceEndPoint string
}

func NewService(
	id DIDURLRef, serviceType, serviceEndPoint string,
) Service {
	return Service{
		id:              id,
		serviceType:     serviceType,
		serviceEndPoint: serviceEndPoint,
	}
}

func (s Service) ID() DIDURLRef {
	return s.id
}

func (d Service) IsValid([]byte) error {
	return nil
}

func (d Service) Bytes() []byte {
	return util.ConcatBytesSlice(
		d.id.Bytes(),
		[]byte(d.serviceType),
		[]byte(d.serviceEndPoint),
	)
}
