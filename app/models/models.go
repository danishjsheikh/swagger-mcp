package models

type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type SwaggerSpec struct {
	// Swagger 2.0 fields
	Host     string `json:"host,omitempty"`
	BasePath string `json:"basePath,omitempty"`
	Swagger  string `json:"swagger,omitempty"`

	// OpenAPI 3.0 fields
	OpenAPI    string      `json:"openapi,omitempty"`
	Servers    []Server    `json:"servers,omitempty"`
	Components *Components `json:"components,omitempty"`

	// Common fields
	Paths       map[string]map[string]Endpoint `json:"paths"`
	Definitions map[string]Definition          `json:"definitions,omitempty"` // Swagger 2.0
}

type Components struct {
	Schemas map[string]Definition `json:"schemas,omitempty"` // OpenAPI 3.0
}

type Definition struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
}

type Property struct {
	Type string `json:"type"`
}

type Endpoint struct {
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Parameters  []Parameter         `json:"parameters"`
	Responses   map[string]Response `json:"responses"`
	Consumes    []string            `json:"consumes"`
	Produces    []string            `json:"produces"`
}

type Parameter struct {
	Name        string     `json:"name"`
	In          string     `json:"in"`
	Required    bool       `json:"required"`
	Type        string     `json:"type"`
	Schema      *SchemaRef `json:"schema,omitempty"`
	Description string     `json:"description"`
}

type Response struct {
	Description string     `json:"description"`
	Schema      *SchemaRef `json:"schema,omitempty"`
	Type        string     `json:"type,omitempty"`
}

type SchemaRef struct {
	Ref  string `json:"$ref,omitempty"`
	Type string `json:"type,omitempty"`
}

// SseConfig stores SSE (Server-Sent Events) related parameters
type SseConfig struct {
	SseMode    bool   `json:"sseMode"`    // Whether to run in SSE mode
	SseAddr    string `json:"sseAddr"`    // SSE server listen address
	SseUrl     string `json:"sseUrl"`     // Base URL for the SSE server
	SseHeaders string `json:"sseHeaders"` // Read headers from sse request, and pass to API request (format: name1,name2)
}

// HttpConfig stores StreamableHTTP related parameters
type HttpConfig struct {
	HttpMode    bool   `json:"httpMode"`    // Whether to run in StreamableHTTP mode
	HttpAddr    string `json:"httpAddr"`    // StreamableHTTP server listen address
	HttpPath    string `json:"httpPath"`    // Base URL for the StreamableHTTP server
	HttpHeaders string `json:"httpHeaders"` // Read headers from StreamableHTTP request, and pass to API request (format: name1,name2)
}

// ApiConfig stores API related parameters
type ApiConfig struct {
	BaseUrl        string `json:"baseUrl"`        // Base URL for API requests
	IncludePaths   string `json:"includePaths"`   // List of paths or regex patterns to include
	ExcludePaths   string `json:"excludePaths"`   // List of paths or regex patterns to exclude
	IncludeMethods string `json:"includeMethods"` // List of HTTP methods to include
	ExcludeMethods string `json:"excludeMethods"` // List of HTTP methods to exclude
	Security       string `json:"security"`       // API security type
	BasicAuth      string `json:"basicAuth"`      // Basic auth credentials
	ApiKeyAuth     string `json:"apiKeyAuth"`     // API key authentication information
	BearerAuth     string `json:"bearerAuth"`     // Bearer token
	Headers        string `json:"headers"`        // Additional headers to include in requests (format: name1=value1,name2=value2)
}

// Config stores all command line parameters
type Config struct {
	SpecUrl string     `json:"specUrl"` // URL of the Swagger JSON specification
	SseCfg  SseConfig  `json:"sseCfg"`  // SSE related configuration
	HttpCfg HttpConfig `json:"httpCfg"` // StreamableHTTP related configuration
	ApiCfg  ApiConfig  `json:"apiCfg"`  // API related configuration
}
