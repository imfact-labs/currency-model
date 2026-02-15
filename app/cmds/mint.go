package cmds

import (
	"context"

	"github.com/imfact-labs/imfact-currency/operation/currency"
	"github.com/imfact-labs/imfact-currency/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

type MintCommand struct {
	BaseCommand
	OperationFlags
	Node     AddressFlag `arg:"" name:"node" help:"node address" required:"true"`
	node     base.Address
	Receiver AddressFlag `arg:"" name:"receiver" help:"receiver address" required:"true"`
	receiver base.Address
	Amount   CurrencyAmountFlag `arg:"" name:"currency-amount" help:"amount (ex: \"<currency>,<amount>\")"`
}

func (cmd *MintCommand) Run(pctx context.Context) error { // nolint:dupl
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	var op base.Operation
	if i, err := cmd.createOperation(); err != nil {
		return errors.Wrap(err, "create mint operation")
	} else if err := i.IsValid(cmd.OperationFlags.NetworkID); err != nil {
		return errors.Wrap(err, "invalid mint operation")
	} else {
		cmd.Log.Debug().Interface("operation", i).Msg("operation loaded")

		op = i
	}

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *MintCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Node.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid node format, %v", cmd.Node.String())
	}
	cmd.node = a

	r, err := cmd.Receiver.Encode(cmd.Encoders.JSON())
	if err != nil {
		return errors.Wrapf(err, "invalid sender format, %v", cmd.Receiver.String())
	}
	cmd.receiver = r

	return nil
}

func (cmd *MintCommand) createOperation() (currency.Mint, error) {
	am := types.NewAmount(cmd.Amount.Big, cmd.Amount.CID)

	fact := currency.NewMintFact([]byte(cmd.Token), cmd.receiver, am)
	cmd.Log.Debug().
		Stringer("receiver", cmd.receiver).
		Stringer("amount", am).
		Msg("mint fact loaded")

	op, err := currency.NewMint(fact)
	if err != nil {
		return currency.Mint{}, err
	}

	err = op.NodeSign(cmd.Privatekey, cmd.NetworkID.NetworkID(), cmd.node)
	if err != nil {
		return currency.Mint{}, errors.Wrap(err, "create mint operation")
	}

	return op, nil
}
