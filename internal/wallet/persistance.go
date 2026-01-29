package wallet

import "encoding/json"

func (w *Wallet) Serialize() ([]byte, error) {
	return json.MarshalIndent(w, "", "\t")
}

func Deserialize(data []byte) (Wallet, error) {
	var w Wallet
	if err := json.Unmarshal(data, &w); err != nil {
		return Wallet{}, err
	}

	return w, nil
}
