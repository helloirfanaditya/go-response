package response

import (
	"net/http"
	"testing"
)

func TestPaginateComputesTotalPage(t *testing.T) {
	c, w := newContext()
	Paginate(c, []map[string]any{{"id": 1}}, Meta{Page: 2, PerPage: 10, Total: 25})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := decode(t, w)
	if len(body) != 5 {
		t.Fatalf("body has %d keys, want 5: %v", len(body), body)
	}
	if body["success"] != true {
		t.Errorf("success = %v, want true", body["success"])
	}
	if body["code"] != CodeSuccess {
		t.Errorf("code = %v, want %s", body["code"], CodeSuccess)
	}
	if _, ok := body["data"].([]any); !ok {
		t.Errorf("data = %v, want a list", body["data"])
	}

	meta := body["meta"].(map[string]any)
	if meta["page"] != float64(2) {
		t.Errorf("meta.page = %v, want 2", meta["page"])
	}
	if meta["perPage"] != float64(10) {
		t.Errorf("meta.perPage = %v, want 10", meta["perPage"])
	}
	if meta["total"] != float64(25) {
		t.Errorf("meta.total = %v, want 25", meta["total"])
	}
	if meta["totalPage"] != float64(3) {
		t.Errorf("meta.totalPage = %v, want 3", meta["totalPage"])
	}
}

func TestPaginateHonorsExplicitTotalPage(t *testing.T) {
	c, w := newContext()
	Paginate(c, []map[string]any{}, Meta{Page: 1, PerPage: 10, Total: 25, TotalPage: 9})

	meta := decode(t, w)["meta"].(map[string]any)
	if meta["totalPage"] != float64(9) {
		t.Errorf("meta.totalPage = %v, want 9", meta["totalPage"])
	}
}

func TestPaginateZeroPerPage(t *testing.T) {
	c, w := newContext()
	Paginate(c, []map[string]any{}, Meta{Page: 1, PerPage: 0, Total: 0})

	meta := decode(t, w)["meta"].(map[string]any)
	if meta["totalPage"] != float64(0) {
		t.Errorf("meta.totalPage = %v, want 0", meta["totalPage"])
	}
}
