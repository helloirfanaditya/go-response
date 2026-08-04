package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// newContext returns a Gin context and recorder for tests.
func newContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

// decode unmarshals the recorded body into a map.
func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body %q: %v", w.Body.String(), err)
	}
	return body
}

func TestSuccess(t *testing.T) {
	c, w := newContext()
	Success(c, gin.H{"id": 1})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := decode(t, w)
	if len(body) != 4 {
		t.Fatalf("body has %d keys, want 4: %v", len(body), body)
	}
	if body["success"] != true {
		t.Errorf("success = %v, want true", body["success"])
	}
	if body["code"] != CodeSuccess {
		t.Errorf("code = %v, want %s", body["code"], CodeSuccess)
	}
	if body["message"] != "Success" {
		t.Errorf("message = %v, want %q", body["message"], "Success")
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data = %v, want an object", body["data"])
	}
	if data["id"] != float64(1) {
		t.Errorf("data.id = %v, want 1", data["id"])
	}
}

func TestCreated(t *testing.T) {
	c, w := newContext()
	Created(c, gin.H{"id": 1})

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	body := decode(t, w)
	if body["success"] != true {
		t.Errorf("success = %v, want true", body["success"])
	}
	if body["code"] != CodeCreated {
		t.Errorf("code = %v, want %s", body["code"], CodeCreated)
	}
	if body["message"] != "Created" {
		t.Errorf("message = %v, want %q", body["message"], "Created")
	}
}

func TestNoContent(t *testing.T) {
	c, w := newContext()
	NoContent(c)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if got := w.Body.String(); got != "" {
		t.Errorf("body = %q, want empty", got)
	}
}
