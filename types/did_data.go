package types

import (
	"fmt"
	"strings"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/pkg/errors"
)

const DIDPrefix = "did"
const DIDSeparator = ":"

var DataHint = hint.MustNewHint("mitum-did-data-v0.0.1")

type Data struct {
	hint.BaseHinter
	address base.Address
	did     DIDRef
}

func NewData(
	address base.Address, method string,
) (*Data, error) {
	data := Data{
		BaseHinter: hint.NewBaseHinter(DataHint),
	}
	data.address = address
	var didRef *DIDRef
	var err error
	if didRef, err = NewDIDRef(method, address.String()); err != nil {
		return nil, err
	}

	data.did = *didRef
	if err := data.IsValid(nil); err != nil {
		return nil, err
	}

	return &data, nil
}

func (d Data) IsValid([]byte) error {
	if err := d.address.IsValid(nil); err != nil {
		return err
	}
	if err := d.did.IsValid(nil); err != nil {
		return err
	}
	return nil
}

func (d Data) Bytes() []byte {
	return util.ConcatBytesSlice(
		d.address.Bytes(),
		d.did.Bytes(),
	)
}

func (d Data) Address() base.Address {
	return d.address
}

func (d Data) DID() DIDRef {
	return d.did
}

func (d Data) Equal(b Data) bool {
	if d.address.Equal(b.address) {
		return false
	}
	if d.did.String() != b.did.String() {
		return false
	}

	return true
}

type DIDRef string

func NewDIDRef(method, methodSpecificID string) (*DIDRef, error) {
	didRef := DIDRef(fmt.Sprintf("%s:%s:%s", DIDPrefix, method, methodSpecificID))
	if err := didRef.IsValid(nil); err != nil {
		return nil, common.ErrValueInvalid.Wrap(err)
	}
	return &didRef, nil
}

func NewDIDRefFromString(s string) (*DIDRef, error) {
	didRef := DIDRef(s)
	if err := didRef.IsValid(nil); err != nil {
		return nil, common.ErrValueInvalid.Wrap(err)
	}

	return &didRef, nil
}

func (d DIDRef) String() string {
	return string(d)
}

func (d DIDRef) Method() string {
	parts := strings.SplitN(string(d), ":", 3)

	return parts[1]
}

func (d DIDRef) MethodSpecificID() string {
	parts := strings.SplitN(string(d), ":", 3)

	return parts[2]
}

func (d DIDRef) Bytes() []byte {
	return []byte(string(d))
}

func (d DIDRef) IsValid(_ []byte) error {
	s := string(d)

	if len(s) < 1 {
		return errors.Errorf("empty DIDRef")
	}

	if !strings.HasPrefix(s, "did:") {
		return errors.Errorf("invalid DID: %q (missing 'did:' prefix)", s)
	}

	parts := strings.SplitN(s, ":", 3)
	if len(parts) < 3 {
		return errors.Errorf("invalid DID: %q (expected 'did:<method>:<id>')", s)
	}

	method := parts[1]
	id := parts[2]

	if method == "" {
		return errors.Errorf("invalid DID: %q (empty method)", s)
	}
	if id == "" {
		return errors.Errorf("invalid DID: %q (empty method-specific-id)", s)
	}

	return nil
}

type DIDURLRef struct {
	uri      string // "did:example:alice#key-1"
	did      DIDRef // "did:example:alice"
	fragment string // "key-1"
}

func NewDIDURLRef(did, fragment string) (*DIDURLRef, error) {
	didRef, err := NewDIDRefFromString(did)
	if err != nil {
		return nil, err
	}

	if fragment == "" {
		return nil, errors.Errorf("empty fragment")
	}
	uri := fmt.Sprintf("%s#%s", did, fragment)

	didURLRef := &DIDURLRef{
		uri:      uri,
		did:      *didRef,
		fragment: fragment,
	}

	return didURLRef, nil
}

func NewDIDURLRefFromString(s string) (*DIDURLRef, error) {
	parsed := strings.Split(s, "#")
	if len(parsed) != 2 {
		return nil, errors.Errorf("DIDURLRef must have a fragment: %q", s)
	}
	if parsed[1] == "" {
		return nil, errors.Errorf("DIDURLRef must have a non-empty fragment: %q", s)
	}

	didURLRef, err := NewDIDURLRef(parsed[0], parsed[1])
	if err != nil {
		return nil, common.ErrValueInvalid.Wrap(err)
	}

	return didURLRef, nil
}

func (r DIDURLRef) String() string { return r.uri }
func (r DIDURLRef) DID() DIDRef    { return r.did }
func (r DIDURLRef) MethodSpecificID() string {
	return r.did.MethodSpecificID()
}
func (r DIDURLRef) Fragment() string { return r.fragment }

func (r DIDURLRef) IsValid(_ []byte) error {
	parsed := strings.Split(string(r.uri), "#")
	if len(parsed) != 2 {
		return errors.Errorf("DIDURLRef must have a fragment: %q", r.uri)
	}
	if parsed[1] == "" {
		return errors.Errorf("DIDURLRef must have a non-empty fragment: %q", r.uri)
	}
	return r.did.IsValid(nil)
}

func (r DIDURLRef) Bytes() []byte {
	return []byte(r.uri)
}

func ParseDIDScheme(did string) (method, methodSpecificID string, err error) {
	didStrings := strings.Split(did, DIDSeparator)
	if len(didStrings) != 3 {
		err = errors.Errorf("invalid DID scheme, %v", did)
		return
	}

	if didStrings[0] != DIDPrefix {
		err = errors.Errorf("invalid DID scheme, %v", did)
		return
	}

	method = didStrings[1]
	methodSpecificID = didStrings[2]
	return
}
