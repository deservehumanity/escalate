package wallet

import (
	"escalate/internal/client"
	"time"
)

type Address struct {
	Addr      string            `json:"addr"`
	Chain     client.ChainStats `json:"chain"`
	Pool      client.PoolStats  `json:"pool"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func NewAddress(addr string) Address {
	return Address{
		Addr:      addr,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (a Address) String() string {
	return ""
}
