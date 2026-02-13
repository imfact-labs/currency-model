package cmds

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/imfact-labs/imfact-currency/digest"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
	isaacstates "github.com/ProtoconNet/mitum2/isaac/states"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/network/quicmemberlist"
	"github.com/ProtoconNet/mitum2/network/quicstream"
	quicstreamheader "github.com/ProtoconNet/mitum2/network/quicstream/header"
	nutil "github.com/ProtoconNet/mitum2/network/util"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type (
	//revive:disable:line-length-limit
	writeNodeValueFunc     func(_ context.Context, key, value, acluser string) (prev, next interface{}, updated bool, _ error)
	writeNodeNextValueFunc func(_ context.Context, key, nextkey, value, acluser string) (prev, next interface{}, updated bool, _ error)
	readNodeValueFunc      func(_ context.Context, key, acluser string) (interface{}, error)
	readNodeNextValueFunc  func(_ context.Context, key, nextkey, acluser string) (interface{}, error)
	//revive:enable:line-length-limit
)

var NodeReadWriteEventLogger launch.EventLoggerName = "node_readwrite"

var (
	DesignACLScope               = launch.ACLScope("design")
	StatesAllowConsensusACLScope = launch.ACLScope("states.allow_consensus")
	DiscoveryACLScope            = launch.ACLScope("discovery")
	ACLACLScope                  = launch.ACLScope("acl")
	BlockItemFilesACLScope       = launch.ACLScope("block_item_files")
)

func PNetworkHandlersReadWriteNode(pctx context.Context) (context.Context, error) {
	var design launch.NodeDesign
	var local base.LocalNode
	var eventLogging *launch.EventLogging

	if err := util.LoadFromContextOK(pctx,
		launch.DesignContextKey, &design,
		launch.LocalContextKey, &local,
		launch.EventLoggingContextKey, &eventLogging,
	); err != nil {
		return pctx, err
	}

	var rl, wl zerolog.Logger

	switch el, found := eventLogging.Logger(NodeReadWriteEventLogger); {
	case !found:
		return pctx, errors.Errorf("node read/write event logger not found")
	default:
		rl = el.With().Str("module", "read").Logger()
		wl = el.With().Str("module", "write").Logger()
	}

	lock := &sync.RWMutex{}

	rf, err := readNode(pctx, lock)
	if err != nil {
		return pctx, err
	}

	wf, err := writeNode(pctx, lock)
	if err != nil {
		return pctx, err
	}

	var gerror error

	launch.EnsureHandlerAdd(pctx, &gerror,
		launch.HandlerNameNodeRead,
		networkHandlerNodeRead(design.LocalParams.ISAAC.NetworkID(), rf, rl), nil)

	launch.EnsureHandlerAdd(pctx, &gerror,
		launch.HandlerNameNodeWrite,
		networkHandlerNodeWrite(design.LocalParams.ISAAC.NetworkID(), wf, wl), nil)

	return pctx, gerror
}

func writeNodeKey(f writeNodeNextValueFunc) writeNodeValueFunc {
	return func(ctx context.Context, key, value, acluser string) (interface{}, interface{}, bool, error) {
		i := strings.SplitN(strings.TrimPrefix(key, "."), ".", 2)

		var nextkey string
		if len(i) > 1 {
			nextkey = i[1]
		}

		return f(ctx, i[0], nextkey, value, acluser)
	}
}

func writeNode(pctx context.Context, lock *sync.RWMutex) (writeNodeValueFunc, error) {
	fStates, err := writeStates(pctx)
	if err != nil {
		return nil, err
	}

	fDesign, err := writeDesign(pctx)
	if err != nil {
		return nil, err
	}

	fDiscovery, err := writeDiscovery(pctx)
	if err != nil {
		return nil, err
	}

	fACL, err := writeACL(pctx)
	if err != nil {
		return nil, err
	}

	fBlockItemFiles, err := writeBlockItemFiles(pctx)
	if err != nil {
		return nil, err
	}

	return writeNodeKey(func(
		ctx context.Context, key, nextkey, value, acluser string,
	) (interface{}, interface{}, bool, error) {
		lock.Lock()
		defer lock.Unlock()

		switch key {
		case "states":
			return fStates(ctx, nextkey, value, acluser)
		case "design":
			return fDesign(ctx, nextkey, value, acluser)
		case "discovery":
			return fDiscovery(ctx, nextkey, value, acluser)
		case "acl":
			return fACL(ctx, nextkey, value, acluser)
		case "block_item_files":
			return fBlockItemFiles(ctx, nextkey, value, acluser)
		default:
			return nil, nil, false, util.ErrNotFound.Errorf("unknown key, %q for params", key)
		}
	}), nil
}

func writeStates(pctx context.Context) (writeNodeValueFunc, error) {
	fAllowConsensus, err := writeAllowConsensus(pctx)
	if err != nil {
		return nil, err
	}

	return writeNodeKey(func(
		ctx context.Context, key, nextkey, value, acluser string,
	) (interface{}, interface{}, bool, error) {
		switch key {
		case "allow_consensus":
			return fAllowConsensus(ctx, nextkey, value, acluser)
		default:
			return nil, nil, false, util.ErrNotFound.Errorf("unknown key, %q for node", key)
		}
	}), nil
}

func writeDesign(pctx context.Context) (writeNodeValueFunc, error) {
	var log *logging.Logging
	var flag launch.DesignFlag
	var encs *encoder.Encoders

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.DesignFlagContextKey, &flag,
		launch.EncodersContextKey, &encs,
	); err != nil {
		return nil, err
	}

	defaultReadDesignFileF := func() (*launch.NodeDesign, *digest.YamlDigestDesign, error) {
		return nil, nil, errors.Errorf("design file; can not read")
	}
	readDesignFileF := defaultReadDesignFileF
	writeDesignFileF := func([]byte) error { return errors.Errorf("design file; can not write") }

	switch i, err := readDesignFileToNodeWriteFunc(flag, encs); {
	case err != nil:
		return nil, err
	case i == nil:
		log.Log().Warn().Stringer("design", flag.URL()).Msg("design file not writable")
	default:
		readDesignFileF = i
	}

	switch i, err := writeDesignFileFunc(flag); {
	case err != nil:
		return nil, err
	case i == nil:
		readDesignFileF = defaultReadDesignFileF

		log.Log().Warn().Stringer("design", flag.URL()).Msg("design file not writable")
	default:
		writeDesignFileF = i
	}

	var m map[string]writeNodeValueFunc

	switch i, err := writeDesignMap(pctx); {
	case err != nil:
		return nil, err
	default:
		m = i
	}

	var aclallow launch.ACLAllowFunc

	switch i, err := launch.PACLAllowFunc(pctx); {
	case err != nil:
		return nil, err
	default:
		aclallow = i
	}

	return func(ctx context.Context, key, value, acluser string) (interface{}, interface{}, bool, error) {
		extra := zerolog.Dict().
			Str("key", key)

		if !aclallow(ctx, acluser, DesignACLScope, launch.WriteAllowACLPerm, extra) {
			return nil, nil, false, launch.ErrACLAccessDenied.WithStack()
		}

		switch f, found := m[key]; {
		case !found:
			return nil, nil, false, util.ErrNotFound.Errorf("unknown key, %q for design", key)
		default:
			switch prev, next, updated, err := f(ctx, key, value, acluser); {
			case err != nil:
				return nil, nil, false, err
			case !updated:
				return prev, next, false, nil
			default:
				return prev, next, updated, errors.WithMessage(
					writeDesignFile(key, value, readDesignFileF, writeDesignFileF),
					"updated in memory, but failed to update design file",
				)
			}
		}
	}, nil
}

func writeDesignMap(pctx context.Context) (map[string]writeNodeValueFunc, error) {
	var encs *encoder.Encoders
	var design launch.NodeDesign
	var params *launch.LocalParams
	var syncSourceChecker *isaacnetwork.SyncSourceChecker

	if err := util.LoadFromContextOK(pctx,
		launch.EncodersContextKey, &encs,
		launch.DesignContextKey, &design,
		launch.LocalParamsContextKey, &params,
		launch.SyncSourceCheckerContextKey, &syncSourceChecker,
	); err != nil {
		return nil, err
	}

	m := map[string]writeNodeValueFunc{
		//revive:disable:line-length-limit
		"parameters.isaac.threshold":                           writeLocalParamISAACThreshold(params.ISAAC),
		"parameters.isaac.interval_broadcast_ballot":           writeLocalParamISAACIntervalBroadcastBallot(params.ISAAC),
		"parameters.isaac.wait_preparing_init_ballot":          writeLocalParamISAACWaitPreparingINITBallot(params.ISAAC),
		"parameters.isaac.ballot_stuck_wait":                   writeLocalParamISAACBallotStuckWait(params.ISAAC),
		"parameters.isaac.ballot_stuck_resolve_after":          writeLocalParamISAACBallotStuckResolveAfter(params.ISAAC),
		"parameters.isaac.min_wait_next_block_init_ballot":     writeLocalParamISAACMinWaitNextBlockINITBallot(params.ISAAC),
		"parameters.isaac.syncer_last_block_map_interval":      writeLocalParamISAACSyncerLastBlockMapInterval(params.ISAAC),
		"parameters.isaac.min_proposer_wait":                   writeLocalParamISAACMinProposerWait(params.ISAAC),
		"parameters.isaac.max_try_handover_y_broker_sync_data": writeLocalParamISAACMaxTryHandoverYBrokerSyncData(params.ISAAC),
		"parameters.isaac.state_cache_size":                    writeLocalParamISAACStateCacheSize(params.ISAAC),
		"parameters.isaac.operation_pool_cache_size":           writeLocalParamISAACOperationPoolCacheSize(params.ISAAC),

		"parameters.misc.sync_source_checker_interval":              writeLocalParamMISCSyncSourceCheckerInterval(params.MISC),
		"parameters.misc.valid_proposal_operation_expire":           writeLocalParamMISCValidProposalOperationExpire(params.MISC),
		"parameters.misc.valid_proposal_suffrage_operations_expire": writeLocalParamMISCValidProposalSuffrageOperationsExpire(params.MISC),
		"parameters.misc.block_item_readers_remove_empty_after":     writeLocalParamMISCBlockItemReadersRemoveEmptyAfter(params.MISC),
		"parameters.misc.block_item_readers_remove_empty_interval":  writeLocalParamMISCBlockItemReadersRemoveEmptyInterval(params.MISC),
		"parameters.misc.max_message_size":                          writeLocalParamMISCMaxMessageSize(params.MISC),

		"parameters.memberlist.extra_same_member_limit":    writeLocalParamMemberlistExtraSameMemberLimit(params.Memberlist),
		"parameters.memberlist.tcp_timeout":                writeLocalParamMemberlistTcpTimeout(params.Memberlist),
		"parameters.memberlist.retransmit_mult":            writeLocalParamMemberlistRetransmitMult(params.Memberlist),
		"parameters.memberlist.probe_timeout":              writeLocalParamMemberlistProbeTimeout(params.Memberlist),
		"parameters.memberlist.probe_interval":             writeLocalParamMemberlistProbeInterval(params.Memberlist),
		"parameters.memberlist.gossip_interval":            writeLocalParamMemberlistGossipInterval(params.Memberlist),
		"parameters.memberlist.gossip_nodes":               writeLocalParamMemberlistGossipNodes(params.Memberlist),
		"parameters.memberlist.suspicion_mult":             writeLocalParamMemberlistSuspicionMult(params.Memberlist),
		"parameters.memberlist.suspicion_max_timeout_mult": writeLocalParamMemberlistSuspicionMaxTimeoutMult(params.Memberlist),
		"parameters.memberlist.udp_buffer_size":            writeLocalParamMemberlistUdpBufferSize(params.Memberlist),
		"parameters.memberlist.broadcast_timer_mult":       writeLocalParamMemberlistBroadcastTimerMult(params.Memberlist),
		"parameters.memberlist.user_msg_loop_interval":     writeLocalParamMemberlistUserMsgLoopInterval(params.Memberlist),

		"parameters.network.timeout_request":    writeLocalParamNetworkTimeoutRequest(params.Network),
		"parameters.network.ratelimit.node":     writeLocalParamNetworkRateLimit(params.Network.RateLimit(), "node"),
		"parameters.network.ratelimit.net":      writeLocalParamNetworkRateLimit(params.Network.RateLimit(), "net"),
		"parameters.network.ratelimit.suffrage": writeLocalParamNetworkRateLimit(params.Network.RateLimit(), "suffrage"),
		"parameters.network.ratelimit.default":  writeLocalParamNetworkRateLimit(params.Network.RateLimit(), "default"),

		"parameters.sync_sources": writeSyncSources(encs.JSON(), design, syncSourceChecker),
		//revive:enable:line-length-limit
	}

	return m, nil
}

func writeLocalParamISAACThreshold(
	params *isaac.Params,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		var s string
		if err := yaml.Unmarshal([]byte(value), &s); err != nil {
			return nil, nil, false, errors.WithStack(err)
		}

		var t base.Threshold
		if err := t.UnmarshalText([]byte(s)); err != nil {
			return nil, nil, false, errors.WithMessagef(err, "invalid threshold, %q", value)
		}

		prevt := params.Threshold()
		if prevt.Equal(t) {
			return prevt, nil, false, nil
		}

		if err := params.SetThreshold(t); err != nil {
			return nil, nil, false, err
		}

		return prevt, params.Threshold(), true, nil
	})
}

func writeLocalParamISAACIntervalBroadcastBallot(
	params *isaac.Params,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.IntervalBroadcastBallot()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetIntervalBroadcastBallot(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.IntervalBroadcastBallot(), true, nil
	})
}

func writeLocalParamISAACWaitPreparingINITBallot(
	params *isaac.Params,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.WaitPreparingINITBallot()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetWaitPreparingINITBallot(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.WaitPreparingINITBallot(), true, nil
	})
}

func writeLocalParamISAACBallotStuckWait(
	params *isaac.Params,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.BallotStuckWait()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetBallotStuckWait(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.BallotStuckWait(), true, nil
	})
}

func writeLocalParamISAACBallotStuckResolveAfter(
	params *isaac.Params,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.BallotStuckResolveAfter()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetBallotStuckResolveAfter(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.BallotStuckResolveAfter(), true, nil
	})
}

func writeLocalParamISAACMinWaitNextBlockINITBallot(
	params *isaac.Params,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.MinWaitNextBlockINITBallot()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetMinWaitNextBlockINITBallot(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.MinWaitNextBlockINITBallot(), true, nil
	})
}

func writeLocalParamISAACSyncerLastBlockMapInterval(
	params *isaac.Params,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.SyncerLastBlockMapInterval()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetSyncerLastBlockMapInterval(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.SyncerLastBlockMapInterval(), true, nil
	})
}

func writeLocalParamISAACMinProposerWait(
	params *isaac.Params,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.MinProposerWait()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetMinProposerWait(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.MinProposerWait(), true, nil
	})
}

func writeLocalParamISAACMaxTryHandoverYBrokerSyncData(
	params *isaac.Params,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, nil, false, errors.WithStack(err)
		}

		prev = params.MaxTryHandoverYBrokerSyncData()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetMaxTryHandoverYBrokerSyncData(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.MaxTryHandoverYBrokerSyncData(), true, nil
	})
}

func writeLocalParamISAACStateCacheSize(
	params *isaac.Params,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		var s string
		if err := yaml.Unmarshal([]byte(value), &s); err != nil {
			return nil, nil, false, errors.WithStack(err)
		}

		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, nil, false, errors.WithStack(err)
		}
		prev = params.StateCacheSize()
		if prev == i {
			return prev, nil, false, nil
		}

		if err := params.SetStateCacheSize(i); err != nil {
			return nil, nil, false, err
		}

		return prev, params.StateCacheSize(), true, nil
	})
}

func writeLocalParamISAACOperationPoolCacheSize(
	params *isaac.Params,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		var s string
		if err := yaml.Unmarshal([]byte(value), &s); err != nil {
			return nil, nil, false, errors.WithStack(err)
		}

		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, nil, false, errors.WithStack(err)
		}

		prev = params.OperationPoolCacheSize()
		if prev == i {
			return prev, nil, false, nil
		}

		if err := params.SetOperationPoolCacheSize(i); err != nil {
			return nil, nil, false, err
		}

		return prev, params.OperationPoolCacheSize(), true, nil
	})
}

func writeLocalParamMISCSyncSourceCheckerInterval(
	params *isaac.MISCParams,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.SyncSourceCheckerInterval()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetSyncSourceCheckerInterval(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.SyncSourceCheckerInterval(), true, nil
	})
}

func writeLocalParamMISCValidProposalOperationExpire(
	params *isaac.MISCParams,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.ValidProposalOperationExpire()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetValidProposalOperationExpire(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.ValidProposalOperationExpire(), true, nil
	})
}

func writeLocalParamMISCValidProposalSuffrageOperationsExpire(
	params *isaac.MISCParams,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.ValidProposalSuffrageOperationsExpire()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetValidProposalSuffrageOperationsExpire(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.ValidProposalSuffrageOperationsExpire(), true, nil
	})
}

func writeLocalParamMISCBlockItemReadersRemoveEmptyAfter(
	params *isaac.MISCParams,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.BlockItemReadersRemoveEmptyAfter()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetBlockItemReadersRemoveEmptyAfter(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.BlockItemReadersRemoveEmptyAfter(), true, nil
	})
}

func writeLocalParamMISCBlockItemReadersRemoveEmptyInterval(
	params *isaac.MISCParams,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.BlockItemReadersRemoveEmptyInterval()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetBlockItemReadersRemoveEmptyInterval(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.BlockItemReadersRemoveEmptyInterval(), true, nil
	})
}

func writeLocalParamMISCMaxMessageSize(
	params *isaac.MISCParams,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, nil, false, errors.WithStack(err)
		}

		prev = params.MaxMessageSize()
		if prev == i {
			return prev, nil, false, nil
		}

		if err := params.SetMaxMessageSize(i); err != nil {
			return nil, nil, false, err
		}

		return prev, params.MaxMessageSize(), true, nil
	})
}

func writeLocalParamMemberlistExtraSameMemberLimit(
	params *quicmemberlist.MemberlistParams,
) writeNodeValueFunc {
	return writeNodeKey(
		func(_ context.Context, _, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
			i, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			prev = params.ExtraSameMemberLimit()
			if prev == i {
				return prev, nil, false, nil
			}

			if err := params.SetExtraSameMemberLimit(i); err != nil {
				return nil, nil, false, err
			}

			return prev, params.ExtraSameMemberLimit(), true, nil
		})
}

func writeLocalParamMemberlistTcpTimeout(
	params *quicmemberlist.MemberlistParams,
) writeNodeValueFunc {
	return writeNodeKey(
		func(_ context.Context, _, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
			d, err := parseNodeValueDuration(value)
			if err != nil {
				return nil, nil, false, err
			}

			prev = params.TCPTimeout()
			if prev == d {
				return prev, nil, false, nil
			}

			if err := params.SetTCPTimeout(d); err != nil {
				return nil, nil, false, err
			}

			return prev, params.TCPTimeout(), true, nil
		})
}

func writeLocalParamMemberlistRetransmitMult(
	params *quicmemberlist.MemberlistParams,
) writeNodeValueFunc {
	return writeNodeKey(
		func(_ context.Context, _, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
			var s string
			if err := yaml.Unmarshal([]byte(value), &s); err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			i, err := strconv.Atoi(s)
			if err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			prev = params.RetransmitMult()
			if prev == i {
				return prev, nil, false, nil
			}

			if err := params.SetRetransmitMult(i); err != nil {
				return nil, nil, false, err
			}

			return prev, params.RetransmitMult(), true, nil
		})
}

func writeLocalParamMemberlistProbeTimeout(
	params *quicmemberlist.MemberlistParams,
) writeNodeValueFunc {
	return writeNodeKey(
		func(_ context.Context, _, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
			d, err := parseNodeValueDuration(value)
			if err != nil {
				return nil, nil, false, err
			}

			prev = params.ProbeTimeout()
			if prev == d {
				return prev, nil, false, nil
			}

			if err := params.SetProbeTimeout(d); err != nil {
				return nil, nil, false, err
			}

			return prev, params.ProbeTimeout(), true, nil
		})
}

func writeLocalParamMemberlistProbeInterval(
	params *quicmemberlist.MemberlistParams,
) writeNodeValueFunc {
	return writeNodeKey(
		func(_ context.Context, _, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
			d, err := parseNodeValueDuration(value)
			if err != nil {
				return nil, nil, false, err
			}

			prev = params.ProbeInterval()
			if prev == d {
				return prev, nil, false, nil
			}

			if err := params.SetProbeInterval(d); err != nil {
				return nil, nil, false, err
			}

			return prev, params.ProbeInterval(), true, nil
		})
}

func writeLocalParamMemberlistGossipInterval(
	params *quicmemberlist.MemberlistParams,
) writeNodeValueFunc {
	return writeNodeKey(
		func(_ context.Context, _, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
			d, err := parseNodeValueDuration(value)
			if err != nil {
				return nil, nil, false, err
			}

			prev = params.GossipInterval()
			if prev == d {
				return prev, nil, false, nil
			}

			if err := params.SetGossipInterval(d); err != nil {
				return nil, nil, false, err
			}

			return prev, params.GossipInterval(), true, nil
		})
}

func writeLocalParamMemberlistGossipNodes(
	params *quicmemberlist.MemberlistParams,
) writeNodeValueFunc {
	return writeNodeKey(
		func(_ context.Context, _, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
			var s string
			if err := yaml.Unmarshal([]byte(value), &s); err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			i, err := strconv.Atoi(s)
			if err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			prev = params.GosshipNodes()
			if prev == i {
				return prev, nil, false, nil
			}

			if err := params.SetGosshipNodes(i); err != nil {
				return nil, nil, false, err
			}

			return prev, params.GosshipNodes(), true, nil
		})
}

func writeLocalParamMemberlistSuspicionMult(
	params *quicmemberlist.MemberlistParams,
) writeNodeValueFunc {
	return writeNodeKey(
		func(_ context.Context, _, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
			var s string
			if err := yaml.Unmarshal([]byte(value), &s); err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			i, err := strconv.Atoi(s)
			if err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			prev = params.SuspicionMult()
			if prev == i {
				return prev, nil, false, nil
			}

			if err := params.SetSuspicionMult(i); err != nil {
				return nil, nil, false, err
			}

			return prev, params.SuspicionMult(), true, nil
		})
}

func writeLocalParamMemberlistSuspicionMaxTimeoutMult(
	params *quicmemberlist.MemberlistParams,
) writeNodeValueFunc {
	return writeNodeKey(
		func(_ context.Context, _, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
			var s string
			if err := yaml.Unmarshal([]byte(value), &s); err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			i, err := strconv.Atoi(s)
			if err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			prev = params.SuspicionMaxTimeoutMult()
			if prev == i {
				return prev, nil, false, nil
			}

			if err := params.SetSuspicionMaxTimeoutMult(i); err != nil {
				return nil, nil, false, err
			}

			return prev, params.SuspicionMaxTimeoutMult(), true, nil
		})
}

func writeLocalParamMemberlistUdpBufferSize(
	params *quicmemberlist.MemberlistParams,
) writeNodeValueFunc {
	return writeNodeKey(
		func(_ context.Context, _, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
			var s string
			if err := yaml.Unmarshal([]byte(value), &s); err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			i, err := strconv.Atoi(s)
			if err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			prev = params.UDPBufferSize()
			if prev == i {
				return prev, nil, false, nil
			}

			if err := params.SetUDPBufferSize(i); err != nil {
				return nil, nil, false, err
			}

			return prev, params.UDPBufferSize(), true, nil
		})
}

func writeLocalParamMemberlistBroadcastTimerMult(
	params *quicmemberlist.MemberlistParams,
) writeNodeValueFunc {
	return writeNodeKey(
		func(_ context.Context, _, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
			var s string
			if err := yaml.Unmarshal([]byte(value), &s); err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			i, err := strconv.Atoi(s)
			if err != nil {
				return nil, nil, false, errors.WithStack(err)
			}

			prev = params.BroadcastTimerMult()
			if prev == i {
				return prev, nil, false, nil
			}

			if err := params.SetBroadcastTimerMult(i); err != nil {
				return nil, nil, false, err
			}

			return prev, params.BroadcastTimerMult(), true, nil
		})
}

func writeLocalParamMemberlistUserMsgLoopInterval(
	params *quicmemberlist.MemberlistParams,
) writeNodeValueFunc {
	return writeNodeKey(
		func(_ context.Context, _, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
			d, err := parseNodeValueDuration(value)
			if err != nil {
				return nil, nil, false, err
			}

			prev = params.UserMsgLoopInterval()
			if prev == d {
				return prev, nil, false, nil
			}

			if err := params.SetUserMsgLoopInterval(d); err != nil {
				return nil, nil, false, err
			}

			return prev, params.UserMsgLoopInterval(), true, nil
		})
}

func writeLocalParamNetworkTimeoutRequest(
	params *launch.NetworkParams,
) writeNodeValueFunc {
	return writeNodeKey(func(
		_ context.Context, _, _, value, _ string,
	) (prev, next interface{}, updated bool, _ error) {
		d, err := parseNodeValueDuration(value)
		if err != nil {
			return nil, nil, false, err
		}

		prev = params.TimeoutRequest()
		if prev == d {
			return prev, nil, false, nil
		}

		if err := params.SetTimeoutRequest(d); err != nil {
			return nil, nil, false, err
		}

		return prev, params.TimeoutRequest(), true, nil
	})
}

func writeSyncSources(
	jsonencoder encoder.Encoder,
	design launch.NodeDesign,
	syncSourceChecker *isaacnetwork.SyncSourceChecker,
) writeNodeValueFunc {
	return func(_ context.Context, _, value, _ string) (prev, next interface{}, updated bool, _ error) {
		e := util.StringError("update sync source")

		var sources *launch.SyncSourcesDesign
		if err := sources.DecodeYAML([]byte(value), jsonencoder); err != nil {
			return nil, nil, false, e.Wrap(err)
		}

		if err := launch.IsValidSyncSourcesDesign(
			sources,
			design.Network.PublishString,
			design.Network.Publish().String(),
		); err != nil {
			return nil, nil, false, e.Wrap(err)
		}

		prev = syncSourceChecker.Sources()

		if err := syncSourceChecker.UpdateSources(context.Background(), sources.Sources()); err != nil {
			return nil, nil, false, err
		}

		return prev, sources, true, nil
	}
}

func writeDiscovery(pctx context.Context) (writeNodeValueFunc, error) {
	var discoveries *util.Locked[[]quicstream.ConnInfo]

	if err := util.LoadFromContextOK(pctx,
		launch.DiscoveryContextKey, &discoveries,
	); err != nil {
		return nil, err
	}

	var aclallow launch.ACLAllowFunc

	switch i, err := launch.PACLAllowFunc(pctx); {
	case err != nil:
		return nil, err
	default:
		aclallow = i
	}

	return func(ctx context.Context, _, value, acluser string) (prev, next interface{}, updated bool, _ error) {
		if !aclallow(ctx, acluser, DiscoveryACLScope, launch.WriteAllowACLPerm, nil) {
			return nil, nil, false, launch.ErrACLAccessDenied.WithStack()
		}

		e := util.StringError("update discoveries")

		var sl []string
		if err := yaml.Unmarshal([]byte(value), &sl); err != nil {
			return nil, nil, false, e.Wrap(err)
		}

		cis := make([]quicstream.ConnInfo, len(sl))

		for i := range sl {
			if err := nutil.IsValidAddr(sl[i]); err != nil {
				return nil, nil, false, e.Wrap(err)
			}

			addr, tlsinsecure := nutil.ParseTLSInsecure(sl[i])

			ci, err := quicstream.NewConnInfoFromStringAddr(addr, tlsinsecure)
			if err != nil {
				return nil, nil, false, e.Wrap(err)
			}

			cis[i] = ci
		}

		prevd := launch.GetDiscoveriesFromLocked(discoveries)

		switch {
		case len(prevd) != len(cis):
		case len(util.Filter2Slices(prevd, cis, func(a, b quicstream.ConnInfo) bool {
			return a.String() == b.String()
		})) < 1:
			return prevd, nil, false, nil
		}

		_ = discoveries.SetValue(cis)

		return prevd, cis, true, nil
	}, nil
}

func writeLocalParamNetworkRateLimit(
	params *launch.NetworkRateLimitParams,
	param string,
) writeNodeValueFunc {
	switch param {
	case "node",
		"net",
		"suffrage",
		"default":
	default:
		panic(fmt.Sprintf("unknown key, %q for network ratelimit", param))
	}

	return func(_ context.Context, _, value, _ string) (prev, next interface{}, updated bool, err error) {
		switch param {
		case "node":
			prev = params.NodeRuleSet()
		case "net":
			prev = params.NetRuleSet()
		case "suffrage":
			prev = params.SuffrageRuleSet()
		case "default":
			prev = params.DefaultRuleMap()
		default:
			return nil, nil, false, util.ErrNotFound.Errorf("unknown key, %q for network ratelimit", param)
		}

		switch i, err := unmarshalRateLimitRule(param, value); {
		case err != nil:
			return nil, nil, false, err
		default:
			next = i
		}

		return prev, next, true, func() error {
			switch param {
			case "node":
				return params.SetNodeRuleSet(next.(launch.RateLimiterRuleSet)) //nolint:forcetypeassert //...
			case "net":
				return params.SetNetRuleSet(next.(launch.RateLimiterRuleSet)) //nolint:forcetypeassert //...
			case "suffrage":
				return params.SetSuffrageRuleSet(next.(launch.RateLimiterRuleSet)) //nolint:forcetypeassert //...
			case "default":
				return params.SetDefaultRuleMap(next.(launch.RateLimiterRuleMap)) //nolint:forcetypeassert //...
			default:
				return util.ErrNotFound.Errorf("unknown key, %q for network", param)
			}
		}()
	}
}

func writeAllowConsensus(pctx context.Context) (writeNodeValueFunc, error) {
	var states *isaacstates.States

	if err := util.LoadFromContext(pctx,
		launch.StatesContextKey, &states,
	); err != nil {
		return nil, err
	}

	var aclallow launch.ACLAllowFunc

	switch i, err := launch.PACLAllowFunc(pctx); {
	case err != nil:
		return nil, err
	default:
		aclallow = i
	}

	return func(ctx context.Context, _, value, acluser string) (prev, next interface{}, updated bool, _ error) {
		extra := zerolog.Dict().
			Str("value", value)

		if !aclallow(ctx, acluser, StatesAllowConsensusACLScope, launch.WriteAllowACLPerm, extra) {
			return nil, nil, false, launch.ErrACLAccessDenied.WithStack()
		}

		var allow bool

		if err := yaml.Unmarshal([]byte(value), &allow); err != nil {
			return nil, nil, false, errors.WithStack(err)
		}

		preva := states.AllowedConsensus()
		if preva == allow {
			return preva, nil, false, nil
		}

		if states.SetAllowConsensus(allow) {
			next = allow
		}

		return preva, next, true, nil
	}, nil
}

func writeACL(pctx context.Context) (writeNodeValueFunc, error) {
	var encs *encoder.Encoders
	var acl *launch.YAMLACL

	if err := util.LoadFromContextOK(pctx,
		launch.EncodersContextKey, &encs,
		launch.ACLContextKey, &acl,
	); err != nil {
		return nil, err
	}

	var aclallow launch.ACLAllowFunc

	switch i, err := launch.PACLAllowFunc(pctx); {
	case err != nil:
		return nil, err
	default:
		aclallow = i
	}

	return writeNodeKey(func(
		ctx context.Context, key, nextkey, value, acluser string,
	) (interface{}, interface{}, bool, error) {
		fullkey := fullKey("acl", key, nextkey)

		if len(key) > 0 {
			return nil, nil, false, errors.Errorf("unknown key, %q", fullkey)
		}

		extra := zerolog.Dict().Str("key", fullkey)

		if !aclallow(ctx, acluser, ACLACLScope, launch.WriteAllowACLPerm, extra) {
			return nil, nil, false, launch.ErrACLAccessDenied.WithStack()
		}

		prev := acl.Export()

		switch updated, err := acl.Import([]byte(value), encs.JSON()); {
		case err != nil:
			return nil, nil, false, err
		default:
			return prev, acl.Export(), updated, nil
		}
	}), nil
}

func writeBlockItemFiles(pctx context.Context) (writeNodeValueFunc, error) {
	var encs *encoder.Encoders
	var db isaac.Database
	var readers *isaac.BlockItemReaders
	var fromRemotes isaac.RemotesBlockItemReadFunc

	if err := util.LoadFromContextOK(pctx,
		launch.EncodersContextKey, &encs,
		launch.CenterDatabaseContextKey, &db,
		launch.BlockItemReadersContextKey, &readers,
		launch.RemotesBlockItemReaderFuncContextKey, &fromRemotes,
	); err != nil {
		return nil, err
	}

	var aclallow launch.ACLAllowFunc

	switch i, err := launch.PACLAllowFunc(pctx); {
	case err != nil:
		return nil, err
	default:
		aclallow = i
	}

	readitemf := isaac.BlockItemReadByBlockItemFileFuncWithRemote(readers, fromRemotes)

	return writeNodeKey(func(
		ctx context.Context, key, nextkey, value, acluser string,
	) (interface{}, interface{}, bool, error) {
		fullkey := fullKey("block_item_files", key, nextkey)

		if len(key) < 1 {
			return nil, nil, false, errors.Errorf("wrong key, %q", fullkey)
		}

		var height base.Height

		switch i, err := base.ParseHeightString(key); {
		case err != nil:
			return nil, nil, false, err
		default:
			height = i
		}

		switch found, err := readers.ItemFilesReader(height, func(io.Reader) error { return nil }); {
		case err != nil:
			return nil, nil, false, err
		case !found:
			return nil, nil, false, errors.Errorf("block item files not found")
		}

		extra := zerolog.Dict().Str("key", fullkey)

		if !aclallow(ctx, acluser, BlockItemFilesACLScope, launch.WriteAllowACLPerm, extra) {
			return nil, nil, false, launch.ErrACLAccessDenied.WithStack()
		}

		if err := isValidUploadedBlockItemFiles(
			ctx,
			encs.JSON(),
			height,
			[]byte(value),
			db.BlockMap,
			readitemf,
		); err != nil {
			return nil, nil, true, err
		}

		switch found, err := readers.WriteItemFiles(height, []byte(value)); {
		case err != nil, !found:
			return nil, nil, found, err
		default:
			return nil, nil, true, nil
		}
	}), nil
}

func isValidUploadedBlockItemFiles(
	ctx context.Context,
	jsonenc encoder.Encoder,
	height base.Height,
	b []byte,
	blockmapf func(base.Height) (base.BlockMap, bool, error),
	readitemf func(
		context.Context, base.Height, base.BlockItemType, base.BlockItemFile, isaac.BlockItemReaderCallbackFunc,
	) (bool, error),
) error {
	var bfiles base.BlockItemFiles

	switch err := encoder.Decode(jsonenc, b, &bfiles); {
	case err != nil:
		return err
	default:
		if err := bfiles.IsValid(nil); err != nil {
			return err
		}
	}

	var bm base.BlockMap

	switch i, found, err := blockmapf(height); {
	case err != nil:
		return err
	case !found:
		return util.ErrNotFound.Errorf("blockmap")
	default:
		bm = i
	}

	if err := base.IsValidBlockItemFilesWithBlockMap(bm, bfiles); err != nil {
		return err
	}

	var berr error

	bm.Items(func(item base.BlockMapItem) bool {
		var bfile base.BlockItemFile

		switch i, found := bfiles.Item(item.Type()); {
		case !found:
			berr = util.ErrNotFound.Errorf("block item file, %q", item.Type())

			return false
		default:
			bfile = i
		}

		// NOTE verify checksum
		switch found, err := readitemf(ctx, height, item.Type(), bfile, func(ir isaac.BlockItemReader) error {
			switch i, err := ir.Reader().Decompress(); {
			case err != nil:
				return err
			default:
				cr := util.NewHashChecksumReader(i, sha256.New())

				if _, err := io.ReadAll(cr); err != nil {
					return errors.Errorf("checksum reader")
				}

				if item.Checksum() != cr.Checksum() {
					return errors.Errorf("checksum does not match")
				}
			}

			return nil
		}); {
		case err != nil:
			berr = err

			return false
		case !found:
			berr = util.ErrNotFound.Errorf("%s; %v", item.Type(), bfile.URI())

			return false
		default:
			return true
		}
	})

	return berr
}

func unmarshalRateLimitRule(rule, value string) (interface{}, error) {
	var u interface{}
	if err := yaml.Unmarshal([]byte(value), &u); err != nil {
		return nil, errors.WithStack(err)
	}

	var i interface{}

	switch rule {
	case "node":
		i = launch.NodeRateLimiterRuleSet{}
	case "net":
		i = launch.NetRateLimiterRuleSet{}
	case "suffrage":
		i = &launch.SuffrageRateLimiterRuleSet{}
	case "default":
		i = launch.RateLimiterRuleMap{}
	default:
		return nil, util.ErrNotFound.Errorf("unknown prefix, %q", rule)
	}

	switch b, err := util.MarshalJSON(u); {
	case err != nil:
		return nil, err
	default:
		if err := util.UnmarshalJSON(b, &i); err != nil {
			return nil, err
		}

		if j, ok := i.(util.IsValider); ok {
			if err := j.IsValid(nil); err != nil {
				return nil, err
			}
		}

		return i, nil
	}
}

func networkHandlerNodeWrite(
	networkID base.NetworkID,
	f writeNodeValueFunc,
	eventlogger zerolog.Logger,
) quicstreamheader.Handler[launch.WriteNodeHeader] {
	handler := func(ctx context.Context, addr net.Addr,
		broker *quicstreamheader.HandlerBroker, header launch.WriteNodeHeader,
		l zerolog.Logger,
	) (sentresponse bool, value string, _ error) {
		if err := isaacnetwork.QuicstreamHandlerVerifyNode(
			ctx, addr, broker,
			header.ACLUser(), networkID,
		); err != nil {
			return false, "", err
		}

		var body io.Reader

		switch bodyType, _, b, _, res, err := broker.ReadBody(ctx); {
		case err != nil:
			return false, "", err
		case res != nil:
			return false, "", res.Err()
		case bodyType == quicstreamheader.FixedLengthBodyType,
			bodyType == quicstreamheader.StreamBodyType:
			body = b
		}

		if body != nil {
			b, err := io.ReadAll(body)
			if err != nil {
				return false, "", errors.WithStack(err)
			}

			value = string(b)
		}

		switch prev, next, updated, err := f(ctx, header.Key, value, header.ACLUser().String()); {
		case err != nil:
			return false, value, err
		default:
			l.Debug().
				Str("key", header.Key).
				Str("value", value).
				Interface("prev", prev).
				Interface("next", next).
				Bool("updated", updated).
				Msg("wrote")

			return true, value, broker.WriteResponseHeadOK(ctx, updated, nil)
		}
	}

	return func(ctx context.Context, addr net.Addr,
		broker *quicstreamheader.HandlerBroker, header launch.WriteNodeHeader,
	) (context.Context, error) {
		e := util.StringError("write node")

		l := quicstream.ConnectionLoggerFromContext(ctx, &eventlogger).With().
			Str("cid", util.UUID().String()).
			Logger()

		switch sentresponse, value, err := handler(ctx, addr, broker, header, l); {
		case err != nil:
			l.Error().Err(err).
				Str("key", header.Key).
				Str("value", value).
				Send()

			if !sentresponse {
				return ctx, e.WithMessage(broker.WriteResponseHeadOK(ctx, false, err), "write response header")
			}

			return ctx, e.Wrap(err)
		default:
			return ctx, nil
		}
	}
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

func readNodeKey(f readNodeNextValueFunc) readNodeValueFunc {
	return func(ctx context.Context, key, acluser string) (interface{}, error) {
		i := strings.SplitN(strings.TrimPrefix(key, "."), ".", 2)

		var nextkey string
		if len(i) > 1 {
			nextkey = i[1]
		}

		return f(ctx, i[0], nextkey, acluser)
	}
}

func readNode(pctx context.Context, lock *sync.RWMutex) (readNodeValueFunc, error) {
	fStates, err := readStates(pctx)
	if err != nil {
		return nil, err
	}

	fDesign, err := readDesign(pctx)
	if err != nil {
		return nil, err
	}

	fDiscovery, err := readDiscovery(pctx)
	if err != nil {
		return nil, err
	}

	fACL, err := readACL(pctx)
	if err != nil {
		return nil, err
	}

	fBlockItemFiles, err := readBlockItemFiles(pctx)
	if err != nil {
		return nil, err
	}

	return readNodeKey(func(ctx context.Context, key, nextkey, acluser string) (interface{}, error) {
		lock.RLock()
		defer lock.RUnlock()

		switch key {
		case "states":
			return fStates(ctx, nextkey, acluser)
		case "design":
			return fDesign(ctx, nextkey, acluser)
		case "discovery":
			return fDiscovery(ctx, nextkey, acluser)
		case "acl":
			return fACL(ctx, nextkey, acluser)
		case "block_item_files":
			return fBlockItemFiles(ctx, nextkey, acluser)
		default:
			return nil, util.ErrNotFound.Errorf("unknown key, %q for params", key)
		}
	}), nil
}

func readStates(pctx context.Context) (readNodeValueFunc, error) {
	fAllowConsensus, err := readAllowConsensus(pctx)
	if err != nil {
		return nil, err
	}

	return readNodeKey(func(ctx context.Context, key, nextkey, acluser string) (interface{}, error) {
		switch key {
		case "allow_consensus":
			return fAllowConsensus(ctx, nextkey, acluser)
		default:
			return nil, util.ErrNotFound.Errorf("unknown key, %q for node", key)
		}
	}), nil
}

func readDesign(pctx context.Context) (readNodeValueFunc, error) {
	var design launch.NodeDesign
	var log *logging.Logging
	var flag launch.DesignFlag
	var encs *encoder.Encoders

	if err := util.LoadFromContextOK(pctx,
		launch.DesignContextKey, &design,
		launch.LoggingContextKey, &log,
		launch.DesignFlagContextKey, &flag,
		launch.EncodersContextKey, &encs,
	); err != nil {
		return nil, err
	}

	var aclallow launch.ACLAllowFunc

	switch i, err := launch.PACLAllowFunc(pctx); {
	case err != nil:
		return nil, err
	default:
		aclallow = i
	}

	readDesignFileF := func() (*launch.NodeDesign, error) {
		return nil, errors.Errorf("design file; can not read")
	}

	switch i, err := readDesignFileToNodeReadFunc(flag, encs); {
	case err != nil:
		return nil, err
	case i == nil:
		log.Log().Warn().Stringer("design", flag.URL()).Msg("design file not readable")
	default:
		readDesignFileF = i
	}

	m := map[string]func() (interface{}, error){}

	m["_source"] = func() (interface{}, error) {
		return readDesignFileF()
	}
	m["_generated"] = func() (interface{}, error) {
		b, err := yaml.Marshal(design)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return b, nil
	}

	return func(ctx context.Context, key, acluser string) (interface{}, error) {
		extra := zerolog.Dict().
			Str("key", "design."+key)

		if !aclallow(ctx, acluser, DesignACLScope, launch.ReadAllowACLPerm, extra) {
			return nil, launch.ErrACLAccessDenied.WithStack()
		}

		switch f, found := m[key]; {
		case !found:
			return nil, util.ErrNotFound.Errorf("unknown key, %q for design", key)
		default:
			return f()
		}
	}, nil
}

func readAllowConsensus(pctx context.Context) (readNodeValueFunc, error) {
	var states *isaacstates.States

	if err := util.LoadFromContext(pctx,
		launch.StatesContextKey, &states,
	); err != nil {
		return nil, err
	}

	var aclallow launch.ACLAllowFunc

	switch i, err := launch.PACLAllowFunc(pctx); {
	case err != nil:
		return nil, err
	default:
		aclallow = i
	}

	return func(ctx context.Context, _, acluser string) (interface{}, error) {
		extra := zerolog.Dict().
			Str("key", "states.allow_consensus")

		if !aclallow(ctx, acluser, StatesAllowConsensusACLScope, launch.ReadAllowACLPerm, extra) {
			return nil, launch.ErrACLAccessDenied.WithStack()
		}

		return states.AllowedConsensus(), nil
	}, nil
}

func readDiscovery(pctx context.Context) (readNodeValueFunc, error) {
	var discoveries *util.Locked[[]quicstream.ConnInfo]

	if err := util.LoadFromContextOK(pctx,
		launch.DiscoveryContextKey, &discoveries,
	); err != nil {
		return nil, err
	}

	var aclallow launch.ACLAllowFunc

	switch i, err := launch.PACLAllowFunc(pctx); {
	case err != nil:
		return nil, err
	default:
		aclallow = i
	}

	return func(ctx context.Context, _, acluser string) (interface{}, error) {
		extra := zerolog.Dict().
			Str("key", "discovery")

		if !aclallow(ctx, acluser, DiscoveryACLScope, launch.ReadAllowACLPerm, extra) {
			return nil, launch.ErrACLAccessDenied.WithStack()
		}

		return launch.GetDiscoveriesFromLocked(discoveries), nil
	}, nil
}

func readACL(pctx context.Context) (readNodeValueFunc, error) {
	var acl *launch.YAMLACL

	if err := util.LoadFromContextOK(pctx,
		launch.ACLContextKey, &acl,
	); err != nil {
		return nil, err
	}

	var aclallow launch.ACLAllowFunc

	switch i, err := launch.PACLAllowFunc(pctx); {
	case err != nil:
		return nil, err
	default:
		aclallow = i
	}

	return readNodeKey(func(ctx context.Context, key, nextkey, acluser string) (interface{}, error) {
		fullkey := fullKey("acl", key, nextkey)

		if len(key) > 0 {
			return nil, errors.Errorf("unknown key, %q", fullkey)
		}

		extra := zerolog.Dict().Str("key", fullkey)

		if !aclallow(ctx, acluser, ACLACLScope, launch.ReadAllowACLPerm, extra) {
			return nil, launch.ErrACLAccessDenied.WithStack()
		}

		b, err := yaml.Marshal(acl.Export())

		return b, errors.WithStack(err)
	}), nil
}

func readBlockItemFiles(pctx context.Context) (readNodeValueFunc, error) {
	var readers *isaac.BlockItemReaders

	if err := util.LoadFromContextOK(pctx,
		launch.BlockItemReadersContextKey, &readers,
	); err != nil {
		return nil, err
	}

	var aclallow launch.ACLAllowFunc

	switch i, err := launch.PACLAllowFunc(pctx); {
	case err != nil:
		return nil, err
	default:
		aclallow = i
	}

	return readNodeKey(func(ctx context.Context, key, nextkey, acluser string) (interface{}, error) {
		fullkey := fullKey("block_item_files", key, nextkey)

		if len(key) < 1 {
			return nil, errors.Errorf("wrong key, %q", fullkey)
		}

		var height base.Height

		switch i, err := base.ParseHeightString(key); {
		case err != nil:
			return nil, err
		default:
			height = i
		}

		extra := zerolog.Dict().Str("key", fullkey)

		if !aclallow(ctx, acluser, BlockItemFilesACLScope, launch.ReadAllowACLPerm, extra) {
			return nil, launch.ErrACLAccessDenied.WithStack()
		}

		var b []byte

		switch found, err := readers.ItemFilesReader(height, func(r io.Reader) error {
			i, err := io.ReadAll(r)

			b = i

			return errors.WithStack(err)
		}); {
		case err != nil:
			return nil, err
		case !found:
			return nil, errors.Errorf("block item files not found")
		default:
			return b, nil
		}
	}), nil
}

func networkHandlerNodeRead(
	networkID base.NetworkID,
	f readNodeValueFunc,
	eventlogger zerolog.Logger,
) quicstreamheader.Handler[launch.ReadNodeHeader] {
	handler := func(ctx context.Context, addr net.Addr,
		broker *quicstreamheader.HandlerBroker, header launch.ReadNodeHeader,
	) (sentresponse bool, _ error) {
		if err := isaacnetwork.QuicstreamHandlerVerifyNode(
			ctx, addr, broker,
			header.ACLUser(), networkID,
		); err != nil {
			return false, err
		}

		switch body, err := func() (*bytes.Buffer, error) {
			switch v, err := f(ctx, header.Key, header.ACLUser().String()); {
			case err != nil:
				return nil, err
			default:
				if b, ok := v.([]byte); ok {
					return bytes.NewBuffer(b), nil
				}

				b, err := broker.Encoder.Marshal(v)
				if err != nil {
					return nil, err
				}

				return bytes.NewBuffer(b), nil
			}
		}(); {
		case err != nil:
			return false, err
		default:
			defer body.Reset()

			if err := broker.WriteResponseHeadOK(ctx, true, nil); err != nil {
				return true, err
			}

			return true, broker.WriteBody(ctx, quicstreamheader.StreamBodyType, 0, body)
		}
	}

	return func(ctx context.Context, addr net.Addr,
		broker *quicstreamheader.HandlerBroker, header launch.ReadNodeHeader,
	) (context.Context, error) {
		e := util.StringError("read node")

		l := quicstream.ConnectionLoggerFromContext(ctx, &eventlogger).With().
			Str("key", header.Key).
			Str("cid", util.UUID().String()).
			Logger()

		switch sentresponse, err := handler(ctx, addr, broker, header); {
		case errors.Is(err, util.ErrNotFound):
			l.Error().Err(err).Msg("key not found")

			return ctx, e.WithMessage(broker.WriteResponseHeadOK(ctx, false, nil), "write response header")
		case err != nil:
			l.Error().Err(err).Send()

			if !sentresponse {
				return ctx, e.WithMessage(broker.WriteResponseHeadOK(ctx, false, err), "write response header")
			}

			return ctx, e.Wrap(err)
		default:
			l.Debug().Msg("read")

			return ctx, nil
		}
	}
}

func writeDesignFile(
	key string, value interface{},
	read func() (*launch.NodeDesign, *digest.YamlDigestDesign, error),
	write func([]byte) error,
) error {
	var nd *launch.NodeDesign
	var ad *digest.YamlDigestDesign
	var ndb []byte
	var adb []byte
	var err error

	switch nd, ad, err = read(); {
	case err != nil:
		return err
	default:
		if ndb, err = yaml.Marshal(nd.MarshalYAML()); err != nil {
			return errors.WithStack(err)
		}

		if adb, err = yaml.Marshal(ad); err != nil {
			return errors.WithStack(err)
		}
	}

	nb, err := updateDesignMap(ndb, key, value)
	if err != nil {
		return err
	}

	var apiNode yaml.Node
	if err := yaml.Unmarshal(adb, &apiNode); err != nil {
		return errors.WithStack(err)
	}

	if len(apiNode.Content) != 0 {
		nb, err = updateDesignMap(nb, "api", apiNode.Content[0])
		if err != nil {
			return err
		}
	}

	return write(nb)
}

func readDesignFileToNodeReadFunc(
	flag launch.DesignFlag, encs *encoder.Encoders) (func() (*launch.NodeDesign, error), error) {
	var design launch.NodeDesign
	var rErr error

	switch flag.Scheme() {
	case "file":
		return func() (*launch.NodeDesign, error) {
			switch d, _, err := launch.NodeDesignFromFile(flag.URL().Path, encs.JSON()); {
			case err != nil:
				return nil, errors.WithStack(err)
			default:
				design = d
				rErr = nil
			}

			return &design, nil
		}, nil

	case "http", "https":
		return func() (*launch.NodeDesign, error) {
			switch d, _, err := launch.NodeDesignFromHTTP(flag.URL().Path, flag.Properties().HTTPSTLSInsecure, encs.JSON()); {
			case err != nil:
				return nil, errors.WithStack(err)
			default:
				design = d
			}

			return &design, errors.WithStack(rErr)
		}, nil
	case "consul":
		return func() (*launch.NodeDesign, error) {
			switch d, _, err := launch.NodeDesignFromConsul(flag.URL().Host, flag.URL().Path, encs.JSON()); {
			case err != nil:
				return nil, errors.WithStack(err)
			default:
				design = d
			}

			return &design, errors.WithStack(rErr)
		}, nil

	default:
		return nil, errors.WithStack(errors.Errorf("unknown design uri, %q", flag.URL()))
	}
}

func readDesignFileToNodeWriteFunc(
	flag launch.DesignFlag, encs *encoder.Encoders) (func() (*launch.NodeDesign, *digest.YamlDigestDesign, error), error) {
	var design launch.NodeDesign
	var yamlDesign *digest.YamlDigestDesign
	var rErr error

	switch flag.Scheme() {
	case "file":
		return func() (*launch.NodeDesign, *digest.YamlDigestDesign, error) {
			switch d, _, err := launch.NodeDesignFromFile(flag.URL().Path, encs.JSON()); {
			case err != nil:
				return nil, nil, errors.WithStack(err)
			default:
				design = d
				rErr = nil
			}

			var m struct {
				API *digest.YamlDigestDesign `yaml:"api"`
			}

			b, err := os.ReadFile(filepath.Clean(flag.URL().Path))
			if err != nil {
				return nil, nil, errors.WithStack(err)
			}

			nb, err := util.ReplaceEnvVariables(b)
			if err != nil {
				return nil, nil, errors.WithStack(err)
			}

			if err := yaml.Unmarshal(nb, &m); err != nil {
				return nil, nil, errors.WithStack(err)
			}

			yamlDesign = m.API

			return &design, yamlDesign, nil
		}, nil

	case "http", "https":
		return func() (*launch.NodeDesign, *digest.YamlDigestDesign, error) {
			switch d, _, err := launch.NodeDesignFromHTTP(flag.URL().Path, flag.Properties().HTTPSTLSInsecure, encs.JSON()); {
			case err != nil:
				return nil, nil, errors.WithStack(err)
			default:
				design = d
			}

			return &design, nil, errors.WithStack(rErr)
		}, nil
	case "consul":
		return func() (*launch.NodeDesign, *digest.YamlDigestDesign, error) {
			switch d, _, err := launch.NodeDesignFromConsul(flag.URL().Host, flag.URL().Path, encs.JSON()); {
			case err != nil:
				return nil, nil, errors.WithStack(err)
			default:
				design = d
			}

			return &design, nil, errors.WithStack(rErr)
		}, nil

	default:
		return nil, errors.WithStack(errors.Errorf("unknown design uri, %q", flag.URL()))
	}
}

func writeDesignFileFunc(flag launch.DesignFlag) (func([]byte) error, error) {
	switch flag.Scheme() {
	case "file":
		f := flag.URL().Path

		switch fi, err := os.Stat(filepath.Clean(f)); {
		case err != nil:
			return nil, errors.WithStack(err)
		case fi.Mode()&os.ModePerm == os.ModePerm:
			return nil, nil
		}

		return func(b []byte) error {
			f, err := os.OpenFile(filepath.Clean(f), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
			if err != nil {
				return errors.WithStack(err)
			}

			defer func() {
				_ = f.Close()
			}()

			_, err = f.Write(b)

			return errors.WithStack(err)
		}, nil
	case "http", "https":
		return nil, nil
	case "consul":
		return func(b []byte) error {
			return launch.WriteConsul(b, flag)
		}, nil
	default:
		return nil, errors.Errorf("unknown design uri, %q", flag.URL())
	}
}

func fullKey(keys ...string) string {
	return strings.Join(util.FilterSlice(keys, func(i string) bool {
		return len(i) > 0
	}), ".")
}
