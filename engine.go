package nova

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/xzl-go/nova/tree"

	"github.com/xzl-go/nova/logger"
	"go.uber.org/zap"
)

// Engine 框架引擎
type Engine struct {
	router *tree.Node
	groups []*RouterGroup
}

// RouterGroup 路由组
type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	engine      *Engine
}

// HandlerFunc 处理函数类型
type HandlerFunc func(*Context)

// handlerAdapter 处理函数适配器
type handlerAdapter struct {
	handler HandlerFunc
}

func (h *handlerAdapter) Handle(ctx interface{}) {
	h.handler(ctx.(*Context))
}

// NewEngine 创建新引擎
func NewEngine() *Engine {
	engine := &Engine{
		router: tree.NewNode(),
	}
	engine.groups = []*RouterGroup{{engine: engine}}
	return engine
}

// Use 添加中间件
func (e *Engine) Use(middlewares ...HandlerFunc) {
	e.groups[0].Use(middlewares...)
}

// GET 添加 GET 路由
func (e *Engine) GET(pattern string, handlers ...HandlerFunc) {
	e.groups[0].GET(pattern, handlers...)
}

// POST 添加 POST 路由
func (e *Engine) POST(pattern string, handlers ...HandlerFunc) {
	e.groups[0].POST(pattern, handlers...)
}

// PUT 添加 PUT 路由
func (e *Engine) PUT(pattern string, handlers ...HandlerFunc) {
	e.groups[0].PUT(pattern, handlers...)
}

// DELETE 添加 DELETE 路由
func (e *Engine) DELETE(pattern string, handlers ...HandlerFunc) {
	e.groups[0].DELETE(pattern, handlers...)
}

// PATCH 添加 PATCH 路由
func (e *Engine) PATCH(pattern string, handlers ...HandlerFunc) {
	e.groups[0].PATCH(pattern, handlers...)
}

// OPTIONS 添加 OPTIONS 路由
func (e *Engine) OPTIONS(pattern string, handlers ...HandlerFunc) {
	e.groups[0].OPTIONS(pattern, handlers...)
}

// HEAD 添加 HEAD 路由
func (e *Engine) HEAD(pattern string, handlers ...HandlerFunc) {
	e.groups[0].HEAD(pattern, handlers...)
}

// Group 创建路由组
func (e *Engine) Group(prefix string) *RouterGroup {
	return e.groups[0].Group(prefix)
}

// Group 创建路由组
func (g *RouterGroup) Group(prefix string) *RouterGroup {
	engine := g.engine
	newGroup := &RouterGroup{
		prefix:      g.prefix + prefix,
		middlewares: make([]HandlerFunc, len(g.middlewares)),
		engine:      engine,
	}
	copy(newGroup.middlewares, g.middlewares)
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// Use 添加中间件
func (g *RouterGroup) Use(middlewares ...HandlerFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

// addRoute 添加路由
func (g *RouterGroup) addRoute(method string, pattern string, handlers ...HandlerFunc) {
	pattern = g.prefix + pattern
	parts := parsePattern(pattern)

	// 转换处理函数为适配器
	adapters := make([]tree.Handler, len(handlers))
	for i, handler := range handlers {
		adapters[i] = &handlerAdapter{handler: handler}
	}

	g.engine.router.Insert(pattern, parts, 0, adapters)
}

// GET 添加 GET 路由
func (g *RouterGroup) GET(pattern string, handlers ...HandlerFunc) {
	g.addRoute("GET", pattern, handlers...)
}

// POST 添加 POST 路由
func (g *RouterGroup) POST(pattern string, handlers ...HandlerFunc) {
	g.addRoute("POST", pattern, handlers...)
}

// PUT 添加 PUT 路由
func (g *RouterGroup) PUT(pattern string, handlers ...HandlerFunc) {
	g.addRoute("PUT", pattern, handlers...)
}

// DELETE 添加 DELETE 路由
func (g *RouterGroup) DELETE(pattern string, handlers ...HandlerFunc) {
	g.addRoute("DELETE", pattern, handlers...)
}

// PATCH 添加 PATCH 路由
func (g *RouterGroup) PATCH(pattern string, handlers ...HandlerFunc) {
	g.addRoute("PATCH", pattern, handlers...)
}

// OPTIONS 添加 OPTIONS 路由
func (g *RouterGroup) OPTIONS(pattern string, handlers ...HandlerFunc) {
	g.addRoute("OPTIONS", pattern, handlers...)
}

// HEAD 添加 HEAD 路由
func (g *RouterGroup) HEAD(pattern string, handlers ...HandlerFunc) {
	g.addRoute("HEAD", pattern, handlers...)
}

// Run 启动服务器
func (e *Engine) Run(addr string) error {
	println("  _   _  ___  __   __  ___ ")
	println(" | \\ | |/ _ \\ \\ \\ / / / _ \\")
	println(" |  \\| | | | | \\ v / | | | |")
	println(" | |\\  | |_| |  | |  | |_| |")
	println(" |_| \\_|\\___/   |_|   \\___/ ")
	println(" nova server is running on http://" + addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return server.ListenAndServe()
}

// ServeHTTP 实现 http.Handler 接口
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := GetContext(w, r)
	defer PutContext(c)

	// 添加请求上下文
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	c.Request = c.Request.WithContext(ctx)

	e.handle(c)
}

// handle 处理请求
func (e *Engine) handle(c *Context) {
	parts := parsePattern(c.Request.URL.Path)
	node := e.router.Search(parts, 0)
	middlewares := e.groups[0].middlewares

	if node != nil {
		c.Params = node.GetParams(c.Request.URL.Path)
		// 正确合并全局中间件和路由 handler
		c.handlers = make([]HandlerFunc, 0, len(middlewares)+len(node.Handlers))
		c.handlers = append(c.handlers, middlewares...)
		for _, handler := range node.Handlers {
			c.handlers = append(c.handlers, handler.(*handlerAdapter).handler)
		}
	} else {
		// 404处理
		c.handlers = []HandlerFunc{func(c *Context) {
			c.JSON(http.StatusNotFound, map[string]interface{}{
				"code":    404,
				"message": "Not Found",
				"path":    c.Request.URL.Path,
			})
		}}
	}

	// 添加错误恢复
	defer func() {
		if err := recover(); err != nil {
			logger.Error("panic recovered",
				zap.Any("error", err),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "Internal Server Error",
				"error":   fmt.Sprintf("%v", err),
			})
		}
	}()

	c.Next()
}

// parsePattern 解析路由模式
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}
