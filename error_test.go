package response

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

// teapotError is a custom ErrorCoder used to verify the extensible mapping.
type teapotError struct{}

func (teapotError) Error() string        { return "short and stout" }
func (teapotError) StatusCode() int      { return http.StatusTeapot }
func (teapotError) ResponseCode() string { return "TEAPOT" }

func TestErrorMapsResponseError(t *testing.T) {
	cases := []struct {
		name    string
		err     *ResponseError
		status  int
		code    string
		message string
	}{
		{"BadRequest", BadRequest("bad request"), http.StatusBadRequest, CodeBadRequest, "bad request"},
		{"Unauthorized", Unauthorized("unauthorized"), http.StatusUnauthorized, CodeUnauthorized, "unauthorized"},
		{"Forbidden", Forbidden("forbidden"), http.StatusForbidden, CodeForbidden, "forbidden"},
		{"NotFound", NotFound("user not found"), http.StatusNotFound, CodeNotFound, "user not found"},
		{"Conflict", Conflict("conflict"), http.StatusConflict, CodeConflict, "conflict"},
		{"UnprocessableEntity", UnprocessableEntity("nope"), http.StatusUnprocessableEntity, CodeValidationError, "nope"},
		{"InternalServerError", InternalServerError("boom"), http.StatusInternalServerError, CodeInternalServerError, "boom"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := newContext()
			Error(c, tc.err)

			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d", w.Code, tc.status)
			}
			body := decode(t, w)
			if len(body) != 3 {
				t.Fatalf("body has %d keys, want 3: %v", len(body), body)
			}
			if body["success"] != false {
				t.Errorf("success = %v, want false", body["success"])
			}
			if body["code"] != tc.code {
				t.Errorf("code = %v, want %s", body["code"], tc.code)
			}
			if body["message"] != tc.message {
				t.Errorf("message = %v, want %q", body["message"], tc.message)
			}
		})
	}
}

func TestErrorMapsWrappedError(t *testing.T) {
	c, w := newContext()
	Error(c, fmt.Errorf("lookup failed: %w", NotFound("user not found")))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	body := decode(t, w)
	if body["code"] != CodeNotFound {
		t.Errorf("code = %v, want %s", body["code"], CodeNotFound)
	}
	if body["message"] != "user not found" {
		t.Errorf("message = %v, want %q", body["message"], "user not found")
	}
}

func TestErrorMapsCustomErrorCoder(t *testing.T) {
	c, w := newContext()
	Error(c, teapotError{})

	if w.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusTeapot)
	}
	body := decode(t, w)
	if body["code"] != "TEAPOT" {
		t.Errorf("code = %v, want TEAPOT", body["code"])
	}
	if body["message"] != "short and stout" {
		t.Errorf("message = %v, want %q", body["message"], "short and stout")
	}
}

func TestErrorFallsBackToInternal(t *testing.T) {
	c, w := newContext()
	Error(c, errors.New("db connection refused"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	body := decode(t, w)
	if body["code"] != CodeInternalServerError {
		t.Errorf("code = %v, want %s", body["code"], CodeInternalServerError)
	}
	if body["message"] != "db connection refused" {
		t.Errorf("message = %v, want %q", body["message"], "db connection refused")
	}
}

func TestErrorPanicsOnNil(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on nil error")
		}
	}()
	c, _ := newContext()
	Error(c, nil)
}

func TestNewError(t *testing.T) {
	e := NewError(http.StatusTeapot, "TEAPOT", "short and stout")

	if e.Error() != "short and stout" {
		t.Errorf("Error() = %q, want %q", e.Error(), "short and stout")
	}
	if e.StatusCode() != http.StatusTeapot {
		t.Errorf("StatusCode() = %d, want %d", e.StatusCode(), http.StatusTeapot)
	}
	if e.ResponseCode() != "TEAPOT" {
		t.Errorf("ResponseCode() = %q, want %q", e.ResponseCode(), "TEAPOT")
	}

	c, w := newContext()
	Error(c, e)
	if w.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusTeapot)
	}
}
