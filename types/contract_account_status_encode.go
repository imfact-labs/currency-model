package types // nolint: dupl, revive

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (cs *ContractAccountStatus) unpack(
	enc encoder.Encoder,
	ht hint.Hint,
	ow string,
	ia bool,
	bs uint8,
	rht *hint.Hint,
	hds, rcps []string,
) error {
	cs.BaseHinter = hint.NewBaseHinter(ht)
	cs.registerOperation = rht

	switch a, err := base.DecodeAddress(ow, enc); {
	case err != nil:
		return errors.Errorf("Decode address, %v", err)
	default:
		cs.owner = a
	}

	cs.isActive = ia
	balanceStatus := BalanceStatus(bs)
	if err := balanceStatus.IsValid(nil); err != nil {
		return errors.Errorf("Invalid balancestatus value, %v", bs)
	}
	cs.balanceStatus = balanceStatus

	handlers := make([]base.Address, len(hds))
	for i, hd := range hds {
		switch handler, err := base.DecodeAddress(hd, enc); {
		case err != nil:
			return err
		default:
			handlers[i] = handler
		}
	}
	cs.handlers = handlers

	recipients := make([]base.Address, len(rcps))
	for i, rcp := range rcps {
		switch recipient, err := base.DecodeAddress(rcp, enc); {
		case err != nil:
			return err
		default:
			recipients[i] = recipient
		}
	}
	cs.recipients = recipients

	return nil
}
