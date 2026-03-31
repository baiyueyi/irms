package request

type HostCredentialCreateRequest struct {
	HostID         uint64 `json:"host_id" binding:"required"`
	AccountName    string `json:"account_name" binding:"required"`
	CredentialName string `json:"credential_name" binding:"required"`
	CredentialKind string `json:"credential_kind" binding:"required"`
	Username       string `json:"username"`
	Secret         string `json:"secret"`
	CertificatePEM string `json:"certificate_pem"`
	PrivateKeyPEM  string `json:"private_key_pem"`
	Passphrase     string `json:"passphrase"`
	Status         string `json:"status"`
	Description    string `json:"description"`
}

type HostCredentialUpdateRequest struct {
	HostID         uint64 `json:"host_id" binding:"required"`
	AccountName    string `json:"account_name" binding:"required"`
	CredentialName string `json:"credential_name" binding:"required"`
	CredentialKind string `json:"credential_kind" binding:"required"`
	Username       string `json:"username"`
	Secret         string `json:"secret"`
	CertificatePEM string `json:"certificate_pem"`
	PrivateKeyPEM  string `json:"private_key_pem"`
	Passphrase     string `json:"passphrase"`
	Status         string `json:"status"`
	Description    string `json:"description"`
}

type ServiceCredentialCreateRequest struct {
	ServiceID      uint64 `json:"service_id" binding:"required"`
	AccountName    string `json:"account_name" binding:"required"`
	CredentialName string `json:"credential_name" binding:"required"`
	CredentialKind string `json:"credential_kind" binding:"required"`
	Username       string `json:"username"`
	Secret         string `json:"secret"`
	CertificatePEM string `json:"certificate_pem"`
	PrivateKeyPEM  string `json:"private_key_pem"`
	Passphrase     string `json:"passphrase"`
	Status         string `json:"status"`
	Description    string `json:"description"`
}

type ServiceCredentialUpdateRequest struct {
	ServiceID      uint64 `json:"service_id" binding:"required"`
	AccountName    string `json:"account_name" binding:"required"`
	CredentialName string `json:"credential_name" binding:"required"`
	CredentialKind string `json:"credential_kind" binding:"required"`
	Username       string `json:"username"`
	Secret         string `json:"secret"`
	CertificatePEM string `json:"certificate_pem"`
	PrivateKeyPEM  string `json:"private_key_pem"`
	Passphrase     string `json:"passphrase"`
	Status         string `json:"status"`
	Description    string `json:"description"`
}
