package currency

import (
	"github.com/imfact-labs/imfact-currency/common"
	"github.com/imfact-labs/imfact-currency/operation/extras"
	"github.com/imfact-labs/imfact-currency/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var (
	UpdateKeyFactHint = hint.MustNewHint("mitum-currency-update-key-operation-fact-v0.0.1")
	UpdateKeyHint     = hint.MustNewHint("mitum-currency-update-key-operation-v0.0.1")
)

type UpdateKeyFact struct {
	base.BaseFact
	sender   base.Address
	keys     types.AccountKeys
	currency types.CurrencyID
}

func NewUpdateKeyFact(
	token []byte,
	sender base.Address,
	keys types.AccountKeys,
	currency types.CurrencyID,
) UpdateKeyFact {
	bf := base.NewBaseFact(UpdateKeyFactHint, token)
	fact := UpdateKeyFact{
		BaseFact: bf,
		sender:   sender,
		keys:     keys,
		currency: currency,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact UpdateKeyFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact UpdateKeyFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact UpdateKeyFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.keys.Bytes(),
		fact.currency.Bytes(),
	)
}

func (fact UpdateKeyFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if err := util.CheckIsValiders(nil, false, fact.sender, fact.keys, fact.currency); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact UpdateKeyFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact UpdateKeyFact) Sender() base.Address {
	return fact.sender
}

func (fact UpdateKeyFact) Signer() base.Address {
	return fact.sender
}

func (fact UpdateKeyFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, 1)
	as[0] = fact.Sender()
	return as, nil
}

func (fact UpdateKeyFact) Keys() types.AccountKeys {
	return fact.keys
}

func (fact UpdateKeyFact) Currency() types.CurrencyID {
	return fact.currency
}

func (fact UpdateKeyFact) Rebuild() UpdateKeyFact {
	fact.SetHash(fact.Hash())
	return fact
}

func (fact UpdateKeyFact) FeeBase() map[types.CurrencyID][]common.Big {
	required := make(map[types.CurrencyID][]common.Big)
	required[fact.Currency()] = []common.Big{common.ZeroBig}

	return required
}

func (fact UpdateKeyFact) FeePayer() base.Address {
	return fact.sender
}

func (fact UpdateKeyFact) FeeItemCount() (uint, bool) {
	return extras.ZeroItem, extras.HasNoItem
}

func (fact UpdateKeyFact) FactUser() base.Address {
	return fact.sender
}

func (fact UpdateKeyFact) DupKey() (map[types.DuplicationKeyType][]string, error) {
	r := make(map[types.DuplicationKeyType][]string)
	r[extras.DuplicationKeyTypeSender] = []string{fact.sender.String()}

	return r, nil
}

type UpdateKey struct {
	extras.ExtendedOperation
}

func NewUpdateKey(fact UpdateKeyFact) (UpdateKey, error) {
	return UpdateKey{
		ExtendedOperation: extras.NewExtendedOperation(UpdateKeyHint, fact),
	}, nil
}
