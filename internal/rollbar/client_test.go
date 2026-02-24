package rollbar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewClientDefaults(t *testing.T) {
	c := NewClient(Config{})
	if c.baseURL != "https://api.rollbar.com" {
		t.Fatalf("unexpected baseURL: %q", c.baseURL)
	}
	if c.httpClient.Timeout != 15*time.Second {
		t.Fatalf("unexpected timeout: %s", c.httpClient.Timeout)
	}
}

func TestListItemsSuccess(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	var gotToken string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		gotToken = r.Header.Get("X-Rollbar-Access-Token")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"err":0,"result":{"items":[{"id":123,"counter":10,"status":"active","last_occurrence":{"body":{"trace":{"exception":{"message":"panic happened"}}},"level":"error","environment":"production","timestamp":1700000000}}]}}`))
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL, Timeout: 2 * time.Second})
	resp, err := client.ListItems(context.Background(), ListItemsOptions{
		Page:        2,
		Status:      "active",
		Environment: "production",
		Level:       []string{"error", "critical"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/1/items" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotToken != "tok" {
		t.Fatalf("unexpected token header: %q", gotToken)
	}
	if gotQuery.Get("page") != "2" || gotQuery.Get("status") != "active" || gotQuery.Get("environment") != "production" {
		t.Fatalf("unexpected query: %#v", gotQuery)
	}
	levels := gotQuery["level"]
	if len(levels) != 2 || levels[0] != "error" || levels[1] != "critical" {
		t.Fatalf("unexpected levels: %#v", levels)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
	item := resp.Items[0]
	if item.ID != 123 || item.Counter != 10 {
		t.Fatalf("unexpected item ids: %#v", item)
	}
	if item.Title != "panic happened" || item.Level != "error" || item.Environment != "production" {
		t.Fatalf("fallback normalization failed: %#v", item)
	}
	if item.LastOccurrenceTimestamp != 1700000000 {
		t.Fatalf("unexpected last occurrence timestamp: %d", item.LastOccurrenceTimestamp)
	}
	if resp.Raw == nil || resp.Raw["err"] == nil {
		t.Fatalf("expected raw response to be present: %#v", resp.Raw)
	}
}

func TestListItemsErrorCases(t *testing.T) {
	t.Run("missing token", func(t *testing.T) {
		client := NewClient(Config{BaseURL: "https://api.rollbar.com"})
		if _, err := client.ListItems(context.Background(), ListItemsOptions{}); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("non-2xx", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "boom", http.StatusBadRequest)
		}))
		defer ts.Close()

		client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
		if _, err := client.ListItems(context.Background(), ListItemsOptions{}); err == nil || !strings.Contains(err.Error(), "status=400") {
			t.Fatalf("expected non-2xx error, got %v", err)
		}
	})

	t.Run("envelope err", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"err":1,"message":"nope"}`))
		}))
		defer ts.Close()

		client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
		if _, err := client.ListItems(context.Background(), ListItemsOptions{}); err == nil || !strings.Contains(err.Error(), "err=1") {
			t.Fatalf("expected envelope error, got %v", err)
		}
	})
}

func TestGetItemByIDAndUUID(t *testing.T) {
	var lastPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastPath = r.URL.Path
		_, _ = w.Write([]byte(`{"err":0,"result":{"item":{"id":42,"counter":7,"title":"boom","level":"warning"}}}`))
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})

	byID, err := client.GetItemByID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected get-by-id error: %v", err)
	}
	if lastPath != "/api/1/item/42" {
		t.Fatalf("unexpected path for get-by-id: %s", lastPath)
	}
	if byID.Item.ID != 42 || byID.Item.Title != "boom" {
		t.Fatalf("unexpected item from get-by-id: %#v", byID.Item)
	}

	byUUID, err := client.GetItemByUUID(context.Background(), "abcd-1234")
	if err != nil {
		t.Fatalf("unexpected get-by-uuid error: %v", err)
	}
	if lastPath != "/api/1/item/abcd-1234" {
		t.Fatalf("unexpected path for get-by-uuid: %s", lastPath)
	}
	if byUUID.Item.ID != 42 {
		t.Fatalf("unexpected item from get-by-uuid: %#v", byUUID.Item)
	}

	if _, err := client.GetItemByUUID(context.Background(), "   "); err == nil {
		t.Fatalf("expected empty-uuid error")
	}
}

func TestListItemInstances(t *testing.T) {
	var gotPath string
	var gotQuery url.Values

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"err":0,"result":{"instances":[{"id":501,"uuid":"inst-1","timestamp":1700001000,"level":"error","environment":"production","body":{"trace":{"frames":[{"filename":"app/main.go","lineno":42,"method":"handler"}]}},"request":{"url":"https://example.com/checkout"},"custom":{"order_id":"ord_123"}}]}}`))
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
	resp, err := client.ListItemInstances(context.Background(), "42", 2)
	if err != nil {
		t.Fatalf("unexpected list instances error: %v", err)
	}

	if gotPath != "/api/1/item/42/instances" {
		t.Fatalf("unexpected instances path: %s", gotPath)
	}
	if gotQuery.Get("page") != "2" {
		t.Fatalf("unexpected instances query: %#v", gotQuery)
	}
	if len(resp.Instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(resp.Instances))
	}

	instance := resp.Instances[0]
	if instance.ID != 501 || instance.UUID != "inst-1" || instance.Timestamp != 1700001000 {
		t.Fatalf("unexpected instance IDs: %#v", instance)
	}
	if len(instance.StackFrames) != 1 {
		t.Fatalf("expected one stack frame, got %#v", instance.StackFrames)
	}
	frame := instance.StackFrames[0]
	if frame.Filename != "app/main.go" || frame.Line != 42 || frame.Method != "handler" {
		t.Fatalf("unexpected frame details: %#v", frame)
	}
	if instance.Payload == nil || instance.Payload["request"] == nil || instance.Payload["custom"] == nil {
		t.Fatalf("expected payload details, got %#v", instance.Payload)
	}
}

func TestListItemInstancesValidation(t *testing.T) {
	client := NewClient(Config{AccessToken: "tok", BaseURL: "https://api.rollbar.com"})
	if _, err := client.ListItemInstances(context.Background(), "  ", 1); err == nil {
		t.Fatalf("expected empty identifier error")
	}
}

func TestUpdateItemByID(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotContentType string
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"err":0,"result":{"item":{"id":99,"counter":11,"title":"Updated"}}}`))
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
	resp, err := client.UpdateItemByID(context.Background(), 99, map[string]any{
		"status":              "resolved",
		"resolved_in_version": "aabbcc1",
	})
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/1/item/99" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotContentType != "application/json" {
		t.Fatalf("unexpected content type: %s", gotContentType)
	}
	if gotBody["status"] != "resolved" || gotBody["resolved_in_version"] != "aabbcc1" {
		t.Fatalf("unexpected request body: %#v", gotBody)
	}
	if resp.Item.ID != 99 || resp.Item.Title != "Updated" {
		t.Fatalf("unexpected updated item: %#v", resp.Item)
	}
}

func TestUpdateItemByIDErrorCases(t *testing.T) {
	client := NewClient(Config{AccessToken: "tok", BaseURL: "https://api.rollbar.com"})

	if _, err := client.UpdateItemByID(context.Background(), 0, map[string]any{"status": "active"}); err == nil {
		t.Fatalf("expected invalid-id error")
	}
	if _, err := client.UpdateItemByID(context.Background(), 1, map[string]any{}); err == nil {
		t.Fatalf("expected empty-body error")
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client = NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
	if _, err := client.UpdateItemByID(context.Background(), 1, map[string]any{"status": "active"}); err == nil || !strings.Contains(err.Error(), "status=500") {
		t.Fatalf("expected non-2xx error, got %v", err)
	}
}
