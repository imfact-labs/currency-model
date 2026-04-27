package did_registry

import (
	"fmt"

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
	CreateDIDFactHint = hint.MustNewHint("mitum-did-create-did-operation-fact-v0.0.1")
	CreateDIDHint     = hint.MustNewHint("mitum-did-create-did-operation-v0.0.1")
)

type CreateDIDFact struct {
	base.BaseFact
	sender   base.Address
	contract base.Address
	currency types.CurrencyID
}

func NewCreateDIDFact(
	token []byte, sender, contract base.Address,
	currency types.CurrencyID) CreateDIDFact {
	bf := base.NewBaseFact(CreateDIDFactHint, token)
	fact := CreateDIDFact{
		BaseFact: bf,
		sender:   sender,
		contract: contract,
		currency: currency,
	}

	fact.SetHash(fact.GenerateHash())
	return fact
}

func (fact CreateDIDFact) IsValid(b []byte) error {
	if fact.sender.Equal(fact.contract) {
		return common.ErrFactInvalid.Wrap(
			common.ErrSelfTarget.Wrap(errors.Errorf("sender %v is same with contract account", fact.sender)))
	}

	if err := util.CheckIsValiders(nil, false,
		fact.BaseHinter,
		fact.sender,
		fact.contract,
		fact.currency,
	); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact CreateDIDFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact CreateDIDFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact CreateDIDFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		fact.currency.Bytes(),
	)
}

func (fact CreateDIDFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact CreateDIDFact) Sender() base.Address {
	return fact.sender
}

func (fact CreateDIDFact) Signer() base.Address {
	return fact.sender
}

func (fact CreateDIDFact) Contract() base.Address {
	return fact.contract
}

func (fact CreateDIDFact) Currency() types.CurrencyID {
	return fact.currency
}

func (fact CreateDIDFact) Addresses() ([]base.Address, error) {
	as := []base.Address{fact.sender}

	return as, nil
}

func (fact CreateDIDFact) FeeBase() (types.CurrencyID, int, int, bool) {
	return fact.Currency(), extras.NoItemFeeBaseItemCount, len(fact.Bytes()), extras.HasNoItem
}

func (fact CreateDIDFact) FeePayer() base.Address {
	return fact.sender
}

func (fact CreateDIDFact) FactUser() base.Address {
	return fact.sender
}

func (fact CreateDIDFact) ActiveContract() []base.Address {
	return []base.Address{fact.contract}
}

func (fact CreateDIDFact) DupKey() (map[types.DuplicationKeyType][]string, error) {
	r := make(map[types.DuplicationKeyType][]string)
	r[extras.DuplicationKeyTypeDIDAccount] = []string{fmt.Sprintf("%s:%s", fact.Contract().String(), fact.Sender())}

	return r, nil
}

type CreateDID struct {
	extras.ExtendedOperation
}

func (op CreateDID) DupKey() (map[types.DuplicationKeyType][]string, error) {
	r := make(map[types.DuplicationKeyType][]string)

	var payer base.Address
	payer, err := extras.FetchPayerHelper(op)
	if err != nil {
		return nil, err
	}

	r[extras.DuplicationKeyTypeSender] = []string{payer.String()}
	return r, nil
}

func NewCreateDID(fact CreateDIDFact) (CreateDID, error) {
	return CreateDID{
		ExtendedOperation: extras.NewExtendedOperation(CreateDIDHint, fact),
	}, nil
}
