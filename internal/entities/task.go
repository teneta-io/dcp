package entities

import (
	"crypto/rsa"
	"time"
)

const (
	TaskStatusNew         Status = "new"
	TaskStatusInitialized Status = "initialized"
	TaskStatusRunning     Status = "running"
	TaskStatusFinished    Status = "finished"
)

type TaskPayload struct {
	Link         string       `json:"link"`
	PriceLimit   uint         `json:"price_limit"`
	Cost         uint         `json:"cost"`
	Requirements Requirements `json:"requirements"`
	ExpiredAt    time.Time    `json:"expired_at"`
}

type Task struct {
	Payload      *TaskPayload   `json:"payload"`
	Status       Status         `json:"status"`
	DCCSign      []byte         `json:"dcc_sign"`
	DCCPublicKey *rsa.PublicKey `json:"dcc_public_key"`
	DCPSign      string         `json:"dcp_sign"`
	DCPPublicKey string         `json:"dcp_public_key"`
	CreatedAt    time.Time      `json:"created_at"`
}

type Status string

type Requirements struct {
	VCPU    uint8 `json:"vcpu"`
	RAM     uint8 `json:"ram"`
	Storage uint8 `json:"storage"`
	Network uint8 `json:"network"`
	GPU     uint8 `json:"gpu"`
}
