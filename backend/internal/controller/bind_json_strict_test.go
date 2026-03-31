package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type testReq struct {
	Name string `json:"name" binding:"required"`
}

func TestBindJSONStrict_UnknownFieldReturnsDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"a","unknown":1}`))
	c.Request.Header.Set("Content-Type", "application/json")

	var req testReq
	ok := bindJSONStrict(c, &req)
	if ok {
		t.Fatalf("expected bind to fail")
	}
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["details"] == nil {
		t.Fatalf("expected details")
	}
}

func TestBindJSONStrict_MissingRequiredReturnsDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	var req testReq
	ok := bindJSONStrict(c, &req)
	if ok {
		t.Fatalf("expected bind to fail")
	}
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["details"] == nil {
		t.Fatalf("expected details")
	}
}

