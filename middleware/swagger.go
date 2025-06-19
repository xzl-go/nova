package middleware

import (
	"framework/router"
	"net/http"
)

// Swagger Swagger 文档中间件
func Swagger() router.HandlerFunc {
	return func(c *router.Context) {
		// 返回 Swagger UI HTML
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html>
<head>
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4/swagger-ui.css">
    <script src="https://unpkg.com/swagger-ui-dist@4/swagger-ui-bundle.js"></script>
</head>
<body>
    <div id="swagger-ui"></div>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/swagger.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
            });
        }
    </script>
</body>
</html>
`)
		c.Abort()
	}
}

// SwaggerInfo Swagger 信息
type SwaggerInfo struct {
	Title       string
	Description string
	Version     string
	Host        string
	BasePath    string
}

// InitSwagger 初始化 Swagger 配置
func InitSwagger(info SwaggerInfo) {
	// @title           API 文档
	// @version         1.0
	// @description     API 接口文档
	// @termsOfService  http://swagger.io/terms/

	// @contact.name   API Support
	// @contact.url    http://www.swagger.io/support
	// @contact.email  support@swagger.io

	// @license.name  Apache 2.0
	// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

	// @host      localhost:8080
	// @BasePath  /api/v1

	// @securityDefinitions.apikey Bearer
	// @in header
	// @name Authorization
	// @description Type "Bearer" followed by a space and JWT token.
}
