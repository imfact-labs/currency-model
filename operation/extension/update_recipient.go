package extension

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/operation/extras"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var (
	UpdateRecipientFactHint = hint.MustNewHint("mitum-extension-update-recipient-operation-fact-v0.0.1")
	UpdateRecipientHint     = hint.MustNewHint("mitum-extension-update-recipient-operation-v0.0.1")
)

type UpdateRecipientFact struct {
	base.BaseFact
	sender     base.Address
	contract   base.Address
	recipients []base.Address
	currency   types.CurrencyID
}

func NewUpdateRecipientFact(
	token []byte,
	sender,
	contract base.Address,
	recipients []base.Address,
	currency types.CurrencyID,
) UpdateRecipientFact {
	fact := UpdateRecipientFact{
		BaseFact:   base.NewBaseFact(UpdateRecipientFactHint, token),
		sender:     sender,
		contract:   contract,
		recipients: recipients,
		currency:   currency,
	}

	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact UpdateRecipientFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact UpdateRecipientFact) Bytes() []byte {
	bs := make([][]byte, len(fact.recipients)+4)
	bs[0] = fact.Token()
	bs[1] = fact.sender.Bytes()
	bs[2] = fact.contract.Bytes()
	bs[3] = fact.currency.Bytes()
	for i := range fact.recipients {
		bs[4+i] = fact.recipients[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (fact UpdateRecipientFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if err := util.CheckIsValiders(nil, false, fact.sender, fact.contract, fact.currency); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if len(fact.recipients) > types.MaxRecipients {
		return common.ErrFactInvalid.Wrap(
			common.ErrArrayLen.Wrap(
				errors.Errorf(
					"number of recipients, %d, exceeds maximum limit, %d", len(fact.recipients), types.MaxRecipients)))
	}

	recipientsMap := make(map[string]struct{})
	for i := range fact.recipients {
		_, found := recipientsMap[fact.recipients[i].String()]
		if found {
			return common.ErrFactInvalid.Wrap(
				common.ErrDupVal.Wrap(errors.Errorf("recipient %v", fact.recipients[i])))
		} else {
			recipientsMap[fact.recipients[i].String()] = struct{}{}
		}
		if err := fact.recipients[i].IsValid(nil); err != nil {
			return common.ErrFactInvalid.Wrap(
				common.ErrValueInvalid.Wrap(errors.Errorf("invalid recipient address %v", err)))
		}
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact UpdateRecipientFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact UpdateRecipientFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact UpdateRecipientFact) Currency() types.CurrencyID {
	return fact.currency
}

func (fact UpdateRecipientFact) Sender() base.Address {
	return fact.sender
}

func (fact UpdateRecipientFact) Signer() base.Address {
	return fact.sender
}

func (fact UpdateRecipientFact) Contract() base.Address {
	return fact.contract
}

func (fact UpdateRecipientFact) Recipients() []base.Address {
	return fact.recipients
}

func (fact UpdateRecipientFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, len(fact.recipients)+2)

	oprs := fact.recipients
	copy(as, oprs)

	as[len(fact.recipients)] = fact.sender
	as[len(fact.recipients)+1] = fact.contract

	return as, nil
}

func (fact UpdateRecipientFact) FeeBase() map[types.CurrencyID][]common.Big {
	required := make(map[types.CurrencyID][]common.Big)
	required[fact.Currency()] = []common.Big{common.ZeroBig}

	return required
}

func (fact UpdateRecipientFact) FeePayer() base.Address {
	return fact.sender
}

func (fact UpdateRecipientFact) FeeItemCount() (uint, bool) {
	return extras.ZeroItem, extras.HasNoItem
}

func (fact UpdateRecipientFact) FactUser() base.Address {
	return fact.sender
}

func (fact UpdateRecipientFact) ContractOwnerOnly() [][2]base.Address {
	return [][2]base.Address{{fact.contract, fact.sender}}
}

func (fact UpdateRecipientFact) DupKey() (map[types.DuplicationKeyType][]string, error) {
	r := make(map[types.DuplicationKeyType][]string)
	r[extras.DuplicationKeyTypeSender] = []string{fact.sender.String()}
	r[extras.DuplicationKeyTypeContractStatus] = []string{fact.Contract().String()}

	return r, nil
}

type UpdateRecipient struct {
	extras.ExtendedOperation
}

func NewUpdateRecipient(fact UpdateRecipientFact) (UpdateRecipient, error) {
	return UpdateRecipient{
		ExtendedOperation: extras.NewExtendedOperation(UpdateRecipientHint, fact),
	}, nil
}
