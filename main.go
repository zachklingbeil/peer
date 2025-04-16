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
	return peers
}

func (p *Peers) NewBlock(addresses []string) {
	p.Factory.Mu.Lock()
	defer p.Factory.Mu.Unlock()

	fmt.Printf("Adding %d new peers\n", len(addresses))

	for _, address := range addresses {
		if _, exists := p.Map[address]; !exists {
			p.Map[address] = &Peer{Address: address}
			p.Addresses = append(p.Addresses, address)
			p.PeerChan <- address
			fmt.Printf("Added peer: %s\n", address)
		}
	}
	fmt.Printf("Total peers in Addresses: %d\n", len(p.Addresses))
	fmt.Println("Signal sent to process peers")
	p.Factory.When.Signal()
}

func (p *Peers) HelloUniverse() {
	const batchSize = 1000
	var batch []*Peer

	for {
		p.Factory.Mu.Lock()

		for len(p.Addresses) == 0 {
			p.saveBatch(&batch)
			fmt.Println("Hello Universe: Waiting for new peers...")
			p.Factory.When.Wait()
			fmt.Println("Hello Universe: Signal received, processing peers...")
		}

		address := <-p.PeerChan
		peer := p.Map[address]
		p.Factory.Mu.Unlock()

		p.processPeer(peer)
		batch = append(batch, peer)
		fmt.Printf("Processed peer: %s, Remaining addresses: %d\n", address, len(p.Addresses))

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
		fmt.Printf("Saving batch of %d peers\n", len(*batch))
		if err := p.SavePeers(*batch); err != nil {
			fmt.Printf("Error saving batch: %v\n", err)
		}
		*batch = (*batch)[:0]
	}
}
