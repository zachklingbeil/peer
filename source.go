package peer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/wealdtech/go-ens/v3"
)

func (p *Peers) Format(address string) string {
	address = strings.ToLower(address)
	if strings.HasPrefix(address, "0x") || strings.HasSuffix(address, ".eth") {
		return address
	}
	return address
}

// ENS -> hex
func (p *Peers) GetAddress(peer *Peer, dotEth string) {
	address, err := ens.Resolve(p.Factory.Eth, dotEth)
	if err != nil {
		peer.Address = dotEth
		return
	}
	peer.Address = p.Format(address.Hex())
}

// hex -> ENS [.eth] or "."
func (p *Peers) GetENS(peer *Peer, address string) {
	addr := common.HexToAddress(address)
	if ensName, err := ens.ReverseResolve(p.Factory.Eth, addr); err != nil || ensName == "" {
		peer.ENS = "."
	} else {
		peer.ENS = p.Format(ensName)
	}
}

// hex -> LoopringENS [.loopring.eth] or "."
func (p *Peers) GetLoopringENS(peer *Peer, address string) {
	url := fmt.Sprintf("https://api3.loopring.io/api/wallet/v3/resolveName?owner=%s", address)
	var response struct {
		Loopring string `json:"data"`
	}
	data, err := p.Factory.Json.In(url, p.LoopringApiKey)
	if err != nil || json.Unmarshal(data, &response) != nil || response.Loopring == "" {
		peer.LoopringENS = "."
		return
	}
	peer.LoopringENS = p.Format(response.Loopring)
}

// hex -> LoopringId or -1
func (p *Peers) GetLoopringID(peer *Peer, address string) {
	const maxRetries = 3
	url := fmt.Sprintf("https://api3.loopring.io/api/v3/account?owner=%s", address)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		var response struct {
			ID int64 `json:"accountId"`
		}

		// Validate the address format before making the request
		if !common.IsHexAddress(address) {
			fmt.Printf("Invalid address format: %s\n", address)
			peer.LoopringID = -1
			return
		}

		data, err := p.Factory.Json.In(url, p.LoopringApiKey)
		if err != nil {
			fmt.Printf("Attempt %d: Failed to fetch LoopringID for address %s (error: %v)\n", attempt, address, err)
			continue
		}

		// Check for unexpected status codes or empty responses
		if json.Unmarshal(data, &response) != nil || response.ID == 0 {
			fmt.Printf("Attempt %d: Unexpected response for address %s: %s\n", attempt, address, string(data))
			continue
		}

		peer.LoopringID = response.ID
		return
	}

	// If all retries fail, assign -1 and log the failure
	fmt.Printf("Failed to fetch LoopringID for address %s after %d attempts\n", address, maxRetries)
	peer.LoopringID = -1
}

// LoopringId -> hex
func (p *Peers) GetLoopringAddress(peer *Peer, id string) {
	accountID, err := strconv.Atoi(id)
	if err != nil {
		return
	}
	url := fmt.Sprintf("https://api3.loopring.io/api/v3/account?accountId=%d", accountID)
	var response struct {
		Address string `json:"owner"`
	}
	if data, err := p.Factory.Json.In(url, p.LoopringApiKey); err == nil && json.Unmarshal(data, &response) == nil {
		peer.Address = p.Format(response.Address)
	} else {
		peer.Address = "!"
	}
}
