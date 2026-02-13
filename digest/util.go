package digest

import (
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
	"github.com/ProtoconNet/mitum2/util"
)

var (
	ErrBadRequest      = util.NewIDError("bad request")
	UnknownProblemJSON []byte
)

func IsAccountState(st base.State) (types.Account, bool, error) {
	if !currency.IsAccountStateKey(st.Key()) {
		return types.Account{}, false, nil
	}

	ac, err := currency.LoadAccountStateValue(st)
	if err != nil {
		return types.Account{}, false, err
	}
	return *ac, true, nil
}

func IsBalanceState(st base.State) (types.Amount, bool, error) {
	if !currency.IsBalanceStateKey(st.Key()) {
		return types.Amount{}, false, nil
	}

	am, err := currency.StateBalanceValue(st)
	if err != nil {
		return types.Amount{}, false, err
	}
	return am, true, nil
}

type NodeInfoHandler func() (isaacnetwork.NodeInfo, error)

type NodeMetricHandler func() (isaacnetwork.NodeMetrics, error)
