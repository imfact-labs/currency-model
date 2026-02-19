package api

import (
	"github.com/imfact-labs/mitum2/network/quicmemberlist"
	"github.com/imfact-labs/mitum2/network/quicstream"
)

func (hd *Handlers) SetNetworkClientFunc(f func() (*quicstream.ConnectionPool, *quicmemberlist.Memberlist, []quicstream.ConnInfo, error)) *Handlers {
	hd.client = f
	return hd
}
