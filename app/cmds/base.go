package cmds

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/imfact-labs/mitum2/base"
	"github.com/pkg/errors"

	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/localtime"
	"github.com/imfact-labs/mitum2/util/logging"
	"github.com/imfact-labs/mitum2/util/ps"
	"github.com/rs/zerolog"
)

type BaseCommand struct {
	Encoder  encoder.Encoder   `kong:"-"`
	Encoders *encoder.Encoders `kong:"-"`
	Log      *zerolog.Logger   `kong:"-"`
	Out      io.Writer         `kong:"-"`
}

func (cmd *BaseCommand) prepare(pctx context.Context) (context.Context, error) {
	cmd.Out = os.Stdout
	pps := ps.NewPS("cmd")

	_ = pps.
		AddOK(launch.PNameEncoder, PEncoder, nil)

	_ = pps.POK(launch.PNameEncoder).
		PostAddOK(launch.PNameAddHinters, PAddHinters)

	var log *logging.Logging
	if err := util.LoadFromContextOK(pctx, launch.LoggingContextKey, &log); err != nil {
		return pctx, err
	}

	cmd.Log = log.Log()

	pctx, err := pps.Run(pctx) //revive:disable-line:modifies-parameter
	if err != nil {
		return pctx, err
	}

	if err := util.LoadFromContextOK(pctx, launch.EncodersContextKey, &cmd.Encoders); err != nil {
		return pctx, err
	}

	cmd.Encoder = cmd.Encoders.JSON()

	return pctx, nil
}

func (cmd *BaseCommand) print(f string, a ...interface{}) {
	_, _ = fmt.Fprintf(cmd.Out, f, a...)
	_, _ = fmt.Fprintln(cmd.Out)
}

func PAddHinters(pctx context.Context) (context.Context, error) {
	e := util.StringError("add hinters")

	var encs *encoder.Encoders
	var f ProposalOperationFactHintFunc = IsSupportedProposalOperationFactHintFunc

	if err := util.LoadFromContextOK(pctx, launch.EncodersContextKey, &encs); err != nil {
		return pctx, e.Wrap(err)
	}
	pctx = context.WithValue(pctx, ProposalOperationFactHintContextKey, f)

	if err := LoadHinters(encs); err != nil {
		return pctx, e.Wrap(err)
	}

	return pctx, nil
}

type OperationFlags struct {
	Privatekey PrivatekeyFlag `arg:"" name:"privatekey" help:"privatekey to sign operation" required:"true"`
	Token      string         `help:"token for operation" optional:""`
	NetworkID  NetworkIDFlag  `name:"network-id" help:"network-id" required:"true" default:"${network_id}"`
	Pretty     bool           `name:"pretty" help:"pretty format"`
}

func (op *OperationFlags) IsValid([]byte) error {
	if len(op.Token) < 1 {
		op.Token = localtime.Now().UTC().String()
	}

	return op.NetworkID.NetworkID().IsValid(nil)
}

type OperationExtensionFlags struct {
	DIDContract        AddressFlag    `name:"authentication-contract" help:"contract account for authentication"`
	AuthenticationID   string         `name:"authentication-id" help:"auth id for authentication"`
	Proof              string         `name:"authentication-proof" help:"data for proof authentication"`
	IsPrivateKey       bool           `name:"is-privatekey" help:"proor-data is private key, not signature"`
	OpSender           AddressFlag    `name:"settlement-op-sender" help:"op sender account for settlement"`
	OpSenderPrivatekey PrivatekeyFlag `name:"settlement-op-sender-privatekey" help:"op sender privatekey for settlement"`
	ProxyPayer         AddressFlag    `name:"settlement-proxy-payer" help:"proxy payer account for settlement"`
	didContract        base.Address
	proxyPayer         base.Address
	opSender           base.Address
}

func (op *OperationExtensionFlags) parseFlags(encoder encoder.Encoder) error {
	if len(op.DIDContract.String()) > 0 {
		a, err := op.DIDContract.Encode(encoder)
		if err != nil {
			return errors.Wrapf(err, "invalid did contract format, %v", op.DIDContract.String())
		}
		op.didContract = a
	}

	if len(op.OpSender.String()) > 0 {
		a, err := op.OpSender.Encode(encoder)
		if err != nil {
			return errors.Wrapf(err, "invalid proxy payer format, %v", op.ProxyPayer.String())
		}
		op.opSender = a
	}

	if len(op.ProxyPayer.String()) > 0 {
		a, err := op.ProxyPayer.Encode(encoder)
		if err != nil {
			return errors.Wrapf(err, "invalid proxy payer format, %v", op.ProxyPayer.String())
		}
		op.proxyPayer = a
	}

	return nil
}
