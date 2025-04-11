package peer

import (
	"fmt"
	"strings"
)

func (p *Peers) LoadPeers() error {
	query := `
        SELECT address, ens, loopringEns, loopringId FROM peers
    `
	rows, err := p.Factory.Db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to load peers from database: %w", err)
	}
	defer rows.Close()

	p.Factory.Mu.Lock()

	defer p.Factory.Mu.Unlock()
	var addresses []string
	for rows.Next() {
		var peer Peer
		if err := rows.Scan(&peer.Address, &peer.ENS, &peer.LoopringENS, &peer.LoopringID); err != nil {
			return fmt.Errorf("failed to scan peer row: %w", err)
		}

		p.Map[peer.Address] = &peer
		if isIncompletePeer(peer) {
			addresses = append(addresses, peer.Address)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over peer rows: %w", err)
	}
	p.Addresses = addresses
	fmt.Printf("%d peers\n", len(p.Map))
	return nil
}

// isIncompletePeer checks if a peer has incomplete data
func isIncompletePeer(peer Peer) bool {
	return peer.ENS == "" || peer.ENS == "!" ||
		peer.LoopringENS == "" || peer.LoopringENS == "!" ||
		peer.LoopringID == -1
}

func (p *Peers) SavePeers(batch []*Peer) error {
	if len(batch) == 0 {
		return nil
	}

	queryTemplate := `
    INSERT INTO peers (address, ens, loopringEns, loopringId)
    VALUES %s
    ON CONFLICT (address) DO UPDATE SET
        ens = EXCLUDED.ens,
        loopringEns = EXCLUDED.loopringEns,
        loopringId = EXCLUDED.loopringId
    `

	placeholders := make([]string, len(batch))
	values := make([]any, 0, len(batch)*4)

	for i, peer := range batch {
		placeholders[i] = fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4)
		values = append(values, peer.Address, peer.ENS, peer.LoopringENS, peer.LoopringID)
	}

	query := fmt.Sprintf(queryTemplate, strings.Join(placeholders, ", "))
	_, err := p.Factory.Db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to save peers batch: %w", err)
	}

	fmt.Printf("%d peers saved to the database\n", len(batch))
	return nil
}
