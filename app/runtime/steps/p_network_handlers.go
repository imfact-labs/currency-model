package steps

import (
	"context"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/mitum2/isaac/database"
	"github.com/imfact-labs/mitum2/isaac/network"
	"github.com/imfact-labs/mitum2/isaac/states"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/network/quicmemberlist"
	"github.com/imfact-labs/mitum2/storage"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/logging"
	"github.com/imfact-labs/mitum2/util/valuehash"
)

func PNetworkHandlers(pctx context.Context) (context.Context, error) {
	e := util.StringError("prepare network handlers")

	var log *logging.Logging
	var encs *encoder.Encoders
	var design launch.NodeDesign
	var local base.LocalNode
	var params *launch.LocalParams
	var db isaac.Database
	var pool *isaacdatabase.TempPool
	var proposalMaker *isaac.ProposalMaker
	var m *quicmemberlist.Memberlist
	var syncSourcePool *isaac.SyncSourcePool
	var nodeinfo *isaacnetwork.NodeInfoUpdater
	var svVoteF isaac.SuffrageVoteFunc
	var ballotBox *isaacstates.Ballotbox
	var filterNotifyMsg quicmemberlist.FilterNotifyMsgFunc
	var lvps *isaac.LastVoteproofsHandler
	var metricsCollector *isaacnetwork.NetworkMetricsCollector

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.EncodersContextKey, &encs,
		launch.DesignContextKey, &design,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
		launch.CenterDatabaseContextKey, &db,
		launch.PoolDatabaseContextKey, &pool,
		launch.ProposalMakerContextKey, &proposalMaker,
		launch.MemberlistContextKey, &m,
		launch.SyncSourcePoolContextKey, &syncSourcePool,
		launch.NodeInfoContextKey, &nodeinfo,
		launch.SuffrageVotingVoteFuncContextKey, &svVoteF,
		launch.BallotboxContextKey, &ballotBox,
		launch.FilterMemberlistNotifyMsgFuncContextKey, &filterNotifyMsg,
		launch.LastVoteproofsHandlerContextKey, &lvps,
		launch.MetricsCollectorContextKey, &metricsCollector,
	); err != nil {
		return pctx, e.Wrap(err)
	}

	isaacParams := params.ISAAC

	lastBlockMapF := launch.QuicstreamHandlerLastBlockMapFunc(db)
	suffrageNodeConnInfoF := launch.QuicstreamHandlerSuffrageNodeConnInfoFunc(db, m)

	var gerror error

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameLastSuffrageProof,
		isaacnetwork.QuicstreamHandlerLastSuffrageProof(
			func(last util.Hash) (string, []byte, []byte, bool, error) {
				enchint, metab, body, found, lastheight, err := db.LastSuffrageProofBytes()

				switch {
				case err != nil:
					return enchint, nil, nil, false, err
				case !found:
					return enchint, nil, nil, false, storage.ErrNotFound.Errorf("Last SuffrageProof not found")
				}

				switch {
				case last != nil && len(metab) > 0 && valuehash.NewBytes(metab).Equal(last):
					nbody, _ := util.NewLengthedBytesSlice([][]byte{lastheight.Bytes(), nil})

					return enchint, nil, nbody, false, nil
				default:
					nbody, _ := util.NewLengthedBytesSlice([][]byte{lastheight.Bytes(), body})

					return enchint, metab, nbody, true, nil
				}
			},
		), nil)

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameSuffrageProof,
		isaacnetwork.QuicstreamHandlerSuffrageProof(db.SuffrageProofBytes), nil)

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameLastBlockMap,
		isaacnetwork.QuicstreamHandlerLastBlockMap(lastBlockMapF), nil)

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameBlockMap,
		isaacnetwork.QuicstreamHandlerBlockMap(db.BlockMapBytes), nil)

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameNodeChallenge,
		isaacnetwork.QuicstreamHandlerNodeChallenge(isaacParams.NetworkID(), local), nil)

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameSuffrageNodeConnInfo,
		isaacnetwork.QuicstreamHandlerSuffrageNodeConnInfo(suffrageNodeConnInfoF), nil)

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameSyncSourceConnInfo,
		isaacnetwork.QuicstreamHandlerSyncSourceConnInfo(
			func() ([]isaac.NodeConnInfo, error) {
				members := make([]isaac.NodeConnInfo, syncSourcePool.Len()*2)

				var i int
				syncSourcePool.Actives(func(nci isaac.NodeConnInfo) bool {
					members[i] = nci
					i++

					return true
				})

				return members[:i], nil
			},
		), nil)

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameState,
		isaacnetwork.QuicstreamHandlerState(db.StateBytes), nil)

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameExistsInStateOperation,
		isaacnetwork.QuicstreamHandlerExistsInStateOperation(db.ExistsInStateOperation), nil)

	if vp := lvps.Last().Cap(); vp != nil {
		_ = nodeinfo.SetLastVote(vp.Point(), vp.Result())
	}

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameNodeInfo,
		isaacnetwork.QuicstreamHandlerNodeInfo(launch.QuicstreamHandlerGetNodeInfoFunc(encs.Default(), nodeinfo)), nil)

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameNodeMetrics,
		isaacnetwork.QuicstreamHandlerNodeMetrics(
			launch.QuicstreamHandlerGetNodeMetricsFunc(encs.Default(), metricsCollector),
		),
		nil)

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerNameSendBallots,
		isaacnetwork.QuicstreamHandlerSendBallots(isaacParams.NetworkID(),
			func(bl base.BallotSignFact) error {
				switch passed, err := filterNotifyMsg(bl); {
				case err != nil:
					log.Log().Trace().
						Str("module", "filter-notify-msg-send-ballots").
						Err(err).
						Interface("handover_message", bl).
						Msg("filter error")

					fallthrough
				case !passed:
					log.Log().Trace().
						Str("module", "filter-notify-msg-send-ballots").
						Interface("handover_message", bl).
						Msg("filtered")

					return nil
				}

				_, err := ballotBox.VoteSignFact(bl)

				return err
			},
			params.MISC.MaxMessageSize,
		), nil)

	if gerror != nil {
		return pctx, gerror
	}

	if err := launch.AttachBlockItemsNetworkHandlers(pctx); err != nil {
		return pctx, err
	}

	if err := launch.AttachMemberlistNetworkHandlers(pctx); err != nil {
		return pctx, err
	}

	return pctx, nil
}
