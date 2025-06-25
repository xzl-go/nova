package swagger

import (
	"encoding/json"
	"net/http"
	"sync"
)

// RouteInfo 路由信息
type RouteInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
}

// SwaggerDoc OpenAPI 3.0 文档结构
type SwaggerDoc struct {
	OpenAPI string                            `json:"openapi"`
	Info    map[string]interface{}            `json:"info"`
	Paths   map[string]map[string]interface{} `json:"paths"`
}

var (
	routes   []RouteInfo
	routesMu sync.RWMutex
)

// RegisterRoute 注册路由信息
func RegisterRoute(method, path, summary, description string) {
	routesMu.Lock()
	defer routesMu.Unlock()
	routes = append(routes, RouteInfo{
		Method:      method,
		Path:        path,
		Summary:     summary,
		Description: description,
	})
}

// GenerateDoc 生成 OpenAPI 3.0 文档
func GenerateDoc() *SwaggerDoc {
	doc := &SwaggerDoc{
		OpenAPI: "3.0.0",
		Info: map[string]interface{}{
			"title":   "Nova API",
			"version": "1.0.0",
		},
		Paths: make(map[string]map[string]interface{}),
	}
	routesMu.RLock()
	defer routesMu.RUnlock()
	for _, r := range routes {
		if doc.Paths[r.Path] == nil {
			doc.Paths[r.Path] = make(map[string]interface{})
		}
		doc.Paths[r.Path][r.Method] = map[string]interface{}{
			"summary":     r.Summary,
			"description": r.Description,
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "OK",
				},
			},
		}
	}
	return doc
}

// SwaggerDocHandler 返回 OpenAPI 文档 JSON
func SwaggerDocHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GenerateDoc())
}

// SwaggerUIHandler 返回 Swagger UI 页面
func SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Nova Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: '/swagger/doc.json',
        dom_id: '#swagger-ui',
      });
    };
  </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
