package vo

import "time"

type CredentialVO struct {
	ID             uint64    `json:"id"`
	HostID         uint64    `json:"host_id"`
	ServiceID      uint64    `json:"service_id"`
	AccountName    string    `json:"account_name"`
	CredentialName string    `json:"credential_name"`
	CredentialKind string    `json:"credential_kind"`
	Username       string    `json:"username"`
	Status         string    `json:"status"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
