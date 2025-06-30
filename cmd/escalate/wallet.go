package main

import (
	"encoding/json"

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
