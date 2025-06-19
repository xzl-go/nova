package core

import (
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

// 测试路由数据
var testRoutes = []struct {
	path   string
	method string
}{
	// 基础路由
	{"/", "GET"},
	{"/user", "GET"},
	{"/user", "POST"},
	{"/user", "PUT"},
	{"/user", "DELETE"},

	// 带参数路由
	{"/user/:id", "GET"},
	{"/user/:id/profile", "GET"},
	{"/user/:id/posts", "GET"},
	{"/user/:id/posts/:postId", "GET"},
	{"/user/:id/posts/:postId/comments", "GET"},
	{"/user/:id/posts/:postId/comments/:commentId", "GET"},

	// 嵌套路由
	{"/api/v1/users", "GET"},
	{"/api/v1/users/:id", "GET"},
	{"/api/v1/users/:id/profile", "GET"},
	{"/api/v1/users/:id/posts", "GET"},
	{"/api/v1/users/:id/posts/:postId", "GET"},
}

// 测试请求数据
var testRequests = []struct {
	path   string
	method string
}{
	// 基础请求
	{"/", "GET"},
	{"/user", "GET"},
	{"/user", "POST"},
	{"/user", "PUT"},
	{"/user", "DELETE"},

	// 带参数请求
	{"/user/123", "GET"},
	{"/user/123/profile", "GET"},
	{"/user/123/posts", "GET"},
	{"/user/123/posts/456", "GET"},
	{"/user/123/posts/456/comments", "GET"},
	{"/user/123/posts/456/comments/789", "GET"},

	// 嵌套请求
	{"/api/v1/users", "GET"},
	{"/api/v1/users/123", "GET"},
	{"/api/v1/users/123/profile", "GET"},
	{"/api/v1/users/123/posts", "GET"},
	{"/api/v1/users/123/posts/456", "GET"},
}

// mockResponseWriter 用于测试
type mockResponseWriter struct {
	http.ResponseWriter
}

func (m *mockResponseWriter) WriteHeader(code int)        {}
func (m *mockResponseWriter) Write(b []byte) (int, error) { return len(b), nil }
func (m *mockResponseWriter) Header() http.Header         { return make(http.Header) }

// Nova 路由测试
func BenchmarkNovaRouter(b *testing.B) {
	router := NewRouter()
	for _, route := range testRoutes {
		router.AddRoute(route.path, route.method, func(ctx *Context) {})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := testRequests[i%len(testRequests)]
		router.FindRoute(req.method, req.path)
	}
}

// Nova 路由并发测试
func BenchmarkNovaRouterConcurrent(b *testing.B) {
	router := NewRouter()
	for _, route := range testRoutes {
		router.AddRoute(route.path, route.method, func(ctx *Context) {})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := testRequests[i%len(testRequests)]
			router.FindRoute(req.method, req.path)
			i++
		}
	})
}

// Nova 路由内存压力测试
func BenchmarkNovaRouterMemory(b *testing.B) {
	router := NewRouter()
	for _, route := range testRoutes {
		router.AddRoute(route.path, route.method, func(ctx *Context) {})
	}

	var wg sync.WaitGroup
	concurrency := 1000
	iterations := 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(concurrency)
		for j := 0; j < concurrency; j++ {
			go func() {
				defer wg.Done()
				for k := 0; k < iterations; k++ {
					req := testRequests[k%len(testRequests)]
					router.FindRoute(req.method, req.path)
				}
			}()
		}
		wg.Wait()
	}
}

// Gin 路由测试
func BenchmarkGinRouter(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	for _, route := range testRoutes {
		router.Handle(route.method, route.path, func(c *gin.Context) {})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := testRequests[i%len(testRequests)]
		w := &mockResponseWriter{}
		httpReq, _ := http.NewRequest(req.method, req.path, nil)
		router.ServeHTTP(w, httpReq)
	}
}

// Gin 路由并发测试
func BenchmarkGinRouterConcurrent(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	for _, route := range testRoutes {
		router.Handle(route.method, route.path, func(c *gin.Context) {})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := testRequests[i%len(testRequests)]
			w := &mockResponseWriter{}
			httpReq, _ := http.NewRequest(req.method, req.path, nil)
			router.ServeHTTP(w, httpReq)
			i++
		}
	})
}

// Echo 路由测试
func BenchmarkEchoRouter(b *testing.B) {
	e := echo.New()
	for _, route := range testRoutes {
		e.Add(route.method, route.path, func(c echo.Context) error { return nil })
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := testRequests[i%len(testRequests)]
		w := &mockResponseWriter{}
		httpReq, _ := http.NewRequest(req.method, req.path, nil)
		e.ServeHTTP(w, httpReq)
	}
}

// Echo 路由并发测试
func BenchmarkEchoRouterConcurrent(b *testing.B) {
	e := echo.New()
	for _, route := range testRoutes {
		e.Add(route.method, route.path, func(c echo.Context) error { return nil })
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := testRequests[i%len(testRequests)]
			w := &mockResponseWriter{}
			httpReq, _ := http.NewRequest(req.method, req.path, nil)
			e.ServeHTTP(w, httpReq)
			i++
		}
	})
}

// Beego 路由测试
func BenchmarkBeegoRouter(b *testing.B) {
	web.BConfig.RunMode = web.PROD
	web.BConfig.WebConfig.AutoRender = false

	for _, route := range testRoutes {
		web.Router(route.path, &testController{}, "get:Get")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := testRequests[i%len(testRequests)]
		httpReq := &http.Request{
			Method: req.method,
			URL:    &url.URL{Path: req.path},
		}
		ctx := &context.Context{
			Request: httpReq,
			Input:   &context.BeegoInput{},
		}
		ctx.Input.Reset(ctx)
		web.BeeApp.Handlers.FindRouter(ctx)
	}
}

// Beego 路由并发测试
func BenchmarkBeegoRouterConcurrent(b *testing.B) {
	web.BConfig.RunMode = web.PROD
	web.BConfig.WebConfig.AutoRender = false

	for _, route := range testRoutes {
		web.Router(route.path, &testController{}, "get:Get")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := testRequests[i%len(testRequests)]
			httpReq := &http.Request{
				Method: req.method,
				URL:    &url.URL{Path: req.path},
			}
			ctx := &context.Context{
				Request: httpReq,
				Input:   &context.BeegoInput{},
			}
			ctx.Input.Reset(ctx)
			web.BeeApp.Handlers.FindRouter(ctx)
			i++
		}
	})
}

type testController struct {
	web.Controller
}

func (c *testController) Get() {}
