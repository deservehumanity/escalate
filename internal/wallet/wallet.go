package wallet

import (
	"errors"
	"escalate/internal/client"
	"log"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tyler-smith/go-bip39"
)

type Wallet struct {
	Mnemo        string
	CurrentIndex uint32
	Addresses    []Address
	client       *client.HttpWalletClient
}

func New(c *client.HttpWalletClient) (*Wallet, error) {
	if c == nil {
		return nil, errors.New("wallet client is nil")
	}

	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		Mnemo:        mnemonic,
		CurrentIndex: 0,
		Addresses:    []Address{},
		client:       c,
	}, nil
}

func NewFromData(data []byte, c *client.HttpWalletClient) (Wallet, error) {
	w, err := Deserialize(data)
	if err != nil {
		return w, err
	}

	w.client = c

	return w, nil
}

func (w *Wallet) GetClient() *client.HttpWalletClient {
	return w.client
}

func (w *Wallet) DeriveAddress() (*Address, error) {
	seed := bip39.NewSeed(w.Mnemo, "")
	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	path := []uint32{
		84 + hdkeychain.HardenedKeyStart,
		0 + hdkeychain.HardenedKeyStart,
		0 + hdkeychain.HardenedKeyStart,
		0,
		0 + w.CurrentIndex,
	}

	key := master
	for _, p := range path {
		key, err = key.Child(p)
		if err != nil {
			return nil, err
		}
	}

	pubKey, err := key.ECPubKey()
	if err != nil {
		return nil, err
	}

	pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())

	addr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	addrStr := addr.EncodeAddress()

	address := NewAddress(addrStr)

	w.Addresses = append(w.Addresses, address)
	w.CurrentIndex++

	return &address, nil
}

func (w *Wallet) GetBalance() (int64, error) {
	var balance int64 = 0

	for idx, address := range w.Addresses {

		bIsOlderThanDay := time.Since(address.UpdatedAt) >= time.Hour*24
		if bIsOlderThanDay {
			stats, err := w.client.GetAddressStats(address.Addr)
			if err != nil {
				log.Printf("failed to fetch data for %s", address.Addr)
				continue
			}

			// Update wallet structure
			w.Addresses[idx].Chain = stats.ChainStats
			w.Addresses[idx].Pool = stats.PoolStats
			w.Addresses[idx].UpdatedAt = time.Now()

			balance += stats.ChainStats.FundedTxoSum - stats.ChainStats.SpentTxoSum
		} else {
			balance += address.Chain.FundedTxoSum - address.Chain.SpentTxoSum
		}
	}

	return balance, nil
}
