package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HttpWalletClient struct {
}

type AddressStats struct {
	Address    string     `json:"address"`
	ChainStats ChainStats `json:"chain_stats"`
	PoolStats  PoolStats  `json:"mempool_stats"`
}

func NewHttpClient() (HttpWalletClient, error) {
	return HttpWalletClient{}, nil
}

func (c HttpWalletClient) GetAddressStats(addr string) (AddressStats, error) {
	url := fmt.Sprintf("https://blockstream.info/api/address/%s", addr)

	res, err := http.Get(url)
	if err != nil {
		return AddressStats{}, err
	}
	defer res.Body.Close()

	var stats AddressStats
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return AddressStats{}, err
	}

	return stats, nil
}
