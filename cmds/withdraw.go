package cmds

import (
	"context"

	"github.com/ProtoconNet/mitum-currency/v3/operation/extension"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	"github.com/ProtoconNet/mitum-currency/v3/types"

	"github.com/pkg/errors"

	"github.com/ProtoconNet/mitum2/base"
)

type WithdrawCommand struct {
	BaseCommand
	OperationFlags
	Sender AddressFlag        `arg:"" name:"sender" help:"sender address" required:"true"`
	Target AddressFlag        `arg:"" name:"target" help:"target contract account address" required:"true"`
	Amount CurrencyAmountFlag `arg:"" name:"currency-amount" help:"amount (ex: \"<currency>,<amount>\")"`
	OperationExtensionFlags
	sender      base.Address
	target      base.Address
	didContract base.Address
	proxyPayer  base.Address
	opSender    base.Address
}

func (cmd *WithdrawCommand) Run(pctx context.Context) error {
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	op, err := cmd.createOperation()
	if err != nil {
		return err
	}

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *WithdrawCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	if sender, err := cmd.Sender.Encode(enc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %v", cmd.Sender.String())
	} else if target, err := cmd.Target.Encode(enc); err != nil {
		return errors.Wrapf(err, "invalid target format, %v", cmd.Target.String())
	} else {
		cmd.sender = sender
		cmd.target = target
	}

	err := cmd.OperationExtensionFlags.parseFlags(cmd.Encoders.JSON())
	if err != nil {
		return err
	}

	return nil
}

func (cmd *WithdrawCommand) createOperation() (base.Operation, error) { // nolint:dupl
	var items []extension.WithdrawItem

	ams := make([]types.Amount, 1)
	am := types.NewAmount(cmd.Amount.Big, cmd.Amount.CID)
	if err := am.IsValid(nil); err != nil {
		return nil, err
	}

	ams[0] = am

	item := extension.NewWithdrawItemMultiAmounts(cmd.target, ams)
	if err := item.IsValid(nil); err != nil {
		return nil, err
	}
	items = append(items, item)

	fact := extension.NewWithdrawFact([]byte(cmd.Token), cmd.sender, items)

	op, err := extension.NewWithdraw(fact)
	if err != nil {
		return nil, errors.Wrap(err, "create withdraw operation")
	}

	var baseAuthentication extras.OperationExtension
	var baseSettlement extras.OperationExtension
	var baseProxyPayer extras.OperationExtension
	var proofData = cmd.Proof
	if cmd.IsPrivateKey {
		prk, err := base.DecodePrivatekeyFromString(cmd.Proof, enc)
		if err != nil {
			return nil, err
		}

		sig, err := prk.Sign(fact.Hash().Bytes())
		if err != nil {
			return nil, err
		}
		proofData = sig.String()
	}

	if cmd.didContract != nil && cmd.AuthenticationID != "" && cmd.Proof != "" {
		baseAuthentication = extras.NewBaseAuthentication(cmd.didContract, cmd.AuthenticationID, proofData)
		if err := op.AddExtension(baseAuthentication); err != nil {
			return nil, err
		}
	}

	if cmd.proxyPayer != nil {
		baseProxyPayer = extras.NewBaseProxyPayer(cmd.proxyPayer)
		if err := op.AddExtension(baseProxyPayer); err != nil {
			return nil, err
		}
	}

	if cmd.opSender != nil {
		baseSettlement = extras.NewBaseSettlement(cmd.opSender)
		if err := op.AddExtension(baseSettlement); err != nil {
			return nil, err
		}

		err = op.HashSign(cmd.OpSenderPrivatekey, cmd.NetworkID.NetworkID())
		if err != nil {
			return nil, errors.Wrapf(err, "create %T operation", op)
		}
	} else {
		err = op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
		if err != nil {
			return nil, errors.Wrapf(err, "create %T operation", op)
		}
	}

	if err := op.IsValid(cmd.OperationFlags.NetworkID); err != nil {
		return nil, errors.Wrapf(err, "create %T operation", op)
	}

	return op, nil
}
