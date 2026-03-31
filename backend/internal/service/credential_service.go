package service

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"irms/backend/internal/config"
	"irms/backend/internal/dto/request"
	"irms/backend/internal/model"
	ecode "irms/backend/internal/pkg/errors"
	"irms/backend/internal/query"
	"irms/backend/internal/repository"
	"irms/backend/internal/store"
	"irms/backend/internal/vo"

	"gorm.io/gorm"
)

type CredentialService struct {
	db        *gorm.DB
	q         *query.Query
	grantRepo *repository.GrantRepository
	permDefs  *PermissionDefinitionService
	crypto    *credentialCrypto
	auditSvc  *AuditService
}

func NewCredentialService(cfg config.Config, st *store.Store) *CredentialService {
	return &CredentialService{
		db:        st.Gorm,
		q:         st.Query,
		grantRepo: repository.NewGrantRepository(st.Gorm),
		permDefs:  NewPermissionDefinitionService(st.Query),
		crypto:    newCredentialCrypto(cfg.CredentialEncryptionKey),
		auditSvc:  NewAuditService(st.Gorm),
	}
}

func (s *CredentialService) ListHostCredentials(ctx context.Context, actor Actor, hostID uint64, page int, pageSize int) ([]vo.CredentialVO, int, error) {
	if err := s.authorizeCredentialAction(ctx, actor, "host", hostID, "list"); err != nil {
		return nil, 0, err
	}
	qhc := s.q.HostCredential
	items, total64, err := qhc.WithContext(ctx).
		Where(qhc.HostID.Eq(hostID)).
		Order(qhc.ID.Desc()).
		FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]vo.CredentialVO, 0, len(items))
	for _, it := range items {
		out = append(out, toHostCredentialVO(*it))
	}
	return out, int(total64), nil
}

func (s *CredentialService) CreateHostCredential(ctx context.Context, actor Actor, req request.HostCredentialCreateRequest) (uint64, error) {
	if err := s.authorizeCredentialAction(ctx, actor, "host", req.HostID, "write"); err != nil {
		return 0, err
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	now := time.Now()
	it := model.HostCredential{
		HostID:         req.HostID,
		AccountName:    strings.TrimSpace(req.AccountName),
		CredentialName: strings.TrimSpace(req.CredentialName),
		CredentialKind: strings.TrimSpace(req.CredentialKind),
		Status:         status,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if strings.TrimSpace(req.Username) != "" {
		v := strings.TrimSpace(req.Username)
		it.Username = &v
	}
	if strings.TrimSpace(req.Description) != "" {
		v := strings.TrimSpace(req.Description)
		it.Description = &v
	}
	if strings.TrimSpace(req.Secret) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.Secret)
		if err != nil {
			return 0, err
		}
		it.SecretCiphertext = &ciphertext
	}
	if strings.TrimSpace(req.CertificatePEM) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.CertificatePEM)
		if err != nil {
			return 0, err
		}
		it.CertificatePemCiphertext = &ciphertext
	}
	if strings.TrimSpace(req.PrivateKeyPEM) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.PrivateKeyPEM)
		if err != nil {
			return 0, err
		}
		it.PrivateKeyPemCiphertext = &ciphertext
	}
	if strings.TrimSpace(req.Passphrase) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.Passphrase)
		if err != nil {
			return 0, err
		}
		it.PassphraseCiphertext = &ciphertext
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := qtx.HostCredential.WithContext(ctx).Create(&it); err != nil {
			return err
		}
		after := struct {
			HostID            uint64 `json:"host_id"`
			AccountName       string `json:"account_name"`
			CredentialName    string `json:"credential_name"`
			CredentialKind    string `json:"credential_kind"`
			Status            string `json:"status"`
			HasSecret         bool   `json:"has_secret"`
			HasCertificatePEM bool   `json:"has_certificate_pem"`
			HasPrivateKeyPEM  bool   `json:"has_private_key_pem"`
			HasPassphrase     bool   `json:"has_passphrase"`
		}{
			HostID:            req.HostID,
			AccountName:       strings.TrimSpace(req.AccountName),
			CredentialName:    strings.TrimSpace(req.CredentialName),
			CredentialKind:    strings.TrimSpace(req.CredentialKind),
			Status:            status,
			HasSecret:         strings.TrimSpace(req.Secret) != "",
			HasCertificatePEM: strings.TrimSpace(req.CertificatePEM) != "",
			HasPrivateKeyPEM:  strings.TrimSpace(req.PrivateKeyPEM) != "",
			HasPassphrase:     strings.TrimSpace(req.Passphrase) != "",
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "create_host_credential", "host_credentials", strconv.FormatUint(it.ID, 10), req.CredentialName, nil, after)
	}); err != nil {
		return 0, err
	}
	return it.ID, nil
}

func (s *CredentialService) UpdateHostCredential(ctx context.Context, actor Actor, id uint64, req request.HostCredentialUpdateRequest) error {
	qhc := s.q.HostCredential
	old, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if err := s.authorizeCredentialAction(ctx, actor, "host", old.HostID, "write"); err != nil {
		return err
	}
	if req.HostID != old.HostID {
		if err := s.authorizeCredentialAction(ctx, actor, "host", req.HostID, "write"); err != nil {
			return err
		}
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	username := strings.TrimSpace(req.Username)
	description := strings.TrimSpace(req.Description)
	var secretCiphertext *string
	if strings.TrimSpace(req.Secret) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.Secret)
		if err != nil {
			return err
		}
		secretCiphertext = &ciphertext
	}
	var certCiphertext *string
	if strings.TrimSpace(req.CertificatePEM) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.CertificatePEM)
		if err != nil {
			return err
		}
		certCiphertext = &ciphertext
	}
	var keyCiphertext *string
	if strings.TrimSpace(req.PrivateKeyPEM) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.PrivateKeyPEM)
		if err != nil {
			return err
		}
		keyCiphertext = &ciphertext
	}
	var passCiphertext *string
	if strings.TrimSpace(req.Passphrase) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.Passphrase)
		if err != nil {
			return err
		}
		passCiphertext = &ciphertext
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qhc := qtx.HostCredential
		beforeRow, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).UpdateSimple(
			qhc.HostID.Value(req.HostID),
			qhc.AccountName.Value(strings.TrimSpace(req.AccountName)),
			qhc.CredentialName.Value(strings.TrimSpace(req.CredentialName)),
			qhc.CredentialKind.Value(strings.TrimSpace(req.CredentialKind)),
			qhc.Status.Value(status),
			qhc.UpdatedAt.Value(time.Now()),
		); err != nil {
			return err
		}
		if _, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).Update(qhc.Username, nullOrString(username)); err != nil {
			return err
		}
		if _, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).Update(qhc.Description, nullOrString(description)); err != nil {
			return err
		}
		if secretCiphertext != nil {
			if _, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).Update(qhc.SecretCiphertext, *secretCiphertext); err != nil {
				return err
			}
		}
		if certCiphertext != nil {
			if _, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).Update(qhc.CertificatePemCiphertext, *certCiphertext); err != nil {
				return err
			}
		}
		if keyCiphertext != nil {
			if _, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).Update(qhc.PrivateKeyPemCiphertext, *keyCiphertext); err != nil {
				return err
			}
		}
		if passCiphertext != nil {
			if _, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).Update(qhc.PassphraseCiphertext, *passCiphertext); err != nil {
				return err
			}
		}
		before := struct {
			HostID            uint64 `json:"host_id"`
			AccountName       string `json:"account_name"`
			CredentialName    string `json:"credential_name"`
			CredentialKind    string `json:"credential_kind"`
			Status            string `json:"status"`
			HasSecret         bool   `json:"has_secret"`
			HasCertificatePEM bool   `json:"has_certificate_pem"`
			HasPrivateKeyPEM  bool   `json:"has_private_key_pem"`
			HasPassphrase     bool   `json:"has_passphrase"`
		}{
			HostID:            beforeRow.HostID,
			AccountName:       beforeRow.AccountName,
			CredentialName:    beforeRow.CredentialName,
			CredentialKind:    beforeRow.CredentialKind,
			Status:            beforeRow.Status,
			HasSecret:         beforeRow.SecretCiphertext != nil && strings.TrimSpace(*beforeRow.SecretCiphertext) != "",
			HasCertificatePEM: beforeRow.CertificatePemCiphertext != nil && strings.TrimSpace(*beforeRow.CertificatePemCiphertext) != "",
			HasPrivateKeyPEM:  beforeRow.PrivateKeyPemCiphertext != nil && strings.TrimSpace(*beforeRow.PrivateKeyPemCiphertext) != "",
			HasPassphrase:     beforeRow.PassphraseCiphertext != nil && strings.TrimSpace(*beforeRow.PassphraseCiphertext) != "",
		}
		after := struct {
			HostID                 uint64 `json:"host_id"`
			AccountName            string `json:"account_name"`
			CredentialName         string `json:"credential_name"`
			CredentialKind         string `json:"credential_kind"`
			Status                 string `json:"status"`
			SecretProvided         bool   `json:"secret_provided"`
			CertificatePEMProvided bool   `json:"certificate_pem_provided"`
			PrivateKeyPEMProvided  bool   `json:"private_key_pem_provided"`
			PassphraseProvided     bool   `json:"passphrase_provided"`
		}{
			HostID:                 req.HostID,
			AccountName:            strings.TrimSpace(req.AccountName),
			CredentialName:         strings.TrimSpace(req.CredentialName),
			CredentialKind:         strings.TrimSpace(req.CredentialKind),
			Status:                 status,
			SecretProvided:         strings.TrimSpace(req.Secret) != "",
			CertificatePEMProvided: strings.TrimSpace(req.CertificatePEM) != "",
			PrivateKeyPEMProvided:  strings.TrimSpace(req.PrivateKeyPEM) != "",
			PassphraseProvided:     strings.TrimSpace(req.Passphrase) != "",
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "update_host_credential", "host_credentials", strconv.FormatUint(id, 10), req.CredentialName, before, after)
	})
}

func (s *CredentialService) DeleteHostCredential(ctx context.Context, actor Actor, id uint64) error {
	qhc := s.q.HostCredential
	old, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if err := s.authorizeCredentialAction(ctx, actor, "host", old.HostID, "write"); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qhc := qtx.HostCredential
		beforeRow, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		before := struct {
			ID             uint64 `json:"id"`
			HostID         uint64 `json:"host_id"`
			CredentialName string `json:"credential_name"`
			CredentialKind string `json:"credential_kind"`
			Status         string `json:"status"`
		}{
			ID:             id,
			HostID:         beforeRow.HostID,
			CredentialName: beforeRow.CredentialName,
			CredentialKind: beforeRow.CredentialKind,
			Status:         beforeRow.Status,
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_host_credential", "host_credentials", strconv.FormatUint(id, 10), beforeRow.CredentialName, before, nil)
	})
}

func (s *CredentialService) RevealHostCredential(ctx context.Context, actor Actor, id uint64) (vo.CredentialRevealVO, error) {
	qhc := s.q.HostCredential
	old, err := qhc.WithContext(ctx).Where(qhc.ID.Eq(id)).First()
	if err != nil {
		return vo.CredentialRevealVO{}, err
	}
	if err := s.authorizeCredentialAction(ctx, actor, "host", old.HostID, "read"); err != nil {
		return vo.CredentialRevealVO{}, err
	}
	data, err := s.revealCredential(old.SecretCiphertext, old.CertificatePemCiphertext, old.PrivateKeyPemCiphertext, old.PassphraseCiphertext, id)
	if err != nil {
		return vo.CredentialRevealVO{}, err
	}
	_ = s.auditSvc.RecordSuccess(ctx, nil, actor, "reveal_credential", "host_credentials", strconv.FormatUint(id, 10), old.CredentialName, nil, struct {
		ParentID uint64 `json:"parent_id"`
	}{ParentID: old.HostID})
	return data, nil
}

func (s *CredentialService) ListServiceCredentials(ctx context.Context, actor Actor, serviceID uint64, page int, pageSize int) ([]vo.CredentialVO, int, error) {
	if err := s.authorizeCredentialAction(ctx, actor, "service", serviceID, "list"); err != nil {
		return nil, 0, err
	}
	qsc := s.q.ServiceCredential
	items, total64, err := qsc.WithContext(ctx).
		Where(qsc.ServiceID.Eq(serviceID)).
		Order(qsc.ID.Desc()).
		FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]vo.CredentialVO, 0, len(items))
	for _, it := range items {
		out = append(out, toServiceCredentialVO(*it))
	}
	return out, int(total64), nil
}

func (s *CredentialService) CreateServiceCredential(ctx context.Context, actor Actor, req request.ServiceCredentialCreateRequest) (uint64, error) {
	if err := s.authorizeCredentialAction(ctx, actor, "service", req.ServiceID, "write"); err != nil {
		return 0, err
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	now := time.Now()
	it := model.ServiceCredential{
		ServiceID:      req.ServiceID,
		AccountName:    strings.TrimSpace(req.AccountName),
		CredentialName: strings.TrimSpace(req.CredentialName),
		CredentialKind: strings.TrimSpace(req.CredentialKind),
		Status:         status,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if strings.TrimSpace(req.Username) != "" {
		v := strings.TrimSpace(req.Username)
		it.Username = &v
	}
	if strings.TrimSpace(req.Description) != "" {
		v := strings.TrimSpace(req.Description)
		it.Description = &v
	}
	if strings.TrimSpace(req.Secret) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.Secret)
		if err != nil {
			return 0, err
		}
		it.SecretCiphertext = &ciphertext
	}
	if strings.TrimSpace(req.CertificatePEM) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.CertificatePEM)
		if err != nil {
			return 0, err
		}
		it.CertificatePemCiphertext = &ciphertext
	}
	if strings.TrimSpace(req.PrivateKeyPEM) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.PrivateKeyPEM)
		if err != nil {
			return 0, err
		}
		it.PrivateKeyPemCiphertext = &ciphertext
	}
	if strings.TrimSpace(req.Passphrase) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.Passphrase)
		if err != nil {
			return 0, err
		}
		it.PassphraseCiphertext = &ciphertext
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := qtx.ServiceCredential.WithContext(ctx).Create(&it); err != nil {
			return err
		}
		after := struct {
			ServiceID         uint64 `json:"service_id"`
			AccountName       string `json:"account_name"`
			CredentialName    string `json:"credential_name"`
			CredentialKind    string `json:"credential_kind"`
			Status            string `json:"status"`
			HasSecret         bool   `json:"has_secret"`
			HasCertificatePEM bool   `json:"has_certificate_pem"`
			HasPrivateKeyPEM  bool   `json:"has_private_key_pem"`
			HasPassphrase     bool   `json:"has_passphrase"`
		}{
			ServiceID:         req.ServiceID,
			AccountName:       strings.TrimSpace(req.AccountName),
			CredentialName:    strings.TrimSpace(req.CredentialName),
			CredentialKind:    strings.TrimSpace(req.CredentialKind),
			Status:            status,
			HasSecret:         strings.TrimSpace(req.Secret) != "",
			HasCertificatePEM: strings.TrimSpace(req.CertificatePEM) != "",
			HasPrivateKeyPEM:  strings.TrimSpace(req.PrivateKeyPEM) != "",
			HasPassphrase:     strings.TrimSpace(req.Passphrase) != "",
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "create_service_credential", "service_credentials", strconv.FormatUint(it.ID, 10), req.CredentialName, nil, after)
	}); err != nil {
		return 0, err
	}
	return it.ID, nil
}

func (s *CredentialService) UpdateServiceCredential(ctx context.Context, actor Actor, id uint64, req request.ServiceCredentialUpdateRequest) error {
	qsc := s.q.ServiceCredential
	old, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if err := s.authorizeCredentialAction(ctx, actor, "service", old.ServiceID, "write"); err != nil {
		return err
	}
	if req.ServiceID != old.ServiceID {
		if err := s.authorizeCredentialAction(ctx, actor, "service", req.ServiceID, "write"); err != nil {
			return err
		}
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	username := strings.TrimSpace(req.Username)
	description := strings.TrimSpace(req.Description)
	var secretCiphertext *string
	if strings.TrimSpace(req.Secret) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.Secret)
		if err != nil {
			return err
		}
		secretCiphertext = &ciphertext
	}
	var certCiphertext *string
	if strings.TrimSpace(req.CertificatePEM) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.CertificatePEM)
		if err != nil {
			return err
		}
		certCiphertext = &ciphertext
	}
	var keyCiphertext *string
	if strings.TrimSpace(req.PrivateKeyPEM) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.PrivateKeyPEM)
		if err != nil {
			return err
		}
		keyCiphertext = &ciphertext
	}
	var passCiphertext *string
	if strings.TrimSpace(req.Passphrase) != "" {
		ciphertext, err := s.crypto.encryptPlaintext(req.Passphrase)
		if err != nil {
			return err
		}
		passCiphertext = &ciphertext
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qsc := qtx.ServiceCredential
		beforeRow, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).UpdateSimple(
			qsc.ServiceID.Value(req.ServiceID),
			qsc.AccountName.Value(strings.TrimSpace(req.AccountName)),
			qsc.CredentialName.Value(strings.TrimSpace(req.CredentialName)),
			qsc.CredentialKind.Value(strings.TrimSpace(req.CredentialKind)),
			qsc.Status.Value(status),
			qsc.UpdatedAt.Value(time.Now()),
		); err != nil {
			return err
		}
		if _, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).Update(qsc.Username, nullOrString(username)); err != nil {
			return err
		}
		if _, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).Update(qsc.Description, nullOrString(description)); err != nil {
			return err
		}
		if secretCiphertext != nil {
			if _, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).Update(qsc.SecretCiphertext, *secretCiphertext); err != nil {
				return err
			}
		}
		if certCiphertext != nil {
			if _, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).Update(qsc.CertificatePemCiphertext, *certCiphertext); err != nil {
				return err
			}
		}
		if keyCiphertext != nil {
			if _, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).Update(qsc.PrivateKeyPemCiphertext, *keyCiphertext); err != nil {
				return err
			}
		}
		if passCiphertext != nil {
			if _, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).Update(qsc.PassphraseCiphertext, *passCiphertext); err != nil {
				return err
			}
		}
		before := struct {
			ServiceID         uint64 `json:"service_id"`
			AccountName       string `json:"account_name"`
			CredentialName    string `json:"credential_name"`
			CredentialKind    string `json:"credential_kind"`
			Status            string `json:"status"`
			HasSecret         bool   `json:"has_secret"`
			HasCertificatePEM bool   `json:"has_certificate_pem"`
			HasPrivateKeyPEM  bool   `json:"has_private_key_pem"`
			HasPassphrase     bool   `json:"has_passphrase"`
		}{
			ServiceID:         beforeRow.ServiceID,
			AccountName:       beforeRow.AccountName,
			CredentialName:    beforeRow.CredentialName,
			CredentialKind:    beforeRow.CredentialKind,
			Status:            beforeRow.Status,
			HasSecret:         beforeRow.SecretCiphertext != nil && strings.TrimSpace(*beforeRow.SecretCiphertext) != "",
			HasCertificatePEM: beforeRow.CertificatePemCiphertext != nil && strings.TrimSpace(*beforeRow.CertificatePemCiphertext) != "",
			HasPrivateKeyPEM:  beforeRow.PrivateKeyPemCiphertext != nil && strings.TrimSpace(*beforeRow.PrivateKeyPemCiphertext) != "",
			HasPassphrase:     beforeRow.PassphraseCiphertext != nil && strings.TrimSpace(*beforeRow.PassphraseCiphertext) != "",
		}
		after := struct {
			ServiceID              uint64 `json:"service_id"`
			AccountName            string `json:"account_name"`
			CredentialName         string `json:"credential_name"`
			CredentialKind         string `json:"credential_kind"`
			Status                 string `json:"status"`
			SecretProvided         bool   `json:"secret_provided"`
			CertificatePEMProvided bool   `json:"certificate_pem_provided"`
			PrivateKeyPEMProvided  bool   `json:"private_key_pem_provided"`
			PassphraseProvided     bool   `json:"passphrase_provided"`
		}{
			ServiceID:              req.ServiceID,
			AccountName:            strings.TrimSpace(req.AccountName),
			CredentialName:         strings.TrimSpace(req.CredentialName),
			CredentialKind:         strings.TrimSpace(req.CredentialKind),
			Status:                 status,
			SecretProvided:         strings.TrimSpace(req.Secret) != "",
			CertificatePEMProvided: strings.TrimSpace(req.CertificatePEM) != "",
			PrivateKeyPEMProvided:  strings.TrimSpace(req.PrivateKeyPEM) != "",
			PassphraseProvided:     strings.TrimSpace(req.Passphrase) != "",
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "update_service_credential", "service_credentials", strconv.FormatUint(id, 10), req.CredentialName, before, after)
	})
}

func (s *CredentialService) DeleteServiceCredential(ctx context.Context, actor Actor, id uint64) error {
	qsc := s.q.ServiceCredential
	old, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if err := s.authorizeCredentialAction(ctx, actor, "service", old.ServiceID, "write"); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qsc := qtx.ServiceCredential
		beforeRow, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		before := struct {
			ID             uint64 `json:"id"`
			ServiceID      uint64 `json:"service_id"`
			CredentialName string `json:"credential_name"`
			CredentialKind string `json:"credential_kind"`
			Status         string `json:"status"`
		}{
			ID:             id,
			ServiceID:      beforeRow.ServiceID,
			CredentialName: beforeRow.CredentialName,
			CredentialKind: beforeRow.CredentialKind,
			Status:         beforeRow.Status,
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_service_credential", "service_credentials", strconv.FormatUint(id, 10), beforeRow.CredentialName, before, nil)
	})
}

func (s *CredentialService) RevealServiceCredential(ctx context.Context, actor Actor, id uint64) (vo.CredentialRevealVO, error) {
	qsc := s.q.ServiceCredential
	old, err := qsc.WithContext(ctx).Where(qsc.ID.Eq(id)).First()
	if err != nil {
		return vo.CredentialRevealVO{}, err
	}
	if err := s.authorizeCredentialAction(ctx, actor, "service", old.ServiceID, "read"); err != nil {
		return vo.CredentialRevealVO{}, err
	}
	data, err := s.revealCredential(old.SecretCiphertext, old.CertificatePemCiphertext, old.PrivateKeyPemCiphertext, old.PassphraseCiphertext, id)
	if err != nil {
		return vo.CredentialRevealVO{}, err
	}
	_ = s.auditSvc.RecordSuccess(ctx, nil, actor, "reveal_credential", "service_credentials", strconv.FormatUint(id, 10), old.CredentialName, nil, struct {
		ParentID uint64 `json:"parent_id"`
	}{ParentID: old.ServiceID})
	return data, nil
}

func (s *CredentialService) authorizeCredentialAction(ctx context.Context, actor Actor, parentType string, parentID uint64, action string) error {
	if actor.Role == "super_admin" {
		return nil
	}
	parentReadable, err := s.subjectHasAnyPermissionCode(ctx, "user", actor.UserID, parentType, parentID, []string{parentType + ".read", parentType + ".write"})
	if err != nil {
		return err
	}
	if !parentReadable {
		return ecode.NewAppError(http.StatusForbidden, ecode.CodeResourcePermissionDenied, "missing parent read permission", nil)
	}
	credentialType := parentType + "_credential"
	if action == "list" || action == "read" || action == "detail" || action == "reveal" {
		canRead, err := s.subjectHasAnyPermissionCode(ctx, "user", actor.UserID, credentialType, parentID, []string{credentialType + ".read"})
		if err != nil {
			return err
		}
		if !canRead {
			return ecode.NewAppError(http.StatusForbidden, ecode.CodeCredentialPermissionDenied, "missing credential read permission", nil)
		}
		return nil
	}
	canWrite, err := s.subjectHasAnyPermissionCode(ctx, "user", actor.UserID, credentialType, parentID, []string{credentialType + ".write"})
	if err != nil {
		return err
	}
	if !canWrite {
		return ecode.NewAppError(http.StatusForbidden, ecode.CodeCredentialPermissionDenied, "missing credential write permission", nil)
	}
	return nil
}

func (s *CredentialService) subjectHasAnyPermissionCode(ctx context.Context, subjectType string, subjectID uint64, objectType string, objectID uint64, codes []string) (bool, error) {
	effectiveCodes, err := s.permDefs.ExpandPermissionCodes(ctx, codes)
	if err != nil {
		return false, err
	}
	return s.grantRepo.SubjectHasAnyPermissionCode(ctx, subjectType, subjectID, objectType, objectID, effectiveCodes)
}

func (s *CredentialService) revealCredential(secretCipher *string, certCipher *string, keyCipher *string, passCipher *string, id uint64) (vo.CredentialRevealVO, error) {
	data := vo.CredentialRevealVO{ID: id}
	if secretCipher != nil && strings.TrimSpace(*secretCipher) != "" {
		plain, err := s.crypto.decryptCiphertext(*secretCipher)
		if err != nil {
			return vo.CredentialRevealVO{}, err
		}
		data.Secret = &plain
	}
	if certCipher != nil && strings.TrimSpace(*certCipher) != "" {
		plain, err := s.crypto.decryptCiphertext(*certCipher)
		if err != nil {
			return vo.CredentialRevealVO{}, err
		}
		data.CertificatePEM = &plain
	}
	if keyCipher != nil && strings.TrimSpace(*keyCipher) != "" {
		plain, err := s.crypto.decryptCiphertext(*keyCipher)
		if err != nil {
			return vo.CredentialRevealVO{}, err
		}
		data.PrivateKeyPEM = &plain
	}
	if passCipher != nil && strings.TrimSpace(*passCipher) != "" {
		plain, err := s.crypto.decryptCiphertext(*passCipher)
		if err != nil {
			return vo.CredentialRevealVO{}, err
		}
		data.Passphrase = &plain
	}
	return data, nil
}

func toHostCredentialVO(it model.HostCredential) vo.CredentialVO {
	out := vo.CredentialVO{
		ID:             it.ID,
		HostID:         it.HostID,
		AccountName:    it.AccountName,
		CredentialName: it.CredentialName,
		CredentialKind: it.CredentialKind,
		Status:         it.Status,
		CreatedAt:      it.CreatedAt,
		UpdatedAt:      it.UpdatedAt,
	}
	if it.Username != nil {
		out.Username = *it.Username
	}
	if it.Description != nil {
		out.Description = *it.Description
	}
	return out
}

func toServiceCredentialVO(it model.ServiceCredential) vo.CredentialVO {
	out := vo.CredentialVO{
		ID:             it.ID,
		ServiceID:      it.ServiceID,
		AccountName:    it.AccountName,
		CredentialName: it.CredentialName,
		CredentialKind: it.CredentialKind,
		Status:         it.Status,
		CreatedAt:      it.CreatedAt,
		UpdatedAt:      it.UpdatedAt,
	}
	if it.Username != nil {
		out.Username = *it.Username
	}
	if it.Description != nil {
		out.Description = *it.Description
	}
	return out
}
