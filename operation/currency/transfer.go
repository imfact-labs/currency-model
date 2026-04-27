package currency

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
	TransferFactHint = hint.MustNewHint("mitum-currency-transfer-operation-fact-v0.0.1")
	TransferHint     = hint.MustNewHint("mitum-currency-transfer-operation-v0.0.1")
)

var MaxTransferItems uint = 3000

type TransferItem interface {
	hint.Hinter
	util.IsValider
	AmountsItem
	util.Byter
	Receiver() base.Address
	Rebuild() TransferItem
}

type TransferFact struct {
	base.BaseFact
	sender   base.Address
	items    []TransferItem
	currency types.CurrencyID
}

func NewTransferFact(
	token []byte,
	sender base.Address,
	items []TransferItem,
	currency types.CurrencyID,
) TransferFact {
	bf := base.NewBaseFact(TransferFactHint, token)
	fact := TransferFact{
		BaseFact: bf,
		sender:   sender,
		items:    items,
		currency: currency,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact TransferFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact TransferFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact TransferFact) Bytes() []byte {
	its := make([][]byte, len(fact.items))
	for i := range fact.items {
		its[i] = fact.items[i].Bytes()
	}

	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.currency.Bytes(),
		util.ConcatBytesSlice(its...),
	)
}

func (fact TransferFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if n := len(fact.items); n < 1 {
		return common.ErrFactInvalid.Wrap(common.ErrArrayLen.Wrap(errors.Errorf("empty items")))
	} else if n > int(MaxTransferItems) {
		return common.ErrFactInvalid.Wrap(common.ErrArrayLen.Wrap(errors.Errorf("items, %d over max, %d", n, MaxTransferItems)))
	}

	if err := util.CheckIsValiders(nil, false, fact.sender, fact.currency); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	foundReceivers := map[string]struct{}{}
	for i := range fact.items {
		it := fact.items[i]
		if err := util.CheckIsValiders(nil, false, it); err != nil {
			return common.ErrFactInvalid.Wrap(err)
		}

		k := it.Receiver().String()
		switch _, found := foundReceivers[k]; {
		case found:
			return common.ErrFactInvalid.Wrap(common.ErrDupVal.Wrap(errors.Errorf("receiver found, %v", it.Receiver())))
		case fact.sender.Equal(it.Receiver()):
			return common.ErrFactInvalid.Wrap(common.ErrSelfTarget.Wrap(errors.Errorf("receiver account is same with sender account, %v", fact.sender)))
		default:
			foundReceivers[k] = struct{}{}
		}
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact TransferFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact TransferFact) Sender() base.Address {
	return fact.sender
}

func (fact TransferFact) Signer() base.Address {
	return fact.sender
}

func (fact TransferFact) Items() []TransferItem {
	return fact.items
}

func (fact TransferFact) ItemsLen() int {
	return len(fact.items)
}

func (fact TransferFact) Currency() types.CurrencyID {
	return fact.currency
}

func (fact TransferFact) Rebuild() TransferFact {
	items := make([]TransferItem, len(fact.items))
	for i := range fact.items {
		it := fact.items[i]
		items[i] = it.Rebuild()
	}

	fact.items = items
	fact.SetHash(fact.Hash())

	return fact
}

func (fact TransferFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, len(fact.items)+1)
	for i := range fact.items {
		as[i] = fact.items[i].Receiver()
	}

	as[len(fact.items)] = fact.Sender()

	return as, nil
}

func (fact TransferFact) FeeBase() (types.CurrencyID, int, int, bool) {
	return fact.Currency(), len(fact.items), len(fact.Bytes()), extras.HasItem
}

func (fact TransferFact) FeePayer() base.Address {
	return fact.sender
}

func (fact TransferFact) FactUser() base.Address {
	return fact.sender
}

func (fact TransferFact) DupKey() (map[types.DuplicationKeyType][]string, error) {
	r := make(map[types.DuplicationKeyType][]string)
	r[extras.DuplicationKeyTypeSender] = []string{fmt.Sprintf("%s:%s", fact.sender.String(), fact.currency.String())}

	return r, nil
}

type Transfer struct {
	extras.ExtendedOperation
}

func (op Transfer) DupKey() (map[types.DuplicationKeyType][]string, error) {
	r := make(map[types.DuplicationKeyType][]string)

	if err := extras.AddOperationFeePayerDupKeys(r, op); err != nil {
		return nil, err
	}

	return r, nil
}

func NewTransfer(fact base.Fact) (Transfer, error) {
	return Transfer{
		ExtendedOperation: extras.NewExtendedOperation(TransferHint, fact),
	}, nil
}
