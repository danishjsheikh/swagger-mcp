package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	mcpserver "github.com/danishjsheikh/swagger-mcp/app/mcp-server"
	"github.com/danishjsheikh/swagger-mcp/app/models"
	"github.com/danishjsheikh/swagger-mcp/app/swagger"
)

func getHttpAddr(httpAddr string) string {
	if httpAddr == "" {
		return "localhost:8080"
	}
	// ":Port" or "IP:Port"
	if strings.HasPrefix(httpAddr, ":") {
		// sseUrl = http://localhost:Port
		return "http://localhost" + httpAddr
	}
	if !strings.Contains(httpAddr, ":") {
		log.Fatal("sseAddr must be in :Port or IP:Port format")
	}
	return httpAddr
}

func getSseUrlAddr(sseUrl, sseAddr string) (string, string) {
	// Only complement if one is empty; if both are set, use as-is
	if sseAddr == "" && sseUrl == "" {
		return "http://localhost:8080", "localhost:8080"
	}
	if sseAddr != "" {
		// ":Port" or "IP:Port"
		if strings.HasPrefix(sseAddr, ":") {
			// sseUrl = http://localhost:Port
			return "http://localhost" + sseAddr, sseAddr
		}
		if !strings.Contains(sseAddr, ":") {
			log.Fatal("sseAddr must be in :Port or IP:Port format")
		}
		return "http://" + sseAddr, sseAddr
	} else if sseUrl != "" {
		u, err := url.Parse(sseUrl)
		if err != nil {
			log.Fatalf("Invalid sseUrl: %v", err)
		}
		host := u.Host
		port := ""
		if strings.Contains(host, ":") {
			parts := strings.Split(host, ":")
			host = parts[0]
			port = parts[1]
		}
		// 没有端口时根据 scheme 补全
		if port == "" {
			switch u.Scheme {
			case "http":
				port = "80"
			case "https":
				port = "443"
			default:
				log.Fatalf("Unknown scheme for sseUrl: %s", u.Scheme)
			}
		}
		return sseUrl, host + ":" + port
	} else {
		log.Fatal("Either sseAddr or sseUrl must be provided")
	}
	return "", ""
}

func main() {
	var finalSseUrl, finalSseAddr string
	var finalHttpPath, finalHttpAddr string
	specUrl := flag.String("specUrl", "", "URL of the Swagger JSON specification")
	sseMode := flag.Bool("sse", false, "Run in SSE mode instead of stdio mode")
	sseAddr := flag.String("sseAddr", "", "SSE server listen address in :Port or IP:Port format")
	sseUrl := flag.String("sseUrl", "", "Base URL for the SSE server")
	httpMode := flag.Bool("http", false, "Run in StreamableHTTP mode instead of stdio mode")
	httpAddr := flag.String("httpAddr", "", "StreamableHTTP server listen address in :Port or IP:Port format")
	httpPath := flag.String("httpPath", "", "Endpoint path for the StreamableHTTP server")
	baseUrl := flag.String("baseUrl", "", "Base URL for API requests")
	includePaths := flag.String("includePaths", "", "Comma-separated list of paths or regex to include")
	excludePaths := flag.String("excludePaths", "", "Comma-separated list of paths or regex to exclude")
	includeMethods := flag.String("includeMethods", "", "Comma-separated list of HTTP methods to include")
	excludeMethods := flag.String("excludeMethods", "", "Comma-separated list of HTTP methods to exclude")
	security := flag.String("security", "", "API security type: basic, apiKey, or bearer")
	basicAuth := flag.String("basicAuth", "", "Basic auth credentials in user:password format, used in Authorization header")
	bearerAuth := flag.String("bearerAuth", "", "Bearer token for Authorization header")
	apiKeyAuth := flag.String("apiKeyAuth", "", "API key auth, format: 'passAs:name=value', passAs=header/query/cookie, multiple by comma")
	headers := flag.String("headers", "", "Additional headers to include in requests (format: name1=value1,name2=value2)")
	sseHeaders := flag.String("sseHeaders", "", "Read headers from sse request, and pass to API request (format: name1,name2)")
	httpHeaders := flag.String("httpHeaders", "", "Read headers from StreamableHTTP request, and pass to API request (format: name1,name2)")

	flag.Parse()

	// Validate spec
	if *specUrl == "" {
		log.Fatal("Please provide the Swagger JSON URL or file path using the --specUrl flag")
	}

	if strings.HasPrefix(*specUrl, "http://") || strings.HasPrefix(*specUrl, "https://") {
		_, err := url.ParseRequestURI(*specUrl)
		if err != nil {
			log.Fatalf("Invalid spec URL: %v", err)
		}
	} else if strings.HasPrefix(*specUrl, "file://") {
		filePath := strings.TrimPrefix(*specUrl, "file://")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Fatalf("Spec file does not exist: %v", err)
		}
	} else {
		log.Fatal("Invalid specUrl format. Must be a valid HTTP URL or file:// path")
	}

	// Validate baseUrl
	if *baseUrl != "" {
		if !strings.HasPrefix(*baseUrl, "http://") && !strings.HasPrefix(*baseUrl, "https://") {
			log.Fatal("baseUrl must start with http:// or https://")
		}
	}
	if *sseMode && *httpMode {
		log.Fatal("Cannot run in both SSE and StreamableHTTP modes")
	}

	if *sseMode { // get final sseAddr and sseUrl
		finalSseUrl, finalSseAddr = getSseUrlAddr(*sseUrl, *sseAddr)
	}
	if *httpMode {
		if *httpPath == "" {
			finalHttpPath = "/mcp"
		}
		finalHttpAddr = getHttpAddr(*httpAddr)
	}
	swaggerSpec, err := swagger.LoadSwagger(*specUrl)
	if err != nil {
		log.Fatalf("Failed to load Swagger spec: %v", err)
	}
	swagger.ExtractSwagger(swaggerSpec)

	config := models.Config{
		SpecUrl: *specUrl,
		SseCfg: models.SseConfig{
			SseMode:    *sseMode,
			SseAddr:    finalSseAddr,
			SseUrl:     finalSseUrl,
			SseHeaders: *sseHeaders,
		},
		HttpCfg: models.HttpConfig{
			HttpMode:    *httpMode,
			HttpAddr:    finalHttpAddr,
			HttpPath:    finalHttpPath,
			HttpHeaders: *httpHeaders,
		},
		ApiCfg: models.ApiConfig{
			BaseUrl:        *baseUrl,
			IncludePaths:   *includePaths,
			ExcludePaths:   *excludePaths,
			IncludeMethods: *includeMethods,
			ExcludeMethods: *excludeMethods,
			Security:       *security,
			BasicAuth:      *basicAuth,
			ApiKeyAuth:     *apiKeyAuth,
			BearerAuth:     *bearerAuth,
			Headers:        *headers,
		},
	}

	if *sseMode {
		fmt.Printf("Starting server with specUrl: %s, SSE mode, SSE URL: %s, SSE Addr: %s, Base URL: %s\n",
			config.SpecUrl, config.SseCfg.SseUrl, config.SseCfg.SseAddr, config.ApiCfg.BaseUrl)

	} else if *httpMode {
		fmt.Printf("Starting server with specUrl: %s, StreamableHTTP mode, HTTP URL: %s, HTTP Addr: %s, Base URL: %s\n",
			config.SpecUrl, config.SseCfg.SseUrl, config.SseCfg.SseAddr, config.ApiCfg.BaseUrl)

	} else {
		fmt.Printf("Starting server with specUrl: %s, Stdio mode.\n", config.SpecUrl)
	}
	mcpserver.CreateServer(swaggerSpec, config)
}
