package core

import (
	"fmt"
	"testing"

	core "github.com/xzl-go/nova"
)

// 基准测试数据
var (
	// 不同路径长度的路由
	shortPath  = "/api/v1/users"
	mediumPath = "/api/v1/users/profile/settings/notifications"
	longPath   = "/api/v1/users/profile/settings/notifications/email/marketing/preferences"

	// 不同参数数量的路由
	noParamPath     = "/api/v1/users"
	singleParamPath = "/api/v1/users/:id"
	multiParamPath  = "/api/v1/users/:id/posts/:postId/comments/:commentId"

	// 通配符路由
	wildcardPath = "/api/v1/*path"

	// 混合路由
	mixedPaths = []string{
		"/api/v1/users",
		"/api/v1/users/:id",
		"/api/v1/users/:id/posts",
		"/api/v1/users/:id/posts/:postId",
		"/api/v1/users/:id/posts/:postId/comments",
		"/api/v1/users/:id/posts/:postId/comments/:commentId",
		"/api/v1/*path",
	}
)

// 测试不同路径长度的性能
func BenchmarkRouterPathLength(b *testing.B) {
	router := NewRouter()
	router.AddRoute("GET", shortPath, func(c *core.Context) {})
	router.AddRoute("GET", mediumPath, func(c *core.Context) {})
	router.AddRoute("GET", longPath, func(c *core.Context) {})

	b.Run("ShortPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			router.FindRoute("GET", shortPath)
		}
	})

	b.Run("MediumPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			router.FindRoute("GET", mediumPath)
		}
	})

	b.Run("LongPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			router.FindRoute("GET", longPath)
		}
	})
}

// 测试不同参数数量的性能
func BenchmarkRouterParams(b *testing.B) {
	router := NewRouter()
	router.AddRoute("GET", noParamPath, func(c *core.Context) {})
	router.AddRoute("GET", singleParamPath, func(c *core.Context) {})
	router.AddRoute("GET", multiParamPath, func(c *core.Context) {})

	b.Run("NoParams", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			router.FindRoute("GET", "/api/v1/users")
		}
	})

	b.Run("SingleParam", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			router.FindRoute("GET", "/api/v1/users/123")
		}
	})

	b.Run("MultiParams", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			router.FindRoute("GET", "/api/v1/users/123/posts/456/comments/789")
		}
	})
}

// 测试通配符路由性能
func BenchmarkRouterWildcard(b *testing.B) {
	router := NewRouter()
	router.AddRoute("GET", wildcardPath, func(c *core.Context) {})

	b.Run("Wildcard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			router.FindRoute("GET", "/api/v1/any/path/here")
		}
	})
}

// 测试混合路由场景
func BenchmarkRouterMixed(b *testing.B) {
	router := NewRouter()
	for _, path := range mixedPaths {
		router.AddRoute("GET", path, func(c *core.Context) {})
	}

	b.Run("Mixed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			router.FindRoute("GET", "/api/v1/users/123/posts/456/comments/789")
		}
	})
}

// 测试路由注册性能
func BenchmarkRouterRegistration(b *testing.B) {
	router := NewRouter()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/api/v1/users/%d", i)
		router.AddRoute("GET", path, func(c *core.Context) {})
	}
}

// 测试内存使用
func BenchmarkRouterMemoryUsage(b *testing.B) {
	router := NewRouter()

	// 预注册一些路由
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("/api/v1/users/%d", i)
		router.AddRoute("GET", path, func(c *core.Context) {})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/api/v1/users/%d", i)
		router.FindRoute("GET", path)
	}
}

// 测试并发性能
func BenchmarkRouterConcurrent(b *testing.B) {
	router := NewRouter()

	// 预注册一些路由
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("/api/v1/users/%d", i)
		router.AddRoute("GET", path, func(c *core.Context) {})
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			path := fmt.Sprintf("/api/v1/users/%d", i%1000)
			router.FindRoute("GET", path)
			i++
		}
	})
}

// 测试热点路由性能
func BenchmarkRouterHotPath(b *testing.B) {
	router := NewRouter()
	hotPath := "/api/v1/users/123"
	router.AddRoute("GET", hotPath, func(c *core.Context) {})

	// 预热
	for i := 0; i < 1000; i++ {
		router.FindRoute("GET", hotPath)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.FindRoute("GET", hotPath)
	}
}
