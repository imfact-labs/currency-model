package types

import (
	"fmt"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonutil "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiformats/go-multibase"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	VMRefKindReference VMRefKind = iota
	VMRefKindEmbedded
)

type VMRefKind uint8

func (k VMRefKind) Bytes() []byte {
	return []byte{byte(k)}
}

type VerificationRelationshipEntry interface {
	Kind() VMRefKind
	Method() IVerificationMethod
	Ref() *DIDURLRef
	jsonutil.Decodable
	bson.ValueMarshaler
	util.Byter
	util.IsValider
}

var (
	VerificationMethodHint      = hint.MustNewHint("mitum-did-verification-method-v0.0.1")
	VerificationMethodOrRefHint = hint.MustNewHint("mitum-did-verification-method-or-ref-v0.0.1")
)

type VerificationMethodType string

func (k VerificationMethodType) String() string {
	return string(k)
}

const (
	AuthTypeED25519   = VerificationMethodType("Ed25519VerificationKey2020")
	AuthTypeECDSASECP = VerificationMethodType("EcdsaSecp256k1VerificationKey2019")
	AuthTypeImFact    = VerificationMethodType("EcdsaSecp256k1VerificationKeyImFact2025")
	AuthTypeLinked    = VerificationMethodType("LinkedVerificationMethod")
)

type AllowedOperation struct {
	contract  base.Address
	operation hint.Hint
}

func NewAllowedOperation(contract base.Address, operation hint.Hint) *AllowedOperation {
	return &AllowedOperation{
		contract:  contract,
		operation: operation,
	}
}

func (a AllowedOperation) Bytes() []byte {
	if a.contract == nil {
		return util.ConcatBytesSlice(
			a.operation.Bytes())
	}

	return util.ConcatBytesSlice(
		a.contract.Bytes(),
		a.operation.Bytes(),
	)
}

func (a AllowedOperation) Equal(b AllowedOperation) bool {
	if a.contract == nil {
		return a.operation.Equal(b.operation)
	}
	if !a.contract.Equal(b.contract) {
		return false
	} else if !a.operation.Equal(b.operation) {
		return false
	}

	return true
}

type AttestationPolicy interface {
	PolicyType() string
	AllowedOperations() []AllowedOperation
	util.Byter
	util.IsValider
	bson.Marshaler
}

type VerificationMethodOrRef struct {
	hint.BaseHinter
	ref    *DIDURLRef
	method IVerificationMethod
	kind   VMRefKind
	policy AttestationPolicy
}

func NewVerificationMethodOrRef() *VerificationMethodOrRef {
	return &VerificationMethodOrRef{
		BaseHinter: hint.NewBaseHinter(VerificationMethodOrRefHint),
	}
}

func (v *VerificationMethodOrRef) SetRef(ref *DIDURLRef) *VerificationMethodOrRef {
	v.ref = ref
	v.kind = VMRefKindReference
	return v
}

func (v *VerificationMethodOrRef) SetVerificationMethod(vrfm IVerificationMethod) {
	v.method = vrfm
	v.kind = VMRefKindEmbedded
}

func (v VerificationMethodOrRef) Kind() VMRefKind {
	return v.kind
}

func (v VerificationMethodOrRef) Method() IVerificationMethod {
	return v.method
}

func (v VerificationMethodOrRef) Ref() *DIDURLRef {
	return v.ref
}

func (v VerificationMethodOrRef) Bytes() []byte {
	var buf [][]byte
	if v.kind == VMRefKindReference {
		buf = append(buf, v.ref.Bytes())
	} else {
		buf = append(buf, v.method.Bytes())
	}
	buf = append(buf, v.kind.Bytes())
	if v.policy != nil {
		buf = append(buf, v.policy.Bytes())
	}

	return util.ConcatBytesSlice(buf...)
}

func (v VerificationMethodOrRef) IsValid([]byte) error {
	return nil
}

type IVerificationMethod interface {
	ID() DIDURLRef
	Controller() DIDRef
	Type() VerificationMethodType
	util.Byter
	util.IsValider
}

type BaseVerificationMethod struct {
	id               DIDURLRef
	verificationType VerificationMethodType
	controller       DIDRef
}

func NewBaseVerificationMethod(id DIDURLRef, controller DIDRef) BaseVerificationMethod {
	return BaseVerificationMethod{
		id:         id,
		controller: controller,
	}
}

func (b *BaseVerificationMethod) SetID(id DIDURLRef) {
	b.id = id
}
func (b BaseVerificationMethod) ID() DIDURLRef { return b.id }
func (b *BaseVerificationMethod) SetController(controller DIDRef) {
	b.controller = controller
}
func (b BaseVerificationMethod) Controller() DIDRef { return b.controller }
func (b *BaseVerificationMethod) SetType(verificationType VerificationMethodType) {
	b.verificationType = verificationType
}
func (b BaseVerificationMethod) Type() VerificationMethodType { return b.verificationType }
func (b BaseVerificationMethod) Bytes() []byte {
	return util.ConcatBytesSlice(
		b.id.Bytes(),
		[]byte(b.verificationType.String()),
		b.controller.Bytes())
}

type VerificationMethod struct {
	hint.BaseHinter
	BaseVerificationMethod
	publicKeyJwk       *JWK
	publicKeyMultibase string
	publicKey          base.Publickey
	targetID           *DIDURLRef
	allowed            []AllowedOperation
}

func NewVerificationMethod(id DIDURLRef, controller DIDRef) VerificationMethod {
	b := NewBaseVerificationMethod(id, controller)
	return VerificationMethod{
		BaseHinter:             hint.NewBaseHinter(VerificationMethodHint),
		BaseVerificationMethod: b,
	}
}

func (v *VerificationMethod) SetPublicKeyJwk(jwk *JWK) {
	v.publicKeyJwk = jwk
}

func (v VerificationMethod) PublicKeyJwk() *JWK {
	return v.publicKeyJwk
}

func (v *VerificationMethod) SetPublicKeyMultibase(publicKey base.Publickey) error {
	if v.Type() == AuthTypeECDSASECP && v.publicKey != nil {

	}
	if publicKey == nil {
		return fmt.Errorf("publicKey is nil")
	}

	MEPbKey, err := ParseMEPublickey(publicKey.String())
	if err != nil {
		return err
	}
	var Secp256k1PubPrefix = []byte{0xe7, 0x01}
	compressedBytes := crypto.CompressPubkey(MEPbKey.k)
	data := append(Secp256k1PubPrefix, compressedBytes...)
	encoded, err := multibase.Encode(multibase.Base58BTC, data)
	if err != nil {
		return fmt.Errorf("multibase encoding failed: %w", err)
	}
	v.publicKeyMultibase = encoded

	return nil
}

func (v VerificationMethod) PublicKeyMultibase() string {
	return v.publicKeyMultibase
}

func (v *VerificationMethod) SetPublicKey(publicKey base.Publickey) {
	v.publicKey = publicKey
}

func (v VerificationMethod) PublicKey() base.Publickey {
	return v.publicKey
}

func (v *VerificationMethod) SetTargetID(id *DIDURLRef) {
	v.targetID = id
}
func (v VerificationMethod) TargetID() *DIDURLRef { return v.targetID }

func (v *VerificationMethod) SetAllowed(operations []AllowedOperation) {
	v.allowed = operations
}

func (v VerificationMethod) IsAllowed(operation AllowedOperation) bool {
	for _, op := range v.allowed {
		if operation.Equal(op) {
			return true
		}
	}
	return false
}

func (v VerificationMethod) Allowed() []AllowedOperation {
	return v.allowed
}

func (v VerificationMethod) IsValid([]byte) error {
	if v.Type() == AuthTypeECDSASECP {
		if v.publicKeyMultibase == "" {
			return fmt.Errorf("EcdsaSecp256k1VerificationKey2019 type must have publicKeyMultibase")
		}
	}
	if v.publicKey != nil && v.publicKeyMultibase != "" {
		pbKey, ok := v.publicKey.(MEPublickey)
		if !ok {
			return fmt.Errorf("verification method publicKey is not a MEPublickey")
		}
		var Secp256k1PubPrefix = []byte{0xe7, 0x01}
		compressedBytes := crypto.CompressPubkey(pbKey.k)
		data := append(Secp256k1PubPrefix, compressedBytes...)
		encoded, err := multibase.Encode(multibase.Base58BTC, data)
		if err != nil {
			return fmt.Errorf("multibase encoding failed: %w", err)
		}
		if v.publicKeyMultibase != encoded {
			return fmt.Errorf("verification method publicKey is not matched with publicKeyMultibase")
		}
	}

	return nil
}

func (v VerificationMethod) Bytes() []byte {
	var pbKey [][]byte
	if v.publicKeyJwk != nil {
		pbKey = append(pbKey, v.publicKeyJwk.Bytes())
	}
	if v.publicKeyMultibase != "" {
		pbKey = append(pbKey, []byte(v.publicKeyMultibase))
	}
	if v.publicKey != nil {
		pbKey = append(pbKey, v.publicKey.Bytes())
	}
	s := util.ConcatBytesSlice(pbKey...)

	var allowed [][]byte
	for _, op := range v.allowed {
		allowed = append(allowed, op.Bytes())
	}
	a := util.ConcatBytesSlice(allowed...)

	return util.ConcatBytesSlice(
		v.BaseVerificationMethod.Bytes(),
		s,
		a)
}

type JWK struct {
	Kty string `json:"kty" bson:"kty"`
	Crv string `json:"crv" bson:"crv"`
	X   string `json:"x" bson:"x"`
	Y   string `json:"y" bson:"y"`
}

func NewJWK(kty string, crv string, x string, y string) JWK {
	return JWK{
		Kty: kty,
		Crv: crv,
		X:   x,
		Y:   y,
	}
}

func (j JWK) Bytes() []byte {
	return util.ConcatBytesSlice(
		[]byte(j.Kty),
		[]byte(j.Crv),
		[]byte(j.X),
		[]byte(j.Y),
	)
}
