package types

import (
	"encoding/hex"
	"strings"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

var (
	AddressHint       = hint.MustNewHint("fca-v0.0.1")
	ZeroAddressSuffix = "-X"
)

const (
	AddressLength = 20
)

type Address struct {
	base.BaseStringAddress
}

func NewAddress(s string) Address {
	ca := Address{BaseStringAddress: base.NewBaseStringAddressWithHint(AddressHint, s)}

	return ca
}

func NewAddressFromString(s string) (Address, error) {
	sad, _, err := hint.ParseFixedTypedString(s, base.AddressTypeSize)
	if err != nil {
		return Address{}, util.ErrInvalid.Errorf("invalid mitum currency address: %v", err)
	}
	adr := NewAddress(sad)
	err = adr.IsValid(nil)
	if err != nil {
		return Address{}, util.ErrInvalid.Errorf("invalid mitum currency address: %v", err)
	}
	return adr, nil
}

func NewAddressFromKeys(keys AccountKeys) (Address, error) {
	if keys == nil {
		return Address{}, errors.Errorf("nil account keys")
	}

	var buf [42]byte
	copy(buf[:2], "0x")
	hex.Encode(buf[2:], keys.Hash().Bytes())
	s := string(ChecksumHex(buf))

	return NewAddress(s), nil
}

func (ad Address) IsValid([]byte) error {
	if err := ad.BaseStringAddress.IsValid(nil); err != nil {
		return util.ErrInvalid.Errorf("invalid mitum currency address: %v", err)
	}

	sad, _, err := hint.ParseFixedTypedString(ad.String(), 3)
	if err != nil {
		return util.ErrInvalid.Errorf("invalid mitum currency address: %v", err)
	}

	switch {
	case IsZeroAddress(sad):
		return nil
	default:
		var buf [42]byte

		copy(buf[:2], "0x")
		lowered := strings.ToLower(strings.TrimPrefix(sad, "0x"))

		bytes, err := hex.DecodeString(lowered)
		if err != nil {
			return util.ErrInvalid.Errorf("invalid mitum currency address: %v", err)
		}
		hex.Encode(buf[2:], bytes)
		if string(ChecksumHex(buf)) != sad {
			return util.ErrInvalid.Errorf("invalid mitum currency address: checksum not matched, expeced %v but %v", string(ChecksumHex(buf)), sad)
		}
	}

	return nil
}

// ChecksumHex return the hex in the manner of EIP55
func ChecksumHex(buf [42]byte) []byte {
	// compute checksum
	sha := sha3.NewLegacyKeccak256()
	sha.Write(buf[2:])
	hash := sha.Sum(nil)
	for i := 2; i < len(buf); i++ {
		hashByte := hash[(i-2)/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if buf[i] > '9' && hashByte > 7 {
			buf[i] -= 32
		}
	}
	return buf[:]
}

type Addresses interface {
	Addresses() ([]base.Address, error)
}

func ZeroAddress(cid CurrencyID) Address {
	return NewAddress(cid.String() + ZeroAddressSuffix)
}

func IsZeroAddress(ad string) bool {
	return strings.HasSuffix(ad, ZeroAddressSuffix)
}
