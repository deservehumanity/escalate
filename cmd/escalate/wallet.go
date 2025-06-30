package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tyler-smith/go-bip39"
)

type Wallet struct {
	Mnemonic  string   `json:"mnemonic"`
	Index     uint32   `json:"index"`
	Addresses []string `json:"addresses"`
}

func New() (*Wallet, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		Mnemonic:  mnemonic,
		Index:     0,
		Addresses: []string{},
	}, nil
}

func (w *Wallet) Serialize() ([]byte, error) {
	return json.MarshalIndent(w, "", "	")
}

func Deserialize(data []byte) (*Wallet, error) {
	var w Wallet
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, err
	}

	return &w, nil
}

func (w *Wallet) DeriveAddress() (string, error) {
	seed := bip39.NewSeed(w.Mnemonic, "")
	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}

	path := []uint32{
		84 + hdkeychain.HardenedKeyStart,
		0 + hdkeychain.HardenedKeyStart,
		0 + hdkeychain.HardenedKeyStart,
		0,
		0 + w.Index,
	}

	key := master
	for _, p := range path {
		key, err = key.Child(p)
		if err != nil {
			return "", err
		}
	}

	pubKey, err := key.ECPubKey()
	if err != nil {
		return "", err
	}

	pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())

	addr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}

	address := addr.EncodeAddress()

	w.Addresses = append(w.Addresses, address)
	w.Index++

	return address, nil
}

func (w *Wallet) ForgetAllButFirst() {
	w.Addresses = w.Addresses[:1]
}

type AddressStats struct {
	Address      string       `json:"address"`
	ChainStats   ChainStats   `json:"chain_stats"`
	MempoolStats MempoolStats `json:"mempool_stats"`
}

type ChainStats struct {
	FundedTxoCount int64 `json:"funded_txo_count"`
	FundedTxoSum   int64 `json:"funded_txo_sum"`
	SpentTxoCount  int64 `json:"spent_txo_count"`
	SpentTxoSum    int64 `json:"spent_txo_sum"`
	TxCount        int64 `json:"tx_count"`
}

type MempoolStats struct {
	FundedTxoCount int64 `json:"funded_txo_count"`
	FundedTxoSum   int64 `json:"funded_txo_sum"`
	SpentTxoCount  int64 `json:"spent_txo_count"`
	SpentTxoSum    int64 `json:"spent_txo_sum"`
	TxCount        int64 `json:"tx_count"`
}

func (w *Wallet) GetBalance() (int64, error) {
	var balance int64 = 0

	for _, address := range w.Addresses {
		stats, err := getAddressStats(address)
		if err != nil {
			return 0, err
		}

		balance += stats.ChainStats.FundedTxoSum - stats.ChainStats.SpentTxoSum
	}

	return balance, nil
}

func getAddressStats(address string) (*AddressStats, error) {
	url := fmt.Sprintf("https://blockstream.info/api/address/%s", address)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var stats AddressStats
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return &stats, nil
}
