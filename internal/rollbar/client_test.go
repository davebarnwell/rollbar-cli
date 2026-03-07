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

func TestListItemsParsesStringID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"err":0,"result":{"items":[{"id":"456","counter":10,"title":"boom","level":"warning","status":"active","environment":"production","last_occurrence_timestamp":"1700000000"}]}}`))
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
	resp, err := client.ListItems(context.Background(), ListItemsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected one item, got %d", len(resp.Items))
	}
	if resp.Items[0].ID != 456 {
		t.Fatalf("expected parsed string id, got %#v", resp.Items[0])
	}
	if resp.Items[0].LastOccurrenceTimestamp != 1700000000 {
		t.Fatalf("expected parsed string timestamp, got %#v", resp.Items[0])
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

func TestListUsers(t *testing.T) {
	var gotPath string
	var gotToken string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotToken = r.Header.Get("X-Rollbar-Access-Token")
		_, _ = w.Write([]byte(`{"err":0,"result":{"users":[{"id":"7","username":"alice","email":"alice@example.com"},{"user_id":8,"name":"bob","email":"bob@example.com"}]}}`))
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
	resp, err := client.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected list users error: %v", err)
	}

	if gotPath != "/api/1/users" {
		t.Fatalf("unexpected users path: %s", gotPath)
	}
	if gotToken != "tok" {
		t.Fatalf("unexpected token header: %q", gotToken)
	}
	if len(resp.Users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(resp.Users))
	}
	if resp.Users[0].ID != 7 || resp.Users[0].Username != "alice" || resp.Users[0].Email != "alice@example.com" {
		t.Fatalf("unexpected first user: %#v", resp.Users[0])
	}
	if resp.Users[1].ID != 8 || resp.Users[1].Username != "bob" {
		t.Fatalf("unexpected fallback user: %#v", resp.Users[1])
	}
	if resp.Raw == nil || resp.Raw["err"] == nil {
		t.Fatalf("expected raw response to be present: %#v", resp.Raw)
	}
}

func TestListEnvironments(t *testing.T) {
	var gotPaths []string
	var gotPages []string
	var gotToken string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		gotPages = append(gotPages, r.URL.Query().Get("page"))
		gotToken = r.Header.Get("X-Rollbar-Access-Token")

		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"err":0,"result":{"environments":[
				{"id":"1","project_id":"88","environment":"production"},
				{"id":2,"project_id":88,"name":"staging"},
				{"id":3,"project_id":88,"environment":"preview-1"},
				{"id":4,"project_id":88,"environment":"preview-2"},
				{"id":5,"project_id":88,"environment":"preview-3"},
				{"id":6,"project_id":88,"environment":"preview-4"},
				{"id":7,"project_id":88,"environment":"preview-5"},
				{"id":8,"project_id":88,"environment":"preview-6"},
				{"id":9,"project_id":88,"environment":"preview-7"},
				{"id":10,"project_id":88,"environment":"preview-8"},
				{"id":11,"project_id":88,"environment":"preview-9"},
				{"id":12,"project_id":88,"environment":"preview-10"},
				{"id":13,"project_id":88,"environment":"preview-11"},
				{"id":14,"project_id":88,"environment":"preview-12"},
				{"id":15,"project_id":88,"environment":"preview-13"},
				{"id":16,"project_id":88,"environment":"preview-14"},
				{"id":17,"project_id":88,"environment":"preview-15"},
				{"id":18,"project_id":88,"environment":"preview-16"},
				{"id":19,"project_id":88,"environment":"preview-17"},
				{"id":20,"project_id":88,"environment":"preview-18"}
			]}}`))
		case "2":
			_, _ = w.Write([]byte(`{"err":0,"result":{"environments":[{"id":21,"project_id":88,"environment":"sandbox"}]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
	resp, err := client.ListEnvironments(context.Background())
	if err != nil {
		t.Fatalf("unexpected list environments error: %v", err)
	}

	if len(gotPaths) != 2 || gotPaths[0] != "/api/1/environments" || gotPaths[1] != "/api/1/environments" {
		t.Fatalf("unexpected environment paths: %#v", gotPaths)
	}
	if len(gotPages) != 2 || gotPages[0] != "1" || gotPages[1] != "2" {
		t.Fatalf("unexpected environment pages: %#v", gotPages)
	}
	if gotToken != "tok" {
		t.Fatalf("unexpected token header: %q", gotToken)
	}
	if len(resp.Environments) != 21 {
		t.Fatalf("expected 21 environments, got %d", len(resp.Environments))
	}
	if resp.Environments[0].ID != 1 || resp.Environments[0].ProjectID != 88 || resp.Environments[0].Name != "production" {
		t.Fatalf("unexpected first environment: %#v", resp.Environments[0])
	}
	if resp.Environments[1].Name != "staging" {
		t.Fatalf("expected fallback environment name, got %#v", resp.Environments[1])
	}
	if len(resp.RawPages) != 2 || resp.RawPages[0]["err"] == nil {
		t.Fatalf("expected raw pages to be present: %#v", resp.RawPages)
	}
}

func TestListEnvironmentsDirectArray(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"err":0,"result":[{"id":1,"project_id":88,"environment":"production"}]}`))
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
	resp, err := client.ListEnvironments(context.Background())
	if err != nil {
		t.Fatalf("unexpected list environments error: %v", err)
	}
	if len(resp.Environments) != 1 || resp.Environments[0].Name != "production" {
		t.Fatalf("unexpected environments: %#v", resp.Environments)
	}
}

func TestGetUserByID(t *testing.T) {
	var gotPath string
	var gotToken string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotToken = r.Header.Get("X-Rollbar-Access-Token")
		_, _ = w.Write([]byte(`{"err":0,"result":{"id":"7","username":"alice","email":"alice@example.com"}}`))
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
	resp, err := client.GetUserByID(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected get user error: %v", err)
	}

	if gotPath != "/api/1/user/7" {
		t.Fatalf("unexpected user path: %s", gotPath)
	}
	if gotToken != "tok" {
		t.Fatalf("unexpected token header: %q", gotToken)
	}
	if resp.User.ID != 7 || resp.User.Username != "alice" || resp.User.Email != "alice@example.com" {
		t.Fatalf("unexpected user: %#v", resp.User)
	}
	if resp.Raw == nil || resp.Raw["err"] == nil {
		t.Fatalf("expected raw response to be present: %#v", resp.Raw)
	}

	if _, err := client.GetUserByID(context.Background(), 0); err == nil {
		t.Fatalf("expected invalid user id error")
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

func TestListItemInstancesDirectArrayAndTraceChain(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"err":0,"result":[{"id":501,"uuid":"inst-1","timestamp":1700001000,"body":{"trace_chain":[{"frames":[{"filename":"app/main.go","lineno":42,"method":"handler"},{"filename":"worker.go","lineno":7,"method":"run"}]}]}}]}`))
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
	resp, err := client.ListItemInstances(context.Background(), "42", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Instances) != 1 {
		t.Fatalf("expected one instance, got %d", len(resp.Instances))
	}
	if len(resp.Instances[0].StackFrames) != 2 {
		t.Fatalf("expected two trace_chain frames, got %#v", resp.Instances[0].StackFrames)
	}
}

func TestListItemInstancesUsesNestedDataFields(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"err":0,"result":{"instances":[{"id":455732314098,"timestamp":1772663384,"data":{"uuid":"b0182d40-9e68-4f83-9cac-d909c85073c2","level":"warning","environment":"production","timestamp":1772663384}}]}}`))
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
	resp, err := client.ListItemInstances(context.Background(), "42", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Instances) != 1 {
		t.Fatalf("expected one instance, got %d", len(resp.Instances))
	}
	instance := resp.Instances[0]
	if instance.UUID != "b0182d40-9e68-4f83-9cac-d909c85073c2" {
		t.Fatalf("expected nested uuid, got %#v", instance)
	}
	if instance.Level != "warning" || instance.Environment != "production" {
		t.Fatalf("expected nested level/environment, got %#v", instance)
	}
}

func TestListItemInstancesValidation(t *testing.T) {
	client := NewClient(Config{AccessToken: "tok", BaseURL: "https://api.rollbar.com"})
	if _, err := client.ListItemInstances(context.Background(), "  ", 1); err == nil {
		t.Fatalf("expected empty identifier error")
	}
}

func TestListItemsDecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"err":0,"result":{"items":["bad"]}}`))
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
	if _, err := client.ListItems(context.Background(), ListItemsOptions{}); err == nil || !strings.Contains(err.Error(), "decode item 0") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestGetOccurrenceByIDAndUUID(t *testing.T) {
	var lastPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastPath = r.URL.Path
		_, _ = w.Write([]byte(`{"err":0,"result":{"instance":{"id":501,"uuid":"inst-1","timestamp":1700001000,"level":"error","environment":"production"}}}`))
	}))
	defer ts.Close()

	client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})

	byID, err := client.GetOccurrenceByID(context.Background(), 501)
	if err != nil {
		t.Fatalf("unexpected get-by-id error: %v", err)
	}
	if lastPath != "/api/1/instance/501" {
		t.Fatalf("unexpected path for get-by-id: %s", lastPath)
	}
	if byID.Occurrence.ID != 501 || byID.Occurrence.UUID != "inst-1" {
		t.Fatalf("unexpected occurrence from get-by-id: %#v", byID.Occurrence)
	}

	byUUID, err := client.GetOccurrenceByUUID(context.Background(), "inst-1")
	if err != nil {
		t.Fatalf("unexpected get-by-uuid error: %v", err)
	}
	if lastPath != "/api/1/instance/inst-1" {
		t.Fatalf("unexpected path for get-by-uuid: %s", lastPath)
	}
	if byUUID.Occurrence.ID != 501 || byUUID.Occurrence.UUID != "inst-1" {
		t.Fatalf("unexpected occurrence from get-by-uuid: %#v", byUUID.Occurrence)
	}

	if _, err := client.GetOccurrenceByID(context.Background(), 0); err == nil {
		t.Fatalf("expected empty-id error")
	}
	if _, err := client.GetOccurrenceByUUID(context.Background(), "   "); err == nil {
		t.Fatalf("expected empty-uuid error")
	}
}

func TestGetOccurrenceErrorCases(t *testing.T) {
	t.Run("missing token", func(t *testing.T) {
		client := NewClient(Config{BaseURL: "https://api.rollbar.com"})
		if _, err := client.GetOccurrenceByID(context.Background(), 123); err == nil {
			t.Fatalf("expected missing token error")
		}
	})

	t.Run("non-2xx", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "boom", http.StatusBadGateway)
		}))
		defer ts.Close()

		client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
		if _, err := client.GetOccurrenceByID(context.Background(), 123); err == nil || !strings.Contains(err.Error(), "status=502") {
			t.Fatalf("expected non-2xx error, got %v", err)
		}
	})

	t.Run("envelope err", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"err":1,"message":"nope"}`))
		}))
		defer ts.Close()

		client := NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
		if _, err := client.GetOccurrenceByID(context.Background(), 123); err == nil || !strings.Contains(err.Error(), "err=1") {
			t.Fatalf("expected envelope error, got %v", err)
		}
	})
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

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"err":1,"message":"update failed"}`))
	}))
	defer ts.Close()

	client = NewClient(Config{AccessToken: "tok", BaseURL: ts.URL})
	if _, err := client.UpdateItemByID(context.Background(), 1, map[string]any{"status": "active"}); err == nil || !strings.Contains(err.Error(), "update failed") {
		t.Fatalf("expected envelope error, got %v", err)
	}
}

func TestFormatErrorBodyTruncatesAndPrefersMessage(t *testing.T) {
	got := formatErrorBody([]byte(`{"message":"this is a structured error message"}`))
	if got != "this is a structured error message" {
		t.Fatalf("unexpected structured message: %q", got)
	}

	long := strings.Repeat("a", 400)
	got = formatErrorBody([]byte(long))
	if len(got) > 203 || !strings.HasSuffix(got, "...") {
		t.Fatalf("expected truncated body, got %q", got)
	}
}
