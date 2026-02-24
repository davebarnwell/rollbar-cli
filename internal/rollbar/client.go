package rollbar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AccessToken string
	BaseURL     string
	Timeout     time.Duration
}

type Client struct {
	accessToken string
	baseURL     string
	httpClient  *http.Client
}

type ListItemsOptions struct {
	Page        int
	Status      string
	Environment string
	Level       []string
}

type Item struct {
	ID                      int64
	Counter                 int64
	Title                   string
	Level                   string
	Status                  string
	Environment             string
	TotalOccurrences        int64
	LastOccurrenceTimestamp int64
}

type StackFrame struct {
	Filename string
	Line     int64
	Method   string
}

type ItemInstance struct {
	ID          int64
	UUID        string
	Level       string
	Environment string
	Timestamp   int64
	StackFrames []StackFrame
	Payload     map[string]any
}

type ListItemsResponse struct {
	Items []Item
	Raw   map[string]any
}

type GetItemResponse struct {
	Item Item
	Raw  map[string]any
}

type ListItemInstancesResponse struct {
	Instances []ItemInstance
	Raw       map[string]any
}

type UpdateItemResponse struct {
	Item Item
	Raw  map[string]any
}

type apiEnvelope struct {
	Err     int             `json:"err"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

type listItemsResult struct {
	Items []json.RawMessage `json:"items"`
}

type listItemInstancesResult struct {
	Instances []json.RawMessage `json:"instances"`
}

func NewClient(cfg Config) *Client {
	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.rollbar.com"
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Client{
		accessToken: cfg.AccessToken,
		baseURL:     baseURL,
		httpClient:  &http.Client{Timeout: timeout},
	}
}

func (c *Client) ListItems(ctx context.Context, opts ListItemsOptions) (*ListItemsResponse, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}

	endpoint, err := url.Parse(c.baseURL + "/api/1/items")
	if err != nil {
		return nil, fmt.Errorf("build request URL: %w", err)
	}

	query := endpoint.Query()
	if opts.Page > 0 {
		query.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Status != "" {
		query.Set("status", opts.Status)
	}
	if opts.Environment != "" {
		query.Set("environment", opts.Environment)
	}
	for _, level := range opts.Level {
		if strings.TrimSpace(level) != "" {
			query.Add("level", level)
		}
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Rollbar-Access-Token", c.accessToken)
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("rollbar API error: status=%d body=%s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("parse envelope: %w", err)
	}
	if env.Err != 0 {
		if env.Message == "" {
			env.Message = "unknown error"
		}
		return nil, fmt.Errorf("rollbar API returned err=%d: %s", env.Err, env.Message)
	}

	var result listItemsResult
	if len(env.Result) > 0 {
		if err := json.Unmarshal(env.Result, &result); err != nil {
			return nil, fmt.Errorf("parse result.items: %w", err)
		}
	}

	items := make([]Item, 0, len(result.Items))
	for _, rawItem := range result.Items {
		items = append(items, normalizeItem(rawItem))
	}

	return &ListItemsResponse{Items: items, Raw: raw}, nil
}

func (c *Client) GetItemByID(ctx context.Context, id int64) (*GetItemResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid item id: must be > 0")
	}
	return c.getItem(ctx, strconv.FormatInt(id, 10))
}

func (c *Client) GetItemByUUID(ctx context.Context, uuid string) (*GetItemResponse, error) {
	uuid = strings.TrimSpace(uuid)
	if uuid == "" {
		return nil, fmt.Errorf("invalid UUID: must not be empty")
	}
	return c.getItem(ctx, uuid)
}

func (c *Client) ListItemInstances(ctx context.Context, identifier string, page int) (*ListItemInstancesResponse, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, fmt.Errorf("invalid item identifier: must not be empty")
	}
	if c.accessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}

	endpoint, err := url.Parse(c.baseURL + "/api/1/item/" + url.PathEscape(identifier) + "/instances")
	if err != nil {
		return nil, fmt.Errorf("build request URL: %w", err)
	}
	if page > 0 {
		query := endpoint.Query()
		query.Set("page", strconv.Itoa(page))
		endpoint.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Rollbar-Access-Token", c.accessToken)
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("rollbar API error: status=%d body=%s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("parse envelope: %w", err)
	}
	if env.Err != 0 {
		if env.Message == "" {
			env.Message = "unknown error"
		}
		return nil, fmt.Errorf("rollbar API returned err=%d: %s", env.Err, env.Message)
	}

	var result listItemInstancesResult
	if len(env.Result) > 0 {
		if err := json.Unmarshal(env.Result, &result); err != nil {
			return nil, fmt.Errorf("parse result.instances: %w", err)
		}
		if len(result.Instances) == 0 {
			var directInstances []json.RawMessage
			if err := json.Unmarshal(env.Result, &directInstances); err == nil {
				result.Instances = directInstances
			}
		}
	}

	instances := make([]ItemInstance, 0, len(result.Instances))
	for _, rawInstance := range result.Instances {
		instances = append(instances, normalizeInstance(rawInstance))
	}

	return &ListItemInstancesResponse{
		Instances: instances,
		Raw:       raw,
	}, nil
}

func (c *Client) UpdateItemByID(ctx context.Context, id int64, body map[string]any) (*UpdateItemResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid item id: must be > 0")
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("missing update fields")
	}
	if c.accessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}

	endpoint, err := url.Parse(c.baseURL + "/api/1/item/" + strconv.FormatInt(id, 10))
	if err != nil {
		return nil, fmt.Errorf("build request URL: %w", err)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Rollbar-Access-Token", c.accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("rollbar API error: status=%d body=%s", res.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var raw map[string]any
	if err := json.Unmarshal(responseBody, &raw); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	var env apiEnvelope
	if err := json.Unmarshal(responseBody, &env); err != nil {
		return nil, fmt.Errorf("parse envelope: %w", err)
	}
	if env.Err != 0 {
		if env.Message == "" {
			env.Message = "unknown error"
		}
		return nil, fmt.Errorf("rollbar API returned err=%d: %s", env.Err, env.Message)
	}

	var result map[string]any
	if len(env.Result) > 0 {
		if err := json.Unmarshal(env.Result, &result); err != nil {
			return nil, fmt.Errorf("parse result: %w", err)
		}
	}

	itemData := result
	if v, ok := result["item"]; ok {
		if nested, ok := v.(map[string]any); ok {
			itemData = nested
		}
	}

	return &UpdateItemResponse{
		Item: normalizeItemMap(itemData),
		Raw:  raw,
	}, nil
}

func (c *Client) getItem(ctx context.Context, identifier string) (*GetItemResponse, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}

	endpoint, err := url.Parse(c.baseURL + "/api/1/item/" + url.PathEscape(identifier))
	if err != nil {
		return nil, fmt.Errorf("build request URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Rollbar-Access-Token", c.accessToken)
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("rollbar API error: status=%d body=%s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("parse envelope: %w", err)
	}
	if env.Err != 0 {
		if env.Message == "" {
			env.Message = "unknown error"
		}
		return nil, fmt.Errorf("rollbar API returned err=%d: %s", env.Err, env.Message)
	}

	var result map[string]any
	if len(env.Result) > 0 {
		if err := json.Unmarshal(env.Result, &result); err != nil {
			return nil, fmt.Errorf("parse item result: %w", err)
		}
	}

	itemData := result
	if v, ok := result["item"]; ok {
		if nested, ok := v.(map[string]any); ok {
			itemData = nested
		}
	}

	item := normalizeItemMap(itemData)
	return &GetItemResponse{Item: item, Raw: raw}, nil
}

func normalizeItem(rawItem json.RawMessage) Item {
	var m map[string]any
	if err := json.Unmarshal(rawItem, &m); err != nil {
		return Item{}
	}

	return normalizeItemMap(m)
}

func normalizeItemMap(m map[string]any) Item {
	if m == nil {
		return Item{}
	}

	item := Item{
		ID:                      getInt64(m, "id"),
		Counter:                 getInt64(m, "counter"),
		Title:                   getString(m, "title"),
		Level:                   getString(m, "level"),
		Status:                  getString(m, "status"),
		Environment:             getString(m, "environment"),
		TotalOccurrences:        getInt64(m, "total_occurrences"),
		LastOccurrenceTimestamp: getInt64(m, "last_occurrence_timestamp"),
	}

	if item.Title == "" {
		item.Title = getString(m, "last_occurrence", "body", "trace", "exception", "message")
	}
	if item.Level == "" {
		item.Level = getString(m, "last_occurrence", "level")
	}
	if item.Environment == "" {
		item.Environment = getString(m, "last_occurrence", "environment")
	}
	if item.LastOccurrenceTimestamp == 0 {
		item.LastOccurrenceTimestamp = getInt64(m, "last_occurrence", "timestamp")
	}

	return item
}

func normalizeInstance(rawInstance json.RawMessage) ItemInstance {
	var m map[string]any
	if err := json.Unmarshal(rawInstance, &m); err != nil {
		return ItemInstance{}
	}
	return normalizeInstanceMap(m)
}

func normalizeInstanceMap(m map[string]any) ItemInstance {
	if m == nil {
		return ItemInstance{}
	}

	instance := ItemInstance{
		ID:          getInt64(m, "id"),
		UUID:        getString(m, "uuid"),
		Level:       getString(m, "level"),
		Environment: getString(m, "environment"),
		Timestamp:   getInt64(m, "timestamp"),
		StackFrames: extractStackFrames(m),
		Payload:     extractPayload(m),
	}

	return instance
}

func extractStackFrames(instance map[string]any) []StackFrame {
	body, ok := instance["body"].(map[string]any)
	if !ok {
		return nil
	}

	frames := make([]StackFrame, 0)

	if trace, ok := body["trace"].(map[string]any); ok {
		frames = append(frames, extractTraceFrames(trace)...)
	}

	if chain, ok := body["trace_chain"].([]any); ok {
		for _, rawTrace := range chain {
			trace, ok := rawTrace.(map[string]any)
			if !ok {
				continue
			}
			frames = append(frames, extractTraceFrames(trace)...)
		}
	}

	return frames
}

func extractTraceFrames(trace map[string]any) []StackFrame {
	rawFrames, ok := trace["frames"].([]any)
	if !ok {
		return nil
	}

	frames := make([]StackFrame, 0, len(rawFrames))
	for _, rawFrame := range rawFrames {
		frameData, ok := rawFrame.(map[string]any)
		if !ok {
			continue
		}

		frame := StackFrame{
			Filename: firstString(frameData, "filename", "abs_path", "path", "file"),
			Line:     firstInt64(frameData, "lineno", "line", "line_number"),
			Method:   firstString(frameData, "method", "function"),
		}
		if frame.Filename == "" && frame.Line == 0 && frame.Method == "" {
			continue
		}
		frames = append(frames, frame)
	}

	return frames
}

func extractPayload(instance map[string]any) map[string]any {
	payload := make(map[string]any)
	for _, key := range []string{"body", "request", "server", "client", "person", "custom", "data", "notifier"} {
		if value, ok := instance[key]; ok {
			payload[key] = value
		}
	}
	if len(payload) == 0 {
		return nil
	}
	return payload
}

func firstString(data map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if s, ok := value.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func firstInt64(data map[string]any, keys ...string) int64 {
	for _, key := range keys {
		if value, ok := data[key]; ok {
			switch t := value.(type) {
			case float64:
				if t != 0 {
					return int64(t)
				}
			case int64:
				if t != 0 {
					return t
				}
			case int:
				if t != 0 {
					return int64(t)
				}
			case json.Number:
				if n, err := t.Int64(); err == nil && n != 0 {
					return n
				}
			}
		}
	}
	return 0
}

func getString(data map[string]any, path ...string) string {
	v, ok := walk(data, path...)
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}

func getInt64(data map[string]any, path ...string) int64 {
	v, ok := walk(data, path...)
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case int:
		return int64(t)
	case json.Number:
		n, _ := t.Int64()
		return n
	default:
		return 0
	}
}

func walk(data map[string]any, path ...string) (any, bool) {
	var cur any = data
	for _, key := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[key]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}
