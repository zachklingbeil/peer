package peer

import "fmt"

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
