package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"irms/backend/internal/model"
	"irms/backend/internal/query"
	ecode "irms/backend/internal/pkg/errors"
	"irms/backend/internal/vo"

	"gorm.io/gorm"
)

type AuditListFilter struct {
	ActorUserID *uint64
	TargetType  string
	TargetID    string
	Result      string
	From        string
	To          string
}

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) ListPaged(ctx context.Context, f AuditListFilter, page int, pageSize int) ([]vo.AuditLogVO, int, error) {
	qa := query.Use(s.db).AuditLog
	do := qa.WithContext(ctx)
	if f.ActorUserID != nil && *f.ActorUserID > 0 {
		do = do.Where(qa.ActorUserID.Eq(*f.ActorUserID))
	}
	if v := strings.TrimSpace(f.TargetType); v != "" {
		do = do.Where(qa.TargetType.Eq(v))
	}
	if v := strings.TrimSpace(f.TargetID); v != "" {
		do = do.Where(qa.TargetID.Eq(v))
	}
	if v := strings.TrimSpace(f.Result); v != "" {
		do = do.Where(qa.Result.Eq(v))
	}
	if v := strings.TrimSpace(f.From); v != "" {
		tm, ok := parseTimeFlexible(v)
		if !ok {
			return nil, 0, ecode.NewAppError(400, ecode.CodeInvalidArgument, "invalid from", nil)
		}
		do = do.Where(qa.OccurredAt.Gte(tm))
	}
	if v := strings.TrimSpace(f.To); v != "" {
		tm, ok := parseTimeFlexible(v)
		if !ok {
			return nil, 0, ecode.NewAppError(400, ecode.CodeInvalidArgument, "invalid to", nil)
		}
		do = do.Where(qa.OccurredAt.Lte(tm))
	}
	rows, total64, err := do.Order(qa.ID.Desc()).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]vo.AuditLogVO, 0, len(rows))
	for _, r := range rows {
		before := ""
		if r.BeforeJSON != nil {
			before = *r.BeforeJSON
		}
		after := ""
		if r.AfterJSON != nil {
			after = *r.AfterJSON
		}
		ip := ""
		if r.IP != nil {
			ip = *r.IP
		}
		out = append(out, vo.AuditLogVO{
			ID:                    r.ID,
			ActorUserID:           r.ActorUserID,
			ActorUsernameSnapshot: r.ActorUsernameSnapshot,
			Action:                r.Action,
			TargetType:            r.TargetType,
			TargetID:              r.TargetID,
			TargetNameSnapshot:    r.TargetNameSnapshot,
			OccurredAt:            r.OccurredAt,
			BeforeJSON:            before,
			AfterJSON:             after,
			Result:                r.Result,
			IP:                    ip,
		})
	}
	return out, int(total64), nil
}

func parseTimeFlexible(v string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local); err == nil {
		return t, true
	}
	return time.Time{}, false
}

func (s *AuditService) RecordSuccess(ctx context.Context, tx *gorm.DB, actor Actor, action string, targetType string, targetID string, targetName string, before interface{}, after interface{}) error {
	return s.record(ctx, tx, actor, action, targetType, targetID, targetName, before, after, "success")
}

func (s *AuditService) RecordFailure(ctx context.Context, tx *gorm.DB, actor Actor, action string, targetType string, targetID string, targetName string, before interface{}, after interface{}) error {
	return s.record(ctx, tx, actor, action, targetType, targetID, targetName, before, after, "failure")
}

func (s *AuditService) record(ctx context.Context, tx *gorm.DB, actor Actor, action string, targetType string, targetID string, targetName string, before interface{}, after interface{}, result string) error {
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(after)
	if err != nil {
		return err
	}
	var beforeStr *string
	if string(beforeJSON) != "null" {
		v := string(beforeJSON)
		beforeStr = &v
	}
	var afterStr *string
	if string(afterJSON) != "null" {
		v := string(afterJSON)
		afterStr = &v
	}
	var ip *string
	if strings.TrimSpace(actor.IP) != "" {
		v := strings.TrimSpace(actor.IP)
		ip = &v
	}
	item := model.AuditLog{
		ActorUserID:           actor.UserID,
		ActorUsernameSnapshot: actor.Username,
		Action:                action,
		TargetType:            targetType,
		TargetID:              targetID,
		TargetNameSnapshot:    targetName,
		OccurredAt:            time.Now(),
		BeforeJSON:            beforeStr,
		AfterJSON:             afterStr,
		Result:                result,
		IP:                    ip,
	}
	db := tx
	if db == nil {
		db = s.db
	}
	return query.Use(db).AuditLog.WithContext(ctx).Create(&item)
}
