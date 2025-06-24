# Nova 企业级高性能 Web 框架

<div align="center">
    <img src="https://img.shields.io/github/stars/xzl/nova?style=social" alt="Stars">
    <img src="https://img.shields.io/github/license/xzl/nova" alt="License">
    <img src="https://img.shields.io/github/go-mod/go-version/xzl/nova" alt="Go Version">
    <img src="https://img.shields.io/github/actions/workflow/status/xzl/nova/ci.yml?branch=main" alt="Build Status">
</div>

## 🌟 项目背景

在现代企业级后端开发中，开发者需要一个高性能、易扩展且安全的框架来支撑复杂应用。Nova 诞生于此，旨在简化后端开发流程，提升系统稳定性与开发效率。

## 🚀 项目愿景

我们的目标是打造一个：
- 🔒 **安全可靠** 的企业级后端开发框架
- 🚀 **高性能** 的企业级解决方案
- 🛠️ **开发友好** 的工具集

## ✨ 核心特性

### 🔒 安全性
- JWT 认证与授权
- 细粒度权限控制
- 动态令牌管理
- 防御常见 Web 漏洞

### 🛠️ 配置管理
- 多环境配置支持（YAML/JSON/ENV）
- 环境变量优先级覆盖
- 配置热重载
- 配置校验与安全检查
- 支持分布式配置中心（如 etcd）

### 📊 可观测性
- 性能指标采集
- 结构化日志
- 分布式追踪
- 智能错误处理

### 🚀 高性能
- 低内存占用
- 协程安全
- 动态资源池
- 高效 HTTP 路由

### 🔧 开发友好
- 直观 API 设计
- 灵活中间件机制
- 代码即配置
- 丰富开发工具
- 插件化扩展能力

### 🌐 服务注册与发现
- 支持 Consul、etcd 等主流服务注册与发现
- 适合多实例部署、服务动态扩缩容

## 🛠 系统要求

### 运行环境
- Go 1.24.3+
- 推荐操作系统：
  - macOS 12.0+
  - Linux (Ubuntu 20.04+, CentOS 8+)
  - Windows 10/11

### 硬件建议
- CPU: 2 核心及以上
- 内存: 4GB RAM
- 磁盘: 10GB 可用空间

## 📦 安装与部署

### 方法 1：Go 包管理器安装

```bash
go get -u github.com/xzl-go/nova
# 或安装指定版本
go get github.com/xzl-go/nova@v1.0.0
```

### 方法 2：源码安装

```bash
git clone https://github.com/xzl-go/nova.git
cd nova
go mod tidy
make build
make test
```

## 🚀 快速开始

### 1. 创建配置文件 `config.yaml`

```yaml
server:
  host: 0.0.0.0
  port: 8080

log:
  level: info
  file: app.log

security:
  jwt_secret: your_secret_key
  token_expiry: 24h

performance:
  max_cpu_percent: 0.8
```

### 2. 编写主程序 `main.go`

```go
package main

import (
    "fmt"
    "net/http"
    "nova/pkg/config"
    "nova/pkg/logger"
    "nova/pkg/middleware"
    "nova/core"
)

func main() {
    // 加载配置
    cfg := config.New(
        config.WithConfigFile("config.yaml"),
    )
    if err := cfg.Load(); err != nil {
        panic(err)
    }

    // 创建日志记录器
    log := logger.NewLogger(
        logger.LogLevel(cfg.GetInt("log.level")),
        cfg.GetString("log.file"),
    )
    defer log.Close()

    // JWT认证中间件
    jwtMiddleware := middleware.NewJWTMiddleware(
        cfg.GetString("security.jwt_secret"),
        log,
    )

    router := core.NewRouter()

    // 公开路由
    router.GET("/", func(w http.ResponseWriter, r *http.Request) {
        ctx := core.NewContext(w, r)
        log.Info("访问根路径")
        ctx.JSON(200, map[string]string{
            "message": "欢迎使用Nova企业级高性能 Web 框架",
        })
    })

    // 受保护的路由
    router.GET("/admin",
        jwtMiddleware.Authorize(
            middleware.PermAdmin,
        )(func(w http.ResponseWriter, r *http.Request) {
            ctx := core.NewContext(w, r)
            ctx.JSON(200, map[string]string{
                "message": "管理员页面",
            })
        }),
    )

    // 启动方式一：推荐，类似 Gin 的 router.Run()
    router.Run(":8080")

    // 启动方式二：自定义 Server 配置
    // server := core.NewServer(
    //     core.WithAddr(fmt.Sprintf("%s:%d",
    //         cfg.GetString("server.host"),
    //         cfg.GetInt("server.port"),
    //     )),
    //     core.WithHandler(router),
    // )
    // server.Start()
}
```

> Nova 支持类似 Gin 的 `router.Run(":8080")` 方式一键启动服务，也支持自定义 Server 配置的高级用法。你可以根据实际需求选择任意一种方式。

### 3. 运行应用

```bash
go run main.go
```

## 📈 性能基准测试

| 指标         | 数据       | 说明                     |
|--------------|------------|--------------------------|
| 平均响应时间 | < 10ms     | 99% 请求在 10ms 内响应   |
| 内存占用     | < 50MB     | 空闲状态下内存使用      |
| 并发支持     | 1000 req/s | 单实例并发处理能力       |

> 我们使用 `wrk` 和 `hey` 进行性能压测，确保框架的高性能表现。

## 🛡️ 安全最佳实践

- 使用强密码和随机JWT密钥
- 限制权限范围
- 定期轮换令牌
- 启用HTTPS
- 实施速率限制

## 🤝 贡献指南

1. Fork 仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交代码变更 (`git commit -m '添加了某个特性'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

感谢所有为项目做出贡献的开发者！

## 📞 社区支持

- 📧 邮件：17633997766@163.com
- 💬 讨论区：[GitHub Discussions](https://github.com/xzl-go/nova/discussions)
- 🐞 问题反馈：[GitHub Issues](https://github.com/xzl-go/nova/issues)

## 📄 许可证

本项目采用 MIT 许可证，详情请参见 [LICENSE](LICENSE) 文件。

## 🌟 赞助与支持

如果你喜欢这个项目：
- 给我们一个 Star ⭐️
- 考虑赞助我们的开源工作

<div align="center">
    <a href="https://github.com/sponsors/xzl-go">
        <img src="https://img.shields.io/badge/sponsor-❤-ff69b4" alt="Sponsor">
    </a>
</div>

---

**💡 提示**：查看我们的 [文档网站](https://nova.com) 获取更多详细信息！ 
