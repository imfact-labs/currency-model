package cmds

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/imfact-labs/mitum2/base"
	isaacnetwork "github.com/imfact-labs/mitum2/isaac/network"
	"github.com/imfact-labs/mitum2/launch"
	quicstreamheader "github.com/imfact-labs/mitum2/network/quicstream/header"
	"github.com/imfact-labs/mitum2/util"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

type baseNetworkClientRWNodeCommand struct { //nolint:govet //...
	BaseNetworkClientCommand
	Privatekey string `arg:"" name:"privatekey" help:"privatekey string"`
	Key        string `arg:"" name:"key" help:"key"`
	Format     string `name:"format" help:"output format, {json, yaml}" default:"yaml"`
	priv       base.Privatekey
}

func (cmd *baseNetworkClientRWNodeCommand) Prepare(pctx context.Context) error {
	if err := cmd.BaseNetworkClientCommand.Prepare(pctx); err != nil {
		return err
	}

	if len(cmd.Key) < 1 {
		return errors.Errorf("empty key")
	}

	switch cmd.Format {
	case "json", "yaml":
	default:
		return errors.Errorf("unsupported format, %q", cmd.Format)
	}

	switch key, err := launch.DecodePrivatekey(cmd.Privatekey, cmd.Encoders.JSON()); {
	case err != nil:
		return err
	default:
		cmd.priv = key
	}

	return nil
}

func (cmd *baseNetworkClientRWNodeCommand) printValue(ctx context.Context, key string) error {
	stream, _, err := cmd.Client.Dial(ctx, cmd.Remote.ConnInfo())
	if err != nil {
		return err
	}

	switch i, found, err := ReadNodeFromNetworkHandler(
		ctx,
		cmd.priv,
		base.NetworkID(cmd.NetworkID),
		key,
		stream,
	); {
	case err != nil:
		return err
	case !found:
		return util.ErrNotFound.Errorf("unknown key, %q", key)
	case cmd.Format == "json":
		b, err := util.MarshalJSON(i)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(os.Stdout, string(b))
	case cmd.Format == "yaml":
		var v string

		if i != nil {
			b, err := yaml.Marshal(i)
			if err != nil {
				return errors.WithStack(err)
			}

			v = string(b)
		}

		_, _ = fmt.Fprintln(os.Stdout, v)
	}

	return nil
}

type NetworkClientReadNodeCommand struct { //nolint:govet //...
	baseNetworkClientRWNodeCommand
}

func (cmd *NetworkClientReadNodeCommand) Run(pctx context.Context) error {
	if err := cmd.Prepare(pctx); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(pctx, cmd.Timeout)
	defer cancel()

	defer func() {
		_ = cmd.Client.Close()
	}()

	return cmd.printValue(ctx, cmd.Key)
}

func (*NetworkClientReadNodeCommand) Help() string {
	s := "## available keys\n\n"

	for i := range launch.AllNodeReadKeys {
		s += fmt.Sprintf("  - %s\n", launch.AllNodeReadKeys[i])
	}

	return s
}

type NetworkClientWriteNodeCommand struct { //nolint:govet //...
	baseNetworkClientRWNodeCommand
	Input  string `arg:"" name:"input" help:"input"`
	IsFile bool   `name:"input.is-file" help:"input is file"`
}

func (cmd *NetworkClientWriteNodeCommand) Run(pctx context.Context) error {
	if err := cmd.Prepare(pctx); err != nil {
		return err
	}

	var input string

	switch i, err := launch.LoadInputFlag(cmd.Input, cmd.IsFile); {
	case err != nil:
		return err
	case len(i) < 1:
		return errors.Errorf("empty body")
	default:
		input = string(i)
	}

	cmd.Log.Debug().
		Str("key", cmd.Key).
		Str("input", input).
		Msg("flags")

	ctx, cancel := context.WithTimeout(pctx, cmd.Timeout)
	defer cancel()

	stream, _, err := cmd.Client.Dial(ctx, cmd.Remote.ConnInfo())
	if err != nil {
		return err
	}

	defer func() {
		_ = cmd.Client.Close()
	}()

	switch updated, err := WriteNodeFromNetworkHandler(
		ctx,
		cmd.priv,
		base.NetworkID(cmd.NetworkID),
		cmd.Key,
		input,
		stream,
	); {
	case err != nil:
		return err
	case !updated:
		return errors.Errorf("key, %q; not updated", cmd.Key)
	default:
		cmd.Log.Debug().Msg("updated")
	}

	key := cmd.Key
	if strings.HasPrefix(key, "design.") {
		key = "design._source"
	}

	return cmd.printValue(ctx, key)
}

func (cmd *NetworkClientWriteNodeCommand) Help() string {
	if slices.Contains(os.Args[1:], "acl") {
		return cmd.helpACL()
	}

	s := "## available keys\n\n"

	for i := range launch.AllNodeWriteKeys {
		s += fmt.Sprintf("  - %s\n", launch.AllNodeWriteKeys[i])
	}

	return s
}

func (*NetworkClientWriteNodeCommand) helpACL() string {
	buf := bytes.NewBuffer(nil)

	sp := func(s string, a ...interface{}) {
		_, _ = fmt.Fprintf(buf, s, a...)
	}

	sp("## ACL: Scopes and permissions\n\n")

	m := map[launch.ACLScope][]string{
		launch.DesignACLScope: {
			fmt.Sprintf("read: %s", launch.ReadAllowACLPerm),
			fmt.Sprintf("write: %s", launch.WriteAllowACLPerm),
		},
		launch.StatesAllowConsensusACLScope: {
			fmt.Sprintf("read: %s", launch.ReadAllowACLPerm),
			fmt.Sprintf("write: %s", launch.WriteAllowACLPerm),
		},
		launch.DiscoveryACLScope: {
			fmt.Sprintf("read: %s", launch.ReadAllowACLPerm),
			fmt.Sprintf("write: %s", launch.WriteAllowACLPerm),
		},
		launch.ACLACLScope: {
			fmt.Sprintf("read: %s", launch.ReadAllowACLPerm),
			fmt.Sprintf("write: %s", launch.WriteAllowACLPerm),
		},
		launch.HandoverACLScope:     {launch.WriteAllowACLPerm.String()},
		launch.EventLoggingACLScope: {launch.ReadAllowACLPerm.String()},
	}

	for i := range launch.AllACLScopes {
		scope := launch.AllACLScopes[i]
		required := m[scope]

		var perm string

		if len(required) < 2 {
			perm = fmt.Sprintf(": %s\n", required[0])
		}

		sp("* %s%s\n", scope, perm)

		if len(required) > 1 {
			for j := range required {
				sp("  - %s\n", required[j])
			}
		}
	}

	return buf.String()
}

func WriteNodeFromNetworkHandler(
	ctx context.Context,
	priv base.Privatekey,
	networkID base.NetworkID,
	key string,
	value string,
	stream quicstreamheader.StreamFunc,
) (found bool, _ error) {
	header := launch.NewWriteNodeHeader(key, priv.Publickey())
	if err := header.IsValid(nil); err != nil {
		return false, err
	}

	body := bytes.NewBuffer([]byte(value))
	bodyclosef := func() {
		body.Reset()
	}

	err := stream(ctx, func(ctx context.Context, broker *quicstreamheader.ClientBroker) error {
		if err := broker.WriteRequestHead(ctx, header); err != nil {
			defer bodyclosef()

			return err
		}

		if err := isaacnetwork.VerifyNode(ctx, broker, priv, networkID); err != nil {
			defer bodyclosef()

			return err
		}

		wch := make(chan error, 1)
		go func() {
			defer bodyclosef()

			wch <- broker.WriteBody(ctx, quicstreamheader.StreamBodyType, 0, body)
		}()

		switch _, res, err := broker.ReadResponseHead(ctx); {
		case err != nil:
			return err
		case res.Err() != nil:
			return res.Err()
		case !res.OK():
			return nil
		default:
			found = true

			return <-wch
		}
	})

	return found, err
}

func ReadNodeFromNetworkHandler(
	ctx context.Context,
	priv base.Privatekey,
	networkID base.NetworkID,
	key string,
	stream quicstreamheader.StreamFunc,
) (t interface{}, found bool, _ error) {
	header := launch.NewReadNodeHeader(key, priv.Publickey())
	if err := header.IsValid(nil); err != nil {
		return t, false, err
	}

	err := stream(ctx, func(ctx context.Context, broker *quicstreamheader.ClientBroker) error {
		switch b, i, err := readNodeFromNetworkHandler(ctx, priv, networkID, broker, header); {
		case err != nil:
			return err
		case !i:
			return nil
		default:
			found = true

			if err := yaml.Unmarshal(b, &t); err != nil {
				return errors.WithStack(err)
			}

			return nil
		}
	})

	return t, found, err
}

func readNodeFromNetworkHandler(
	ctx context.Context,
	priv base.Privatekey,
	networkID base.NetworkID,
	broker *quicstreamheader.ClientBroker,
	header launch.ReadNodeHeader,
) ([]byte, bool, error) {
	if err := broker.WriteRequestHead(ctx, header); err != nil {
		return nil, false, err
	}

	if err := isaacnetwork.VerifyNode(ctx, broker, priv, networkID); err != nil {
		return nil, false, err
	}

	switch _, res, err := broker.ReadResponseHead(ctx); {
	case err != nil:
		return nil, false, err
	case res.Err() != nil, !res.OK():
		return nil, res.OK(), res.Err()
	}

	var body io.Reader

	switch bodyType, bodyLength, b, _, res, err := broker.ReadBody(ctx); {
	case err != nil:
		return nil, false, err
	case res != nil:
		return nil, res.OK(), res.Err()
	case bodyType == quicstreamheader.FixedLengthBodyType:
		if bodyLength > 0 {
			body = b
		}
	case bodyType == quicstreamheader.StreamBodyType:
		body = b
	}

	if body == nil {
		return nil, false, errors.Errorf("empty value")
	}

	b, err := io.ReadAll(body)

	return b, true, errors.WithStack(err)
}
