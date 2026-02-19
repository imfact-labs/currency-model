package digest

import (
	"context"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	isaacdatabase "github.com/imfact-labs/mitum2/isaac/database"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/logging"
	"github.com/imfact-labs/mitum2/util/ps"
	"github.com/pkg/errors"
)

const (
	PNameDigester      = ps.Name("digester")
	PNameStartDigester = ps.Name("start_digester")
)

func ProcessDigester(ctx context.Context) (context.Context, error) {
	var vs util.Version
	var log *logging.Logging
	var digestDesign YamlDigestDesign

	if err := util.LoadFromContextOK(ctx,
		launch.VersionContextKey, &vs,
		launch.LoggingContextKey, &log,
		ContextValueDigestDesign, &digestDesign,
	); err != nil {
		return ctx, err
	}

	if digestDesign.Equal(YamlDigestDesign{}) || !digestDesign.Digest {
		return ctx, nil
	}
	var st *Database
	if err := util.LoadFromContextOK(ctx,
		ContextValueDigestDatabase, &st,
	); err != nil {
		return ctx, err
	}

	if st == nil {
		return ctx, nil
	}

	var design launch.NodeDesign
	if err := util.LoadFromContext(ctx,
		launch.DesignContextKey, &design,
	); err != nil {
		return ctx, err
	}
	root := launch.LocalFSDataDirectory(design.Storage.Base)

	var newReaders func(context.Context, string, *isaac.BlockItemReadersArgs) (*isaac.BlockItemReaders, error)
	var fromRemotes isaac.RemotesBlockItemReadFunc

	if err := util.LoadFromContextOK(ctx,
		launch.NewBlockItemReadersFuncContextKey, &newReaders,
		launch.RemotesBlockItemReaderFuncContextKey, &fromRemotes,
	); err != nil {
		return ctx, err
	}

	var sourceReaders *isaac.BlockItemReaders

	switch i, err := newReaders(ctx, root, nil); {
	case err != nil:
		return ctx, err
	default:
		sourceReaders = i
	}

	di := NewDigester(st, root, sourceReaders, fromRemotes, design.NetworkID, vs.String(), nil)
	_ = di.SetLogging(log)
	di.PrepareFunc = []BlockSessionPrepareFunc{PrepareCurrencies, PrepareAccounts, PrepareDIDRegistry}

	return context.WithValue(ctx, ContextValueDigester, di), nil
}

func ProcessStartDigester(ctx context.Context) (context.Context, error) {
	var di *Digester
	var digestDesign YamlDigestDesign

	if err := util.LoadFromContext(ctx,
		ContextValueDigester, &di,
		ContextValueDigestDesign, &digestDesign,
	); err != nil {
		return ctx, err
	}

	if digestDesign.Equal(YamlDigestDesign{}) || !digestDesign.Digest || di == nil {
		return ctx, nil
	}

	return ctx, di.Start(ctx)
}

func PdigesterFollowUp(ctx context.Context) (context.Context, error) {
	var log *logging.Logging
	if err := util.LoadFromContextOK(ctx, launch.LoggingContextKey, &log); err != nil {
		return ctx, err
	}

	log.Log().Debug().Msg("digester trying to follow up")

	var mst *isaacdatabase.Center
	var digestDesign YamlDigestDesign
	if err := util.LoadFromContextOK(ctx,
		launch.CenterDatabaseContextKey, &mst,
		ContextValueDigestDesign, &digestDesign,
	); err != nil {
		return ctx, err
	}

	if digestDesign.Equal(YamlDigestDesign{}) || !digestDesign.Digest {
		return ctx, nil
	}

	var st *Database
	if err := util.LoadFromContext(ctx, ContextValueDigestDatabase, &st); err != nil {
		return ctx, err
	}

	if st == nil {
		return ctx, nil
	}

	switch m, found, err := mst.LastBlockMap(); {
	case err != nil:
		return ctx, err
	case !found:
		log.Log().Debug().Msg("last BlockMap not found")
	case m.Manifest().Height() > st.LastBlock():
		log.Log().Info().
			Int64("last_manifest", m.Manifest().Height().Int64()).
			Int64("last_block", st.LastBlock().Int64()).
			Msg("new blocks found to digest")

		if err := DigestFollowup(ctx, m.Manifest().Height()); err != nil {
			log.Log().Error().Err(err).Msg("follow up")

			return ctx, err
		}
		log.Log().Info().Msg("digested new blocks")
	default:
		log.Log().Info().Msg("digested blocks is up-to-dated")
	}

	return ctx, nil
}

func DigestFollowup(ctx context.Context, height base.Height) error {
	var vs util.Version
	if err := util.LoadFromContextOK(ctx, launch.VersionContextKey, &vs); err != nil {
		return err
	}

	var di *Digester
	if err := util.LoadFromContext(ctx,
		ContextValueDigester, &di,
	); err != nil {
		return err
	}

	var st *Database
	if di.Database() != nil {
		st = di.Database()
	} else {
		return errors.New("Digester Database is nil")
	}

	if height <= st.LastBlock() {
		return nil
	}

	lastBlock := st.LastBlock()
	if lastBlock < base.GenesisHeight {
		lastBlock = base.GenesisHeight
	}

	for h := lastBlock; h <= height; h++ {
		if err := di.DigestBlockMap(ctx, h); err != nil {
			return err
		}
	}
	return nil
}
