package requests

import "time"

type TaskRequest struct {
	Link         string `json:"link"`
	PriceLimit   uint   `json:"price_limit"`
	PublicKey    string `json:"public_key"`
	PrivateKey   string `json:"private_key"`
	Requirements struct {
		VCPU    uint8 `json:"vcpu"`
		RAM     uint8 `json:"ram"`
		Storage uint8 `json:"storage"`
		Network uint8 `json:"network"`
	} `json:"requirements"`
	ExpiredAt time.Time `json:"expired_at"`
}
