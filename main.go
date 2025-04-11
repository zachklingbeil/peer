package peer

import (
	"fmt"

	_ "github.com/lib/pq"

	"github.com/zachklingbeil/factory"
)

type Peers struct {
	Factory   *factory.Factory
	Map       map[string]*Peer
	Addresses []string
	PeerChan  chan string
}

type Peer struct {
	Address     string `json:"address"`
	ENS         string `json:"ens"`
	LoopringENS string `json:"loopringEns"`
	LoopringID  int64  `json:"loopringId"`
}

func NewPeers(factory *factory.Factory) *Peers {
	peers := &Peers{
		Factory:   factory,
		Map:       make(map[string]*Peer),
		Addresses: nil,
	}

	if err := peers.LoadPeers(); err != nil {
		fmt.Printf("Error loading peers: %v\n", err)
	}

	peers.PeerChan = make(chan string, len(peers.Addresses))
	for _, address := range peers.Addresses {
		peers.PeerChan <- address
	}

	return peers
}

func (p *Peers) NewBlock(addresses []string) {
	p.Factory.Mu.Lock()
	defer p.Factory.Mu.Unlock()

	fmt.Printf("%d new peers\n", len(addresses))

	for _, address := range addresses {
		if _, exists := p.Map[address]; !exists {
			p.Map[address] = &Peer{Address: address}
			p.Addresses = append(p.Addresses, address)
			p.PeerChan <- address
		}
	}
}

func (p *Peers) HelloUniverse() {
	const batchSize = 1000
	var batch []*Peer

	for {
		p.Factory.Rw.RLock()
		peers := len(p.Addresses)
		p.Factory.Rw.RUnlock()

		if peers == 0 && len(batch) == 0 {
			break
		}

		if len(batch) >= batchSize || (peers == 0 && len(batch) > 0) {
			if err := p.SavePeers(batch); err != nil {
				fmt.Printf("Error saving batch: %v\n", err)
			}
			batch = batch[:0]
		}

		if peers > 0 {
			address := <-p.PeerChan

			p.Factory.Mu.Lock()
			peer := p.Map[address]
			p.Factory.Mu.Unlock()

			p.GetENS(peer, peer.Address)
			p.GetLoopringENS(peer, peer.Address)
			p.GetLoopringID(peer, peer.Address)
			batch = append(batch, peer)

			peers--
			fmt.Printf("%d %s %s %d\n", peers, peer.ENS, peer.LoopringENS, peer.LoopringID)
		}
	}
	fmt.Println("Hello Universe")
}
