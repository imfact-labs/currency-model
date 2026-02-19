package cmds

import (
	"context"

	"github.com/imfact-labs/currency-model/operation/extension"
	"github.com/imfact-labs/currency-model/operation/extras"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/pkg/errors"
)

type CreateContractAccountCommand struct {
	BaseCommand
	OperationFlags
	Sender    AddressFlag        `arg:"" name:"sender" help:"sender address" required:"true"`
	Threshold uint               `help:"threshold for keys (default: ${create_contract_account_threshold})" default:"${create_contract_account_threshold}"` // nolint
	Key       KeyFlag            `name:"key" help:"key for new account (ex: \"<public key>,<weight>\") separator @"`
	Amount    CurrencyAmountFlag `arg:"" name:"currency-amount" help:"amount (ex: \"<currency>,<amount>\")"`
	OperationExtensionFlags
	sender base.Address
	keys   types.AccountKeys
}

func (cmd *CreateContractAccountCommand) Run(pctx context.Context) error { // nolint:dupl
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	//encs = cmd.Encoders
	//enc = cmd.Encoder

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

func (cmd *CreateContractAccountCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Sender.Encode(cmd.Encoder)
	if err != nil {
		return errors.Wrapf(err, "invalid sender format, %v", cmd.Sender.String())
	}
	cmd.sender = a

	{
		ks := make([]types.AccountKey, len(cmd.Key.Values))
		for i := range cmd.Key.Values {
			ks[i] = cmd.Key.Values[i]
		}

		var kys types.AccountKeys
		if kys, err = types.NewBaseAccountKeys(ks, cmd.Threshold); err != nil {
			return err
		}

		if err := kys.IsValid(nil); err != nil {
			return err
		} else {
			cmd.keys = kys
		}
	}

	err = cmd.OperationExtensionFlags.parseFlags(cmd.Encoders.JSON())
	if err != nil {
		return err
	}

	return nil
}

func (cmd *CreateContractAccountCommand) createOperation() (base.Operation, error) { // nolint:dupl}
	var items []extension.CreateContractAccountItem

	ams := make([]types.Amount, 1)
	am := types.NewAmount(cmd.Amount.Big, cmd.Amount.CID)
	if err := am.IsValid(nil); err != nil {
		return nil, err
	}

	ams[0] = am

	item := extension.NewCreateContractAccountItemMultiAmounts(cmd.keys, ams)
	if err := item.IsValid(nil); err != nil {
		return nil, err
	}
	items = append(items, item)

	fact := extension.NewCreateContractAccountFact([]byte(cmd.Token), cmd.sender, items)

	op, err := extension.NewCreateContractAccount(fact)
	if err != nil {
		return nil, errors.Wrap(err, "create create-contract-account operation")
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
