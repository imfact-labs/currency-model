package cmds

import (
	"context"

	did "github.com/imfact-labs/currency-model/operation/did-registry"
	"github.com/imfact-labs/currency-model/operation/extras"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/pkg/errors"
)

type RegisterModelCommand struct {
	BaseCommand
	OperationFlags
	Sender    AddressFlag    `arg:"" name:"sender" help:"sender address" required:"true"`
	Contract  AddressFlag    `arg:"" name:"contract" help:"contract account to register policy" required:"true"`
	DIDMethod string         `arg:"" name:"did-method" help:"did method" required:"true"`
	Currency  CurrencyIDFlag `arg:"" name:"currency" help:"currency id" required:"true"`
	OperationExtensionFlags
	sender   base.Address
	contract base.Address
}

func (cmd *RegisterModelCommand) Run(pctx context.Context) error {
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

func (cmd *RegisterModelCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	if a, err := cmd.Sender.Encode(cmd.Encoders.JSON()); err != nil {
		return errors.Wrapf(err, "invalid sender format; %q", cmd.Sender)
	} else {
		cmd.sender = a
	}

	if a, err := cmd.Contract.Encode(cmd.Encoders.JSON()); err != nil {
		return errors.Wrapf(err, "invalid contract format; %q", cmd.Contract)
	} else {
		cmd.contract = a
	}

	if len(cmd.DIDMethod) < 1 {
		return errors.Errorf("invalid DID Method, %s", cmd.DIDMethod)
	}

	err := cmd.OperationExtensionFlags.parseFlags(cmd.Encoders.JSON())
	if err != nil {
		return err
	}

	return nil
}

func (cmd *RegisterModelCommand) createOperation() (base.Operation, error) {
	e := util.StringError("failed to create register-model operation")

	fact := did.NewRegisterModelFact(
		[]byte(cmd.Token), cmd.sender, cmd.contract, cmd.DIDMethod, cmd.Currency.CID,
	)

	op, err := did.NewRegisterModel(fact)
	if err != nil {
		return nil, e.Wrap(err)
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

		err = op.Sign(cmd.OpSenderPrivatekey, cmd.NetworkID.NetworkID())
		if err != nil {
			return nil, errors.Wrapf(err, "create %T operation", op)
		}
	} else {
		err = op.Sign(cmd.Privatekey, cmd.NetworkID.NetworkID())
		if err != nil {
			return nil, errors.Wrapf(err, "create %T operation", op)
		}
	}

	if err := op.IsValid(cmd.OperationFlags.NetworkID); err != nil {
		return nil, errors.Wrapf(err, "create %T operation", op)
	}

	return op, nil
}
