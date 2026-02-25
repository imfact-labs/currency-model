package steps

import (
	"context"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/logging"
	"github.com/imfact-labs/mitum2/util/ps"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var PNameGenerateGenesis = ps.Name("mitum-currency-generate-genesis")

func PGenerateGenesis(pctx context.Context) (context.Context, error) {
	e := util.StringError("generate genesis block")

	var log *logging.Logging
	var design launch.NodeDesign
	var genesisDesign launch.GenesisDesign
	var encs *encoder.Encoders
	var local base.LocalNode
	var isaacParams *isaac.Params
	var db isaac.Database
	var fsnodeinfo launch.NodeInfo
	var eventLogging *launch.EventLogging
	var newReaders func(context.Context, string, *isaac.BlockItemReadersArgs) (*isaac.BlockItemReaders, error)

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.DesignContextKey, &design,
		launch.GenesisDesignContextKey, &genesisDesign,
		launch.EncodersContextKey, &encs,
		launch.LocalContextKey, &local,
		launch.ISAACParamsContextKey, &isaacParams,
		launch.CenterDatabaseContextKey, &db,
		launch.FSNodeInfoContextKey, &fsnodeinfo,
		launch.EventLoggingContextKey, &eventLogging,
		launch.NewBlockItemReadersFuncContextKey, &newReaders,
	); err != nil {
		return pctx, e.Wrap(err)
	}

	var el zerolog.Logger

	switch i, found := eventLogging.Logger(launch.NodeEventLogger); {
	case !found:
		return pctx, errors.Errorf("Node event logger not found")
	default:
		el = i
	}

	root := launch.LocalFSDataDirectory(design.Storage.Base)

	var readers *isaac.BlockItemReaders

	switch i, err := newReaders(pctx, root, nil); {
	case err != nil:
		return pctx, err
	default:
		defer i.Close()

		readers = i
	}

	g := NewGenesisBlockGenerator(
		local,
		isaacParams.NetworkID(),
		encs,
		db,
		root,
		genesisDesign.Facts,
		func() (base.BlockMap, bool, error) {
			return isaac.BlockItemReadersDecode[base.BlockMap](
				readers.Item,
				base.GenesisHeight,
				base.BlockItemMap,
				nil,
			)
		},
		pctx,
	)
	_ = g.SetLogging(log)

	if _, err := g.Generate(); err != nil {
		return pctx, e.Wrap(err)
	}

	el.Debug().Interface("node_info", fsnodeinfo).Msg("node initialized")

	return pctx, nil
}
