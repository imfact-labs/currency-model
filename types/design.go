package types

import (
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/valuehash"
)

var DesignHint = hint.MustNewHint("mitum-did-design-v0.0.1")

type Design struct {
	hint.BaseHinter
	didMethod string
}

func NewDesign(didMethod string) Design {
	return Design{
		BaseHinter: hint.NewBaseHinter(DesignHint),
		didMethod:  didMethod,
	}
}

func (de Design) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false,
		de.BaseHinter,
	); err != nil {
		return err
	}

	return nil
}

func (de Design) Bytes() []byte {
	return util.ConcatBytesSlice(
		[]byte(de.didMethod),
	)
}

func (de Design) Hash() util.Hash {
	return de.GenerateHash()
}

func (de Design) GenerateHash() util.Hash {
	return valuehash.NewSHA256(de.Bytes())
}

func (de Design) DIDMethod() string {
	return de.didMethod
}

func (de Design) Equal(cd Design) bool {
	if de.didMethod != cd.didMethod {
		return false
	}

	return true
}
