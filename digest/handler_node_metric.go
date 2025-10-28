package digest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
	"github.com/ProtoconNet/mitum2/network/quicmemberlist"
	"github.com/ProtoconNet/mitum2/network/quicstream"
	quicstreamheader "github.com/ProtoconNet/mitum2/network/quicstream/header"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (hd *Handlers) handleNodeMetric(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	self := ParseBoolQuery(r.URL.Query().Get("self"))

	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		i, err := hd.handleNodeMetricInGroup(self)

		return i, err
	}); err != nil {
		hd.Log().Err(err).Msg("get node metric")

		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)

		if !shared {
			HTTP2WriteCache(w, cacheKey, hd.expireShortLived)
		}
	}
}

func (hd *Handlers) handleNodeMetricInGroup(self bool) (interface{}, error) {
	connectionPool, memberList, nodeList, err := hd.client()
	client := isaacnetwork.NewBaseClient( //nolint:gomnd //...
		hd.encs, hd.enc,
		connectionPool.Dial,
		connectionPool.CloseAll,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = client.Close()
	}()

	var nodeMetricList []map[string]interface{}
	connInfo := make(map[string]quicstream.ConnInfo)

	if !self {
		memberList.Members(func(node quicmemberlist.Member) bool {
			connInfo[node.ConnInfo().String()] = node.ConnInfo()
			return true
		})
		for _, c := range nodeList {
			connInfo[c.String()] = c
		}
	} else {
		connInfo[hd.node.String()] = hd.node
	}
	for i := range connInfo {
		nodeMetric, err := NodeMetric(client, connInfo[i])

		if err != nil {
			continue
		}

		nm := map[string]interface{}{"node-metric": nodeMetric, "conn-info": connInfo[i]}

		nodeMetricList = append(nodeMetricList, nm)
	}

	if i, err := hd.buildNodeMetricHal(nodeMetricList); err != nil {
		return nil, err
	} else {
		return hd.enc.Marshal(i)
	}
}

func (hd *Handlers) buildNodeMetricHal(ni []map[string]interface{}) (Hal, error) {
	var hal Hal = NewBaseHal(ni, NewHalLink(HandlerPathNodeMetric, nil))

	return hal, nil
}

func NodeMetric(client *isaacnetwork.BaseClient, connInfo quicstream.ConnInfo) (*isaacnetwork.NodeMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	stream, _, err := client.Dial(ctx, connInfo)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = client.Close()
	}()

	header := isaacnetwork.NewNodeMetricsRequestHeader("1m")

	var nodeMetric *isaacnetwork.NodeMetrics
	err = stream(ctx, func(ctx context.Context, broker *quicstreamheader.ClientBroker) error {
		if err := broker.WriteRequestHead(ctx, header); err != nil {
			return err
		}

		var enc encoder.Encoder

		switch rEnc, rh, err := broker.ReadResponseHead(ctx); {
		case err != nil:
			return err
		case !rh.OK():
			return errors.Errorf("Not ok")
		case rh.Err() != nil:
			return rh.Err()
		default:
			enc = rEnc
		}

		switch bodyType, bodyLength, r, err := broker.ReadBodyErr(ctx); {
		case err != nil:
			return err
		case bodyType == quicstreamheader.EmptyBodyType,
			bodyType == quicstreamheader.FixedLengthBodyType && bodyLength < 1:
			return errors.Errorf("Empty body")
		default:
			var v interface{}
			if err := enc.StreamDecoder(r).Decode(&v); err != nil {
				return err
			}

			b, err := enc.Marshal(v)
			if err != nil {
				return err
			}

			h, err := enc.Decode(b)
			if err != nil {
				fmt.Println(err)
				return err
			}

			ni, ok := h.(isaacnetwork.NodeMetrics)
			if !ok {
				return errors.Errorf("expected isaacnetwork.NodeMetrics, not %T", v)
			}

			nodeMetric = &ni

			return nil
		}
	})
	if err != nil {
		return nil, err
	}

	return nodeMetric, nil
}
