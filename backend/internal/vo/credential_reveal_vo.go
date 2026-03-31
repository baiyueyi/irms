package vo

type CredentialRevealVO struct {
	ID             uint64  `json:"id"`
	Secret         *string `json:"secret,omitempty"`
	CertificatePEM *string `json:"certificate_pem,omitempty"`
	PrivateKeyPEM  *string `json:"private_key_pem,omitempty"`
	Passphrase     *string `json:"passphrase,omitempty"`
}
