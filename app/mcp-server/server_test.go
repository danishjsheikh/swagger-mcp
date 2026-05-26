package mcpserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danishjsheikh/swagger-mcp/app/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// callRequest builds a CallToolRequest with the given arguments.
func callRequest(args map[string]any) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	return req
}

// textContent extracts the text from the first content block.
func textContent(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}
	if tc, ok := result.Content[0].(mcp.TextContent); ok {
		return tc.Text
	}
	return ""
}

// makeServer starts an httptest server that records the request body and
// returns a fixed JSON response. The caller is responsible for closing it.
func makeServer(t *testing.T, capturedBody *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if capturedBody != nil {
			_ = json.Unmarshal(body, capturedBody)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
}

// --- CreateMCPToolHandler: body parameter tests ---

func TestHandler_RequiredBodyParam_Present(t *testing.T) {
	var captured map[string]interface{}
	ts := makeServer(t, &captured)
	defer ts.Close()

	handler := CreateMCPToolHandler(
		nil, nil, nil,
		ts.URL,
		map[string]string{"name": "string"},
		map[string]bool{"name": true},
		"POST",
		nil, nil,
		models.ApiConfig{},
	)

	result, err := handler(context.Background(), callRequest(map[string]any{"name": "alice"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(result))
	}
	if captured["name"] != "alice" {
		t.Errorf("expected name=alice in body, got %v", captured["name"])
	}
}

func TestHandler_RequiredBodyParam_Missing(t *testing.T) {
	ts := makeServer(t, nil)
	defer ts.Close()

	handler := CreateMCPToolHandler(
		nil, nil, nil,
		ts.URL,
		map[string]string{"name": "string"},
		map[string]bool{"name": true},
		"POST",
		nil, nil,
		models.ApiConfig{},
	)

	result, err := handler(context.Background(), callRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing required body param")
	}
	if !strings.Contains(textContent(result), "name") {
		t.Errorf("error message should mention param name, got: %s", textContent(result))
	}
}

func TestHandler_OptionalBodyParam_Missing(t *testing.T) {
	var captured map[string]interface{}
	ts := makeServer(t, &captured)
	defer ts.Close()

	handler := CreateMCPToolHandler(
		nil, nil, nil,
		ts.URL,
		map[string]string{"note": "string"},
		map[string]bool{"note": false},
		"POST",
		nil, nil,
		models.ApiConfig{},
	)

	result, err := handler(context.Background(), callRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("optional missing param should not error, got: %s", textContent(result))
	}
	if _, present := captured["note"]; present {
		t.Error("optional missing param should not appear in request body")
	}
}

func TestHandler_OptionalBodyParam_Present(t *testing.T) {
	var captured map[string]interface{}
	ts := makeServer(t, &captured)
	defer ts.Close()

	handler := CreateMCPToolHandler(
		nil, nil, nil,
		ts.URL,
		map[string]string{"note": "string"},
		map[string]bool{"note": false},
		"POST",
		nil, nil,
		models.ApiConfig{},
	)

	result, err := handler(context.Background(), callRequest(map[string]any{"note": "hello"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", textContent(result))
	}
	if captured["note"] != "hello" {
		t.Errorf("expected note=hello in body, got %v", captured["note"])
	}
}

func TestHandler_MixedRequiredOptional(t *testing.T) {
	var captured map[string]interface{}
	ts := makeServer(t, &captured)
	defer ts.Close()

	handler := CreateMCPToolHandler(
		nil, nil, nil,
		ts.URL,
		map[string]string{"required_field": "string", "optional_field": "string"},
		map[string]bool{"required_field": true, "optional_field": false},
		"POST",
		nil, nil,
		models.ApiConfig{},
	)

	// Only required field provided — should succeed and not include optional
	result, err := handler(context.Background(), callRequest(map[string]any{"required_field": "yes"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", textContent(result))
	}
	if captured["required_field"] != "yes" {
		t.Errorf("expected required_field=yes, got %v", captured["required_field"])
	}
	if _, present := captured["optional_field"]; present {
		t.Error("optional_field should not appear when not provided")
	}
}

// --- Type conversion tests ---

func TestHandler_TypeConversion_Integer(t *testing.T) {
	var captured map[string]interface{}
	ts := makeServer(t, &captured)
	defer ts.Close()

	handler := CreateMCPToolHandler(
		nil, nil, nil,
		ts.URL,
		map[string]string{"count": "integer"},
		map[string]bool{"count": true},
		"POST", nil, nil, models.ApiConfig{},
	)

	result, _ := handler(context.Background(), callRequest(map[string]any{"count": "42"}))
	if result.IsError {
		t.Fatalf("unexpected error: %s", textContent(result))
	}
	// JSON numbers unmarshal as float64
	if captured["count"] != float64(42) {
		t.Errorf("expected count=42, got %v (%T)", captured["count"], captured["count"])
	}
}

func TestHandler_TypeConversion_Boolean(t *testing.T) {
	var captured map[string]interface{}
	ts := makeServer(t, &captured)
	defer ts.Close()

	handler := CreateMCPToolHandler(
		nil, nil, nil,
		ts.URL,
		map[string]string{"active": "boolean"},
		map[string]bool{"active": true},
		"POST", nil, nil, models.ApiConfig{},
	)

	result, _ := handler(context.Background(), callRequest(map[string]any{"active": "true"}))
	if result.IsError {
		t.Fatalf("unexpected error: %s", textContent(result))
	}
	if captured["active"] != true {
		t.Errorf("expected active=true, got %v", captured["active"])
	}
}

func TestHandler_TypeConversion_Float(t *testing.T) {
	var captured map[string]interface{}
	ts := makeServer(t, &captured)
	defer ts.Close()

	handler := CreateMCPToolHandler(
		nil, nil, nil,
		ts.URL,
		map[string]string{"ratio": "float"},
		map[string]bool{"ratio": true},
		"POST", nil, nil, models.ApiConfig{},
	)

	result, _ := handler(context.Background(), callRequest(map[string]any{"ratio": "3.14"}))
	if result.IsError {
		t.Fatalf("unexpected error: %s", textContent(result))
	}
	if captured["ratio"] != 3.14 {
		t.Errorf("expected ratio=3.14, got %v", captured["ratio"])
	}
}

func TestHandler_TypeConversion_InvalidInteger(t *testing.T) {
	ts := makeServer(t, nil)
	defer ts.Close()

	handler := CreateMCPToolHandler(
		nil, nil, nil,
		ts.URL,
		map[string]string{"count": "integer"},
		map[string]bool{"count": true},
		"POST", nil, nil, models.ApiConfig{},
	)

	result, _ := handler(context.Background(), callRequest(map[string]any{"count": "not-a-number"}))
	if !result.IsError {
		t.Fatal("expected error for invalid integer")
	}
}

// --- LoadSwaggerServer: tool registration tests ---

func TestLoadSwaggerServer_OpenAPI3_InlineRequestBody(t *testing.T) {
	spec := models.SwaggerSpec{
		OpenAPI: "3.0.0",
		Servers: []models.Server{{URL: "http://example.com"}},
		Paths: map[string]map[string]models.Endpoint{
			"/items": {
				"post": {
					OperationID: "createItem",
					Parameters:  []models.Parameter{},
					RequestBody: &models.RequestBody{
						Content: map[string]models.MediaTypeObject{
							"application/json": {
								Schema: &models.SchemaRef{
									Type: "object",
									Properties: map[string]models.Property{
										"name":  {Type: "string"},
										"count": {Type: "integer"},
									},
									Required: []string{"name"},
								},
							},
						},
					},
					Responses: map[string]models.Response{},
				},
			},
		},
	}

	mcpServer := server.NewMCPServer("test", "1.0.0")
	// Should not panic
	LoadSwaggerServer(mcpServer, spec, models.ApiConfig{BaseUrl: "http://example.com"})
}

func TestLoadSwaggerServer_OpenAPI3_RefRequestBody(t *testing.T) {
	spec := models.SwaggerSpec{
		OpenAPI: "3.0.0",
		Servers: []models.Server{{URL: "http://example.com"}},
		Components: &models.Components{
			Schemas: map[string]models.Definition{
				"CreateRequest": {
					Type: "object",
					Properties: map[string]models.Property{
						"title":       {Type: "string"},
						"description": {Type: "string"},
					},
					Required: []string{"title"},
				},
			},
		},
		Paths: map[string]map[string]models.Endpoint{
			"/posts": {
				"post": {
					OperationID: "createPost",
					Parameters:  []models.Parameter{},
					RequestBody: &models.RequestBody{
						Content: map[string]models.MediaTypeObject{
							"application/json": {
								Schema: &models.SchemaRef{
									Ref: "#/components/schemas/CreateRequest",
								},
							},
						},
					},
					Responses: map[string]models.Response{},
				},
			},
		},
	}

	mcpServer := server.NewMCPServer("test", "1.0.0")
	LoadSwaggerServer(mcpServer, spec, models.ApiConfig{BaseUrl: "http://example.com"})
}

func TestLoadSwaggerServer_Swagger2_BodyParam(t *testing.T) {
	spec := models.SwaggerSpec{
		Swagger:  "2.0",
		Host:     "example.com",
		BasePath: "/api",
		Definitions: map[string]models.Definition{
			"CreateBody": {
				Type: "object",
				Properties: map[string]models.Property{
					"email":    {Type: "string"},
					"password": {Type: "string"},
				},
				Required: []string{"email", "password"},
			},
		},
		Paths: map[string]map[string]models.Endpoint{
			"/users": {
				"post": {
					OperationID: "createUser",
					Parameters: []models.Parameter{
						{
							In:     "body",
							Name:   "body",
							Schema: &models.SchemaRef{Ref: "#/definitions/CreateBody"},
						},
					},
					Responses: map[string]models.Response{},
				},
			},
		},
	}

	mcpServer := server.NewMCPServer("test", "1.0.0")
	LoadSwaggerServer(mcpServer, spec, models.ApiConfig{BaseUrl: "http://example.com"})
}

// --- OpenAPI 3.0 end-to-end: required/optional enforcement via handler ---

func TestLoadSwaggerServer_OpenAPI3_RequiredOptionalEnforced(t *testing.T) {
	var captured map[string]interface{}
	ts := makeServer(t, &captured)
	defer ts.Close()

	// Build the handler directly mirroring what LoadSwaggerServer produces for an
	// inline requestBody with one required and one optional field.
	handler := CreateMCPToolHandler(
		nil, nil, nil,
		ts.URL,
		map[string]string{"title": "string", "description": "string"},
		map[string]bool{"title": true, "description": false},
		"POST", nil, nil, models.ApiConfig{},
	)

	// Missing required field → error
	result, _ := handler(context.Background(), callRequest(map[string]any{"description": "optional only"}))
	if !result.IsError {
		t.Fatal("expected error when required field 'title' is missing")
	}

	// Provide required, omit optional → success, optional absent from body
	captured = nil
	result, _ = handler(context.Background(), callRequest(map[string]any{"title": "Hello"}))
	if result.IsError {
		t.Fatalf("unexpected error: %s", textContent(result))
	}
	if captured["title"] != "Hello" {
		t.Errorf("expected title=Hello in body, got %v", captured["title"])
	}
	if _, present := captured["description"]; present {
		t.Error("optional field description should be absent when not provided")
	}

	// Provide both → success, both in body
	captured = nil
	result, _ = handler(context.Background(), callRequest(map[string]any{"title": "Hello", "description": "World"}))
	if result.IsError {
		t.Fatalf("unexpected error: %s", textContent(result))
	}
	if captured["title"] != "Hello" || captured["description"] != "World" {
		t.Errorf("both fields should be present, got %v", captured)
	}
}
