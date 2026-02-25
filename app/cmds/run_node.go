package cmds

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/arl/statsviz"
	"github.com/gorilla/mux"
	"github.com/imfact-labs/currency-model/api"
	"github.com/imfact-labs/currency-model/app/runtime/pipeline"
	"github.com/imfact-labs/currency-model/app/runtime/steps"
	"github.com/imfact-labs/currency-model/digest"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	isaacstates "github.com/imfact-labs/mitum2/isaac/states"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/network/quicmemberlist"
	"github.com/imfact-labs/mitum2/network/quicstream"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/logging"
	"github.com/imfact-labs/mitum2/util/ps"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type RunCommand struct { //nolint:govet //...
	//revive:disable:line-length-limit
	launch.DesignFlag
	launch.DevFlags `embed:"" prefix:"dev."`
	launch.PrivatekeyFlags
	Discovery []launch.ConnInfoFlag `help:"member discovery" placeholder:"ConnInfo"`
	Hold      launch.HeightFlag     `help:"hold consensus states"`
	HTTPState string                `name:"http-state" help:"runtime statistics thru https" placeholder:"bind address"`
	launch.ACLFlags
	exitf  func(error)
	log    *zerolog.Logger
	holded bool
	//revive:enable:line-length-limit
}

func (cmd *RunCommand) Log() *zerolog.Logger {
	return cmd.log
}

func (cmd *RunCommand) SetLog(l *zerolog.Logger) {
	cmd.log = l
}

func (cmd *RunCommand) Run(pctx context.Context) error {
	var log *logging.Logging
	if err := util.LoadFromContextOK(pctx, launch.LoggingContextKey, &log); err != nil {
		return err
	}

	log.Log().Debug().
		Interface("design", cmd.DesignFlag).
		Interface("privatekey", cmd.PrivatekeyFlags).
		Interface("discovery", cmd.Discovery).
		Interface("hold", cmd.Hold).
		Interface("http_state", cmd.HTTPState).
		Interface("dev", cmd.DevFlags).
		Interface("acl", cmd.ACLFlags).
		Msg("flags")

	cmd.log = log.Log()

	if len(cmd.HTTPState) > 0 {
		if err := cmd.RunHTTPState(cmd.HTTPState); err != nil {
			return errors.Wrap(err, "run http state")
		}
	}

	nctx := util.ContextWithValues(pctx, map[util.ContextKey]interface{}{
		launch.DesignFlagContextKey:    cmd.DesignFlag,
		launch.DevFlagsContextKey:      cmd.DevFlags,
		launch.DiscoveryFlagContextKey: cmd.Discovery,
		launch.PrivatekeyContextKey:    string(cmd.PrivatekeyFlags.Flag.Body()),
		launch.ACLFlagsContextKey:      cmd.ACLFlags,
	})

	pps := pipeline.DefaultRunPS()

	_ = pps.AddOK(digest.PNameDigester, digest.ProcessDigester, nil, digest.PNameDigesterDataBase).
		AddOK(digest.PNameStartDigester, digest.ProcessStartDigester, nil, api.PNameStartAPI)
	_ = pps.POK(launch.PNameStorage).PostAddOK(ps.Name("check-hold"), cmd.PCheckHold)
	_ = pps.POK(launch.PNameStates).
		PreAddOK(ps.Name("when-new-block-saved-in-consensus-state-func"), cmd.PWhenNewBlockSavedInConsensusStateFunc).
		PreAddOK(ps.Name("when-new-block-confirmed-func"), cmd.PWhenNewBlockConfirmed).
		PreAddOK(ps.Name("when-new-block-saved-in-syncing-state-func"), cmd.PWhenNewBlockSavedInSyncingStateFunc)
	_ = pps.POK(launch.PNameEncoder).
		PostAddOK(launch.PNameAddHinters, steps.PAddHinters)
	_ = pps.POK(api.PNameAPI).
		PostAddOK(PNameDigestAPIHandlers, cmd.pDigestAPIHandlers)
	_ = pps.POK(digest.PNameDigester).
		PostAddOK(PNameDigesterFollowUp, digest.PdigesterFollowUp)

	_ = pps.SetLogging(log)

	log.Log().Debug().Interface("process", pps.Verbose()).Msg("process ready")

	nctx, err := pps.Run(nctx)
	defer func() {
		log.Log().Debug().Interface("process", pps.Verbose()).Msg("process will be closed")

		if _, err = pps.Close(nctx); err != nil {
			log.Log().Error().Err(err).Msg("failed to close")
		}
	}()

	if err != nil {
		return err
	}

	log.Log().Debug().
		Interface("discovery", cmd.Discovery).
		Interface("hold", cmd.Hold.Height()).
		Msg("node started")

	return cmd.RunNode(nctx)
}

var errHoldStop = util.NewIDError("hold stop")

func (cmd *RunCommand) RunNode(pctx context.Context) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	exitch := make(chan error)

	cmd.exitf = func(err error) {
		exitch <- err
	}

	stopstates := func() {}

	if !cmd.holded {
		deferred, err := cmd.runStates(ctx, pctx)
		if err != nil {
			return err
		}

		stopstates = deferred
	}

	select {
	case <-ctx.Done(): // NOTE graceful stop
		return errors.WithStack(ctx.Err())
	case err := <-exitch:
		if errors.Is(err, errHoldStop) {
			stopstates()

			<-ctx.Done()

			return errors.WithStack(ctx.Err())
		}

		return err
	}
}

func (cmd *RunCommand) runStates(ctx, pctx context.Context) (func(), error) {
	var discoveries *util.Locked[[]quicstream.ConnInfo]
	var states *isaacstates.States

	if err := util.LoadFromContextOK(pctx,
		launch.DiscoveryContextKey, &discoveries,
		launch.StatesContextKey, &states,
	); err != nil {
		return nil, err
	}

	if dis := launch.GetDiscoveriesFromLocked(discoveries); len(dis) < 1 {
		cmd.log.Warn().Msg("empty discoveries; will wait to be joined by remote nodes")
	}

	go func() {
		cmd.exitf(<-states.Wait(ctx))
	}()

	return func() {
		if err := states.Hold(); err != nil && !errors.Is(err, util.ErrDaemonAlreadyStopped) {
			cmd.log.Error().Err(err).Msg("stop states")

			return
		}

		cmd.log.Debug().Msg("states stopped")
	}, nil
}

func (cmd *RunCommand) PWhenNewBlockSavedInSyncingStateFunc(pctx context.Context) (context.Context, error) {
	var log *logging.Logging
	var db isaac.Database
	var design digest.YamlDigestDesign

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.CenterDatabaseContextKey, &db,
		digest.ContextValueDigestDesign, &design,
	); err != nil {
		return pctx, err
	}

	var f func(height base.Height)
	if !design.Equal(digest.YamlDigestDesign{}) && design.Digest {
		var di *digest.Digester
		if err := util.LoadFromContextOK(pctx,
			digest.ContextValueDigester, &di,
		); err != nil {
			return pctx, err
		}

		g := cmd.whenBlockSaved(db, di)

		f = func(height base.Height) {
			g(pctx)
			l := log.Log().With().Interface("height", height).Logger()

			if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
				l.Debug().Msg("will be stopped by hold")
				cmd.exitf(errHoldStop.WithStack())

				return
			}
		}
	} else {
		f = func(height base.Height) {
			l := log.Log().With().Interface("height", height).Logger()

			if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
				l.Debug().Msg("will be stopped by hold")
				cmd.exitf(errHoldStop.WithStack())

				return
			}
		}
	}

	return context.WithValue(pctx,
		launch.WhenNewBlockSavedInSyncingStateFuncContextKey, f,
	), nil
}

func (cmd *RunCommand) PWhenNewBlockSavedInConsensusStateFunc(pctx context.Context) (context.Context, error) {
	var log *logging.Logging

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
	); err != nil {
		return pctx, err
	}

	f := func(bm base.BlockMap) {
		l := log.Log().With().
			Interface("blockmap", bm).
			Interface("height", bm.Manifest().Height()).
			Logger()

		if cmd.Hold.IsSet() && bm.Manifest().Height() == cmd.Hold.Height() {
			l.Debug().Msg("will be stopped by hold")

			cmd.exitf(errHoldStop.WithStack())

			return
		}
	}

	return context.WithValue(pctx, launch.WhenNewBlockSavedInConsensusStateFuncContextKey, f), nil
}

func (cmd *RunCommand) PWhenNewBlockConfirmed(pctx context.Context) (context.Context, error) {
	var log *logging.Logging
	var db isaac.Database
	var design digest.YamlDigestDesign

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.CenterDatabaseContextKey, &db,
		digest.ContextValueDigestDesign, &design,
	); err != nil {
		return pctx, err
	}

	var f func(height base.Height)
	if !design.Equal(digest.YamlDigestDesign{}) && design.Digest {
		f = func(height base.Height) {
			l := log.Log().With().Interface("height", height).Logger()
			err := digest.DigestFollowup(pctx, height)
			if err != nil {
				cmd.exitf(err)

				return
			}

			if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
				l.Debug().Msg("will be stopped by hold")
				cmd.exitf(errHoldStop.WithStack())

				return
			}
		}
	} else {
		f = func(height base.Height) {
			l := log.Log().With().Interface("height", height).Logger()

			if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
				l.Debug().Msg("will be stopped by hold")
				cmd.exitf(errHoldStop.WithStack())

				return
			}
		}
	}

	return context.WithValue(pctx,
		launch.WhenNewBlockConfirmedFuncContextKey, f,
	), nil
}

func (cmd *RunCommand) whenBlockSaved(
	db isaac.Database,
	di *digest.Digester,
) ps.Func {
	return func(ctx context.Context) (context.Context, error) {
		switch m, found, err := db.LastBlockMap(); {
		case err != nil:
			return ctx, err
		case !found:
			return ctx, errors.Errorf("Last BlockMap not found")
		default:
			if di != nil {
				go func() {
					di.Digest([]base.BlockMap{m})
				}()
			}
		}
		return ctx, nil
	}
}

func (cmd *RunCommand) PCheckHold(pctx context.Context) (context.Context, error) {
	var db isaac.Database
	if err := util.LoadFromContextOK(pctx, launch.CenterDatabaseContextKey, &db); err != nil {
		return pctx, err
	}

	switch {
	case !cmd.Hold.IsSet():
	case cmd.Hold.Height() < base.GenesisHeight:
		cmd.holded = true
	default:
		switch m, found, err := db.LastBlockMap(); {
		case err != nil:
			return pctx, err
		case !found:
		case cmd.Hold.Height() <= m.Manifest().Height():
			cmd.holded = true
		}
	}

	return pctx, nil
}

func (cmd *RunCommand) RunHTTPState(bind string) error {
	addr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return errors.Wrap(err, "parse --http-state")
	}

	m := http.NewServeMux()
	if err := statsviz.Register(m); err != nil {
		return errors.Wrap(err, "register statsviz for http-state")
	}

	cmd.log.Debug().Stringer("bind", addr).Msg("statsviz started")

	go func() {
		_ = http.ListenAndServe(addr.String(), m)
	}()

	return nil
}

func LoadCache(log *zerolog.Logger, _ context.Context, design digest.YamlDigestDesign) (api.Cache, error) {
	c, err := api.NewCacheFromURI(design.Cache().String())
	if err != nil {
		log.Err(err).Str("cache", design.Cache().String()).Msg("connect cache server")
		log.Warn().Msg("instead of remote cache server, internal mem cache can be available, `memory://`")

		return nil, err
	}
	return c, nil
}

func SetDigestAPIDefaultHandlers(
	log *zerolog.Logger,
	ctx context.Context,
	params *launch.LocalParams,
	cache api.Cache,
	router *mux.Router,
	queue chan api.RequestWrapper,
) (*api.Handlers, error) {
	var nodeDesign launch.NodeDesign
	var design digest.YamlDigestDesign
	var st *digest.Database
	if err := util.LoadFromContext(ctx,
		launch.DesignContextKey, &nodeDesign,
		digest.ContextValueDigestDesign, &design,
	); err != nil {
		return nil, err
	}
	if design.Digest {
		if err := util.LoadFromContext(ctx, digest.ContextValueDigestDatabase, &st); err != nil {
			return nil, err
		}
	}

	node, err := quicstream.NewConnInfoFromStringAddr(nodeDesign.Network.PublishString, nodeDesign.Network.TLSInsecure)
	if err != nil {
		return nil, err
	}

	handlers := api.NewHandlers(ctx, params.ISAAC.NetworkID(), encs, enc, st, cache, router, queue, node)

	h, err := SetDigestAPINetworkClient(log, ctx, params, handlers)
	if err != nil {
		return nil, err
	}
	handlers = h

	return handlers, nil
}

func SetDigestAPINetworkClient(
	log *zerolog.Logger,
	ctx context.Context,
	params *launch.LocalParams,
	handlers *api.Handlers,
) (*api.Handlers, error) {
	var memberList *quicmemberlist.Memberlist
	if err := util.LoadFromContextOK(ctx, launch.MemberlistContextKey, &memberList); err != nil {
		return nil, err
	}

	connectionPool, err := launch.NewConnectionPool(
		1<<9,
		params.ISAAC.NetworkID(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	handlers = handlers.SetNetworkClientFunc(
		func() (*quicstream.ConnectionPool, *quicmemberlist.Memberlist, []quicstream.ConnInfo, error) { // nolint:contextcheck
			return connectionPool, memberList, []quicstream.ConnInfo{}, nil
		},
	)

	var design digest.YamlDigestDesign
	if err := util.LoadFromContext(ctx, digest.ContextValueDigestDesign, &design); err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return handlers, nil
		}

		return nil, err
	}

	if design.Equal(digest.YamlDigestDesign{}) {
		return handlers, nil
	}

	//handlers = handlers.SetConnectionPool(connectionPool)

	handlers = handlers.SetNetworkClientFunc(
		func() (*quicstream.ConnectionPool, *quicmemberlist.Memberlist, []quicstream.ConnInfo, error) { // nolint:contextcheck
			return connectionPool, memberList, design.ConnInfo, nil
		},
	)

	log.Debug().Msg("send handler attached")

	return handlers, nil
}
