package node

import (
	"context"
	"fedSharing/mainchain/log"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// 节点发现的通告结构体，继承Notifee
type discoveryNotifee struct {
	h host.Host
}

// setupDiscovery
// @Description: 设置mDNS节点发现服务
// @param node
// @param dst
// @return error
func setupDiscovery(node host.Host, dst string) error {
	disc := mdns.NewMdnsService(node, dst)
	n := discoveryNotifee{h: node}
	disc.RegisterNotifee(&n)
	return nil
}

// HandlePeerFound
// @Description: 节点发现后的处理函数：自动链接节点
// @receiver n
// @param pi
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID != n.h.ID() {
		log.Logger.Infof("discovered new peer %s", pi.ID.Pretty())
		err := n.h.Connect(context.Background(), pi)
		if err != nil {
			log.Logger.Warnf("error connecting to peer %s: %s", pi.ID.Pretty(), err)
		}
		log.Logger.Infof("success connect new peer %s !", pi.ID.Pretty())
		connectPeerChan <- true
	}
}
