package cmds

import (
	"context"

	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	"github.com/ProtoconNet/mitum-currency/v3/types"

	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/pkg/errors"

	"github.com/ProtoconNet/mitum2/base"
)

type TransferCommand struct {
	BaseCommand
	OperationFlags
	Sender         AddressFlag               `arg:"" name:"sender" help:"sender address" required:"true"`
	ReceiverAmount AddressCurrencyAmountFlag `arg:"" name:"receiver-currency-amount" help:"receiver amount (ex: \"<address>,<currency>,<amount>\") separator @" required:"true"`
	OperationExtensionFlags
	sender base.Address
}

func (cmd *TransferCommand) Run(pctx context.Context) error {
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	encs = cmd.Encoders
	enc = cmd.Encoder

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

func (cmd *TransferCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	if sender, err := cmd.Sender.Encode(enc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %v", cmd.Sender.String())
	} else {
		cmd.sender = sender
	}

	cmd.OperationExtensionFlags.parseFlags(cmd.Encoders.JSON())

	return nil
}

func (cmd *TransferCommand) createOperation() (base.Operation, error) { // nolint:dupl
	var items []currency.TransferItem
	for i := range cmd.ReceiverAmount.Address() {
		item := currency.NewTransferItemMultiAmounts(cmd.ReceiverAmount.Address()[i], []types.Amount{cmd.ReceiverAmount.Amount()[i]})
		if err := item.IsValid(nil); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	fact := currency.NewTransferFact([]byte(cmd.Token), cmd.sender, items)

	op, err := currency.NewTransfer(fact)
	if err != nil {
		return nil, errors.Wrap(err, "create transfer operation")
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
		baseAuthentication = extras.NewBaseAuthentication(cmd.didContract, cmd.DID, cmd.AuthenticationID, proofData)
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
