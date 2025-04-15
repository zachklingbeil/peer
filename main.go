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

func HelloPeers(factory *factory.Factory) *Peers {
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

	go peers.HelloUniverse()
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
	p.Factory.When.Signal()
}

func (p *Peers) HelloUniverse() {
	const batchSize = 1000
	var batch []*Peer

	for {
		p.Factory.Mu.Lock()

		for len(p.Addresses) == 0 {
			p.saveBatch(&batch)
			fmt.Println("Hello Universe")
			p.Factory.When.Wait()
		}

		address := <-p.PeerChan
		peer := p.Map[address]
		p.Factory.Mu.Unlock()
		p.processPeer(peer)
		batch = append(batch, peer)
		fmt.Printf("%d %s %s %d\n", len(p.Addresses), peer.ENS, peer.LoopringENS, peer.LoopringID)

		if len(batch) >= batchSize {
			p.saveBatch(&batch)
		}
	}
}

func (p *Peers) processPeer(peer *Peer) {
	p.GetENS(peer, peer.Address)
	p.GetLoopringENS(peer, peer.Address)
	p.GetLoopringID(peer, peer.Address)
}

func (p *Peers) saveBatch(batch *[]*Peer) {
	if len(*batch) > 0 {
		if err := p.SavePeers(*batch); err != nil {
			fmt.Printf("Error saving batch: %v\n", err)
		}
		*batch = (*batch)[:0]
	}
}

func (p *Peers) GetPeer(key any, value uint) any {
	p.Factory.Mu.Lock()
	defer p.Factory.Mu.Unlock()

	var peer *Peer

	switch id := key.(type) {
	case string:
		peer = p.Map[id]
		if peer == nil {
			for _, v := range p.Map {
				if v.ENS == id || v.LoopringENS == id {
					peer = v
					break
				}
			}
		}
	case int64:
		for _, v := range p.Map {
			if v.LoopringID == id {
				peer = v
				break
			}
		}
	default:
		fmt.Printf("Unsupported key type: %T\n", key)
		return nil
	}

	if peer == nil {
		fmt.Printf("Peer %v not found\n", key)
		return nil
	}

	switch value {
	case 0:
		return peer
	case 1:
		return peer.Address
	case 2:
		return peer.ENS
	case 3:
		return peer.LoopringENS
	case 4:
		return peer.LoopringID
	default:
		fmt.Printf("Unsupported value: %d\n", value)
		return nil
	}
}

func (p *Peers) GetAddressByLoopringID(loopringID int64) string {
	p.Factory.Mu.Lock()
	defer p.Factory.Mu.Unlock()

	for _, peer := range p.Map {
		if peer.LoopringID == loopringID {
			return peer.Address
		}
	}

	fmt.Printf("No peer found with LoopringID: %d\n", loopringID)
	return ""
}
