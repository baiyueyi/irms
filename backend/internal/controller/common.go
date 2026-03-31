package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	dtoresp "irms/backend/internal/dto/response"
	"irms/backend/internal/middleware"
	ecode "irms/backend/internal/pkg/errors"
	"irms/backend/internal/pkg/response"
	"irms/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func parseUint64Param(c *gin.Context, name string) (uint64, bool) {
	raw := c.Param(name)
	if raw == "" {
		response.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "missing param: "+name, dtoresp.ParamErrorDetails{Param: name})
		return 0, false
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid param: "+name, dtoresp.ParamErrorDetails{Param: name})
		return 0, false
	}
	return v, true
}

func currentActor(c *gin.Context) (service.Actor, bool) {
	cu, ok := middleware.CurrentUserFromContext(c)
	if !ok {
		return service.Actor{}, false
	}
	return service.Actor{
		UserID:   cu.ID,
		Username: cu.Username,
		Role:     cu.Role,
		IP:       c.ClientIP(),
	}, true
}

func bindJSONStrict(c *gin.Context, dst interface{}) bool {
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		msg := "invalid json"
		var details interface{}
		if strings.HasPrefix(err.Error(), "json: unknown field ") {
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			field = strings.Trim(field, "\"")
			details = dtoresp.ValidationErrorDetails{Errors: []dtoresp.FieldErrorItem{{Field: field, Tag: "unknown"}}}
			msg = "unknown json field"
		} else {
			var ute *json.UnmarshalTypeError
			if errors.As(err, &ute) && strings.TrimSpace(ute.Field) != "" {
				details = dtoresp.ValidationErrorDetails{Errors: []dtoresp.FieldErrorItem{{
					Field: ute.Field,
					Tag:   "type",
					Param: ute.Type.String(),
				}}}
				msg = "invalid field type"
			}
		}
		response.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, msg, details)
		return false
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		response.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid json", struct {
			Error string `json:"error"`
		}{Error: "multiple_json_values"})
		return false
	}
	if err := binding.Validator.ValidateStruct(dst); err != nil {
		if ves, ok := err.(validator.ValidationErrors); ok {
			items := make([]dtoresp.FieldErrorItem, 0, len(ves))
			for _, fe := range ves {
				items = append(items, dtoresp.FieldErrorItem{
					Field: jsonFieldName(dst, fe.StructField()),
					Tag:   fe.Tag(),
					Param: fe.Param(),
				})
			}
			response.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid request", dtoresp.ValidationErrorDetails{Errors: items})
			return false
		}
		response.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid request", dtoresp.ValidationErrorDetails{Errors: []dtoresp.FieldErrorItem{{Field: "", Tag: "invalid"}}})
		return false
	}
	return true
}

func jsonFieldName(dst interface{}, structField string) string {
	t := reflect.TypeOf(dst)
	if t == nil {
		return structField
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return structField
	}
	if f, ok := t.FieldByName(structField); ok {
		if tag := strings.TrimSpace(f.Tag.Get("json")); tag != "" {
			if idx := strings.Index(tag, ","); idx >= 0 {
				tag = tag[:idx]
			}
			if tag != "" && tag != "-" {
				return tag
			}
		}
	}
	return structField
}

func parseUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

func parsePaginationStrict(c *gin.Context) (page int, pageSize int, ok bool) {
	page = 1
	pageSize = 20
	if raw := strings.TrimSpace(c.Query("page")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			failInvalidQuery(c, "page", "invalid")
			return 0, 0, false
		}
		page = v
	}
	if raw := strings.TrimSpace(c.Query("page_size")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 100 {
			failInvalidQuery(c, "page_size", "invalid")
			return 0, 0, false
		}
		pageSize = v
	}
	return page, pageSize, true
}

func parseQueryUint64PtrStrict(c *gin.Context, name string) (*uint64, bool) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, true
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		failInvalidQuery(c, name, "invalid")
		return nil, false
	}
	return &v, true
}

func parseQueryEnumStrict(c *gin.Context, name string, allowed map[string]struct{}) (string, bool) {
	v := strings.TrimSpace(c.Query(name))
	if v == "" {
		return "", true
	}
	if _, ok := allowed[v]; !ok {
		failInvalidQuery(c, name, "invalid")
		return "", false
	}
	return v, true
}

func failInvalidQuery(c *gin.Context, field string, tag string) {
	response.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid query", dtoresp.ValidationErrorDetails{
		Errors: []dtoresp.FieldErrorItem{{Field: field, Tag: tag}},
	})
}

func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	var ae *ecode.AppError
	if errors.As(err, &ae) {
		response.Fail(c, ae.Status, ae.Code, ae.Message, ae.Details)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Fail(c, http.StatusNotFound, ecode.CodeNotFound, "not found", nil)
		return
	}
	response.Fail(c, http.StatusInternalServerError, ecode.CodeInternal, "internal error", nil)
}
