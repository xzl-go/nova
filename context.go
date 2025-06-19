package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/xzl/nova/pkg/binding"
)

// Context 请求上下文
type Context struct {
	Request    *http.Request
	Response   http.ResponseWriter
	Params     map[string]string
	Data       interface{}
	start      time.Time
	Index      int
	aborted    bool
	engine     *Engine
	Writer     *ResponseWriter
	handlers   []HandlerFunc
	StatusCode int
	Errors     []error
	store      map[string]interface{}
	storeMutex sync.RWMutex
}

// NewContext 创建新的上下文
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  r,
		Response: w,
		Params:   make(map[string]string),
		start:    time.Now(),
		Index:    -1,
		aborted:  false,
	}
}

// ClientIP 获取客户端IP
func (c *Context) ClientIP() string {
	// 按优先级获取IP
	ip := c.Request.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	ip = c.Request.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}
	return c.Request.RemoteAddr
}

// Header 设置响应头
func (c *Context) Header(key, value string) {
	c.Response.Header().Set(key, value)
}

// Status 获取响应状态码
func (c *Context) Status() int {
	return c.Writer.Status
}

// JSON 返回JSON响应
func (c *Context) JSON(code int, data interface{}) {
	c.Header("Content-Type", "application/json")
	c.Response.WriteHeader(code)
	json.NewEncoder(c.Response).Encode(data)
}

// ErrorResponse 返回错误响应
func (c *Context) ErrorResponse(code int, message string) {
	c.Writer.WriteHeader(code)
	c.Writer.Write([]byte(message))
}

// SuccessResponse 返回成功响应
func (c *Context) SuccessResponse(data interface{}) {
	c.JSON(http.StatusOK, data)
}

// Fail 返回失败响应
func (c *Context) Fail(code int, msg string) {
	c.ErrorResponse(code, msg)
}

// String 返回字符串响应
func (c *Context) String(code int, format string, values ...interface{}) {
	c.Header("Content-Type", "text/plain")
	c.Response.WriteHeader(code)
	fmt.Fprintf(c.Response, format, values...)
}

// Next 执行下一个中间件
func (c *Context) Next() {
	c.Index++
	for c.Index < len(c.handlers) {
		c.handlers[c.Index](c)
		c.Index++
	}
}

// Abort 中断中间件链
func (c *Context) Abort() {
	c.Index = len(c.handlers)
}

// IsAborted 检查是否已中断
func (c *Context) IsAborted() bool {
	return c.Index >= len(c.handlers)
}

// Error 添加错误
func (c *Context) Error(err error) {
	c.Errors = append(c.Errors, err)
}

// HasError 检查是否有错误
func (c *Context) HasError() bool {
	return len(c.Errors) > 0
}

// GetError 获取所有错误
func (c *Context) GetError() error {
	if len(c.Errors) == 0 {
		return nil
	}
	return c.Errors[len(c.Errors)-1]
}

// ResponseWriter 自定义响应写入器
type ResponseWriter struct {
	http.ResponseWriter
	Status int
}

// WriteHeader 重写 WriteHeader 方法
func (w *ResponseWriter) WriteHeader(code int) {
	w.Status = code
	w.ResponseWriter.WriteHeader(code)
}

// Status 获取状态码
func (w *ResponseWriter) GetStatus() int {
	return w.Status
}

// reset 重置上下文状态
func (c *Context) reset() {
	c.Params = make(map[string]string)
	c.Index = -1
	c.aborted = false
}

// GetParam 获取路由参数
func (c *Context) GetParam(key string) string {
	return c.Params[key]
}

// SetParam 设置路由参数
func (c *Context) SetParam(key, value string) {
	c.Params[key] = value
}

// ShouldBind 绑定请求参数
func (c *Context) ShouldBind(obj interface{}) error {
	b := c.getBinding()
	return b.Bind(c.Request, obj)
}

// ShouldBindJSON 绑定 JSON 参数
func (c *Context) ShouldBindJSON(obj interface{}) error {
	return binding.JSON.Bind(c.Request, obj)
}

// ShouldBindXML 绑定 XML 参数
func (c *Context) ShouldBindXML(obj interface{}) error {
	return binding.XML.Bind(c.Request, obj)
}

// ShouldBindQuery 绑定 Query 参数
func (c *Context) ShouldBindQuery(obj interface{}) error {
	return binding.Query.Bind(c.Request, obj)
}

// ShouldBindForm 绑定 Form 参数
func (c *Context) ShouldBindForm(obj interface{}) error {
	return binding.Form.Bind(c.Request, obj)
}

// getBinding 获取绑定器
func (c *Context) getBinding() binding.Binding {
	contentType := c.Request.Header.Get("Content-Type")
	switch {
	case strings.Contains(contentType, "application/json"):
		return binding.JSON
	case strings.Contains(contentType, "application/xml"):
		return binding.XML
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		return binding.FormPost
	case strings.Contains(contentType, "multipart/form-data"):
		return binding.FormMultipart
	default:
		return binding.Form
	}
}

// Set 设置值
func (c *Context) Set(key string, value interface{}) {
	c.storeMutex.Lock()
	defer c.storeMutex.Unlock()
	if c.store == nil {
		c.store = make(map[string]interface{})
	}
	c.store[key] = value
}

// Get 获取值
func (c *Context) Get(key string) (interface{}, bool) {
	c.storeMutex.RLock()
	defer c.storeMutex.RUnlock()
	value, exists := c.store[key]
	return value, exists
}
