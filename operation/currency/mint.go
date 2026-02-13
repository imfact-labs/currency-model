package currency

import (
	"github.com/imfact-labs/imfact-currency/common"
	"github.com/imfact-labs/imfact-currency/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var (
	MintFactHint = hint.MustNewHint("mitum-currency-mint-operation-fact-v0.0.1")
	MintHint     = hint.MustNewHint("mitum-currency-mint-operation-v0.0.1")
)

var maxMintItem = 10

type MintFact struct {
	base.BaseFact
	receiver base.Address
	amount   types.Amount
}

func NewMintFact(token []byte, receiver base.Address, amount types.Amount) MintFact {
	fact := MintFact{
		BaseFact: base.NewBaseFact(MintFactHint, token),
		receiver: receiver,
		amount:   amount,
	}

	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact MintFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact MintFact) Bytes() []byte {
	return util.ConcatBytesSlice(fact.Token(), fact.receiver.Bytes(), fact.amount.Bytes())
}

func (fact MintFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if err := util.CheckIsValiders(nil, false, fact.receiver, fact.amount); err != nil {
		return err
	}

	if !fact.amount.Big().OverZero() {
		return common.ErrValOOR.Wrap(errors.Errorf("Under zero amount of Mint"))
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact MintFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact MintFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact MintFact) Receiver() base.Address {
	return fact.receiver
}

func (fact MintFact) Currency() types.CurrencyID {
	return fact.amount.Currency()
}

func (fact MintFact) Amount() types.Amount {
	return fact.amount
}

type Mint struct {
	common.BaseNodeOperation
}

func NewMint(
	fact MintFact,
) (Mint, error) {
	return Mint{BaseNodeOperation: common.NewBaseNodeOperation(MintHint, fact)}, nil
}
