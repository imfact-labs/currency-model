package types // nolint: dupl, revive

import (
	"bytes"
	"regexp"
	"sort"

	"github.com/imfact-labs/currency-model/common"
	"github.com/pkg/errors"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/valuehash"
)

type AccountStatus interface {
	util.IsValider
}

var ContractAccountStatusHint = hint.MustNewHint("mitum-currency-contract-account-status-v0.0.1")

const MaxHandlers = 20
const MaxRecipients = 20

type BalanceStatus uint8

const (
	Allowed = iota
	WithdrawalBlocked
	MaxBalanceStatus // last value is used for read length of BalanceStatus
)

func (bs BalanceStatus) IsValid([]byte) error {
	if uint(bs) > MaxBalanceStatus-1 {
		return common.ErrValueInvalid.Errorf("unexpected BalanceStatus, %v", bs)
	}

	return nil
}

func (bs BalanceStatus) Bytes() []byte {
	return util.Uint8ToBytes(uint8(bs))
}

type ContractAccountStatus struct {
	hint.BaseHinter
	owner             base.Address
	isActive          bool
	balanceStatus     BalanceStatus
	registerOperation *hint.Hint
	handlers          []base.Address
	recipients        []base.Address
}

func NewContractAccountStatus(owner base.Address, handlers []base.Address) ContractAccountStatus {
	sort.Slice(handlers, func(i, j int) bool {
		return bytes.Compare(handlers[i].Bytes(), handlers[j].Bytes()) < 0
	})

	us := ContractAccountStatus{
		BaseHinter:    hint.NewBaseHinter(ContractAccountStatusHint),
		owner:         owner,
		isActive:      false,
		balanceStatus: Allowed,
		handlers:      handlers,
	}
	return us
}

func (cs ContractAccountStatus) Bytes() []byte {
	var isActive int8
	if cs.isActive {
		isActive = 1
	}

	handlers := make([][]byte, len(cs.handlers))
	for i := range cs.handlers {
		handlers[i] = cs.handlers[i].Bytes()
	}
	recipients := make([][]byte, len(cs.recipients))
	for i := range cs.recipients {
		recipients[i] = cs.recipients[i].Bytes()
	}

	var h []byte
	if cs.registerOperation != nil {
		h = cs.registerOperation.Bytes()
	}

	return util.ConcatBytesSlice(
		cs.owner.Bytes(),
		[]byte{byte(isActive)},
		cs.balanceStatus.Bytes(),
		h,
		util.ConcatBytesSlice(handlers...),
		util.ConcatBytesSlice(recipients...),
	)
}

func (cs ContractAccountStatus) Hash() util.Hash {
	return cs.GenerateHash()
}

func (cs ContractAccountStatus) GenerateHash() util.Hash {
	return valuehash.NewSHA256(cs.Bytes())
}

func (cs ContractAccountStatus) IsValid([]byte) error { // nolint:revive
	if err := util.CheckIsValiders(nil, false,
		cs.BaseHinter,
		cs.owner,
		cs.balanceStatus,
	); err != nil {
		return err
	}

	if len(cs.handlers) > MaxHandlers {
		return common.ErrArrayLen.Wrap(
			errors.Errorf(
				"number of handlers, %d, exceeds maximum limit, %d", len(cs.handlers), MaxHandlers))
	}
	if len(cs.recipients) > MaxRecipients {
		return common.ErrArrayLen.Wrap(
			errors.Errorf(
				"number of recipients, %d, exceeds maximum limit, %d", len(cs.recipients), MaxRecipients))
	}

	return nil
}

func (cs ContractAccountStatus) Owner() base.Address { // nolint:revive
	return cs.owner
}

func (cs *ContractAccountStatus) SetOwner(a base.Address) error { // nolint:revive
	err := a.IsValid(nil)
	if err != nil {
		return err
	}

	cs.owner = a

	return nil
}

func (cs ContractAccountStatus) RegisterOperation() *hint.Hint {
	return cs.registerOperation
}

func (cs *ContractAccountStatus) SetRegisterOperation(h *hint.Hint) {
	cs.registerOperation = h
}

func (cs ContractAccountStatus) Handlers() []base.Address { // nolint:revive
	return cs.handlers
}

func (cs *ContractAccountStatus) SetHandlers(handlers []base.Address) error {
	sort.Slice(handlers, func(i, j int) bool {
		return bytes.Compare(handlers[i].Bytes(), handlers[j].Bytes()) < 0
	})

	for i := range handlers {
		err := handlers[i].IsValid(nil)
		if err != nil {
			return err
		}
	}

	cs.handlers = handlers

	return nil
}

func (cs ContractAccountStatus) IsHandler(ad base.Address) bool { // nolint:revive
	for i := range cs.Handlers() {
		if ad.Equal(cs.Handlers()[i]) {
			return true
		}
	}
	return false
}

func (cs ContractAccountStatus) Recipients() []base.Address { // nolint:revive
	return cs.recipients
}

func (cs *ContractAccountStatus) SetRecipients(recipients []base.Address) error {
	sort.Slice(recipients, func(i, j int) bool {
		return bytes.Compare(recipients[i].Bytes(), recipients[j].Bytes()) < 0
	})

	for i := range recipients {
		err := recipients[i].IsValid(nil)
		if err != nil {
			return err
		}
	}

	cs.recipients = recipients

	return nil
}

func (cs ContractAccountStatus) IsRecipients(ad base.Address) bool { // nolint:revive
	for i := range cs.Recipients() {
		if ad.Equal(cs.Recipients()[i]) {
			return true
		}
	}
	return false
}

func (cs ContractAccountStatus) IsActive() bool { // nolint:revive
	return cs.isActive
}

func (cs *ContractAccountStatus) SetActive(b bool) { // nolint:revive
	cs.isActive = b
}

func (cs ContractAccountStatus) BalanceStatus() BalanceStatus { // nolint:revive
	return cs.balanceStatus
}

func (cs *ContractAccountStatus) SetBalanceStatus(b BalanceStatus) { // nolint:revive
	cs.balanceStatus = b
}

func (cs ContractAccountStatus) Equal(b ContractAccountStatus) bool {
	if cs.isActive != b.isActive {
		return false
	} else if cs.balanceStatus != b.balanceStatus {
		return false
	} else if !cs.owner.Equal(b.owner) {
		return false
	}

	for i := range cs.handlers {
		if !cs.handlers[i].Equal(b.handlers[i]) {
			return false
		}
	}

	for i := range cs.recipients {
		if !cs.recipients[i].Equal(b.recipients[i]) {
			return false
		}
	}

	return true
}

var (
	MinLengthContractID = 3
	MaxLengthContractID = 50
	REContractIDExp     = regexp.MustCompile(`^[A-Z0-9][A-Z0-9-_\.\!\$\*\@]*[A-Z0-9]$`)
)

type ContractID string

func (cid ContractID) Bytes() []byte {
	return []byte(cid)
}

func (cid ContractID) String() string {
	return string(cid)
}

func (cid ContractID) IsValid([]byte) error {
	if l := len(cid); l < MinLengthContractID || l > MaxLengthContractID {
		return util.ErrInvalid.Errorf(
			"invalid length of contract id, %d <= %d <= %d", MinLengthContractID, l, MaxLengthContractID)
	}
	if !REContractIDExp.Match([]byte(cid)) {
		return util.ErrInvalid.Errorf("wrong contract id, %q", cid)
	}

	return nil
}
