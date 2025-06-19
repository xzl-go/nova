package template

import (
	"bytes"
	"html/template"
	"io"
	"path/filepath"
	"sync"

	"github.com/xzl/nova/core"
)

// Engine 模板引擎
type Engine struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
	mu        sync.RWMutex
}

// New 创建新的模板引擎
func New() *Engine {
	return &Engine{
		templates: make(map[string]*template.Template),
		funcMap:   make(template.FuncMap),
	}
}

// AddFunc 添加模板函数
func (e *Engine) AddFunc(name string, fn interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.funcMap[name] = fn
}

// Load 加载模板
func (e *Engine) Load(pattern string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		name := filepath.Base(file)
		tmpl := template.New(name).Funcs(e.funcMap)
		if _, err := tmpl.ParseFiles(file); err != nil {
			return err
		}
		e.templates[name] = tmpl
	}

	return nil
}

// Render 渲染模板
func (e *Engine) Render(c *core.Context, name string, data interface{}) error {
	e.mu.RLock()
	tmpl, ok := e.templates[name]
	e.mu.RUnlock()

	if !ok {
		return core.NewError(500, "template not found: "+name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	_, err := io.Copy(c.Writer, &buf)
	return err
}

// Template 模板中间件
func Template(e *Engine) core.HandlerFunc {
	return func(c *core.Context) {
		c.Set("template", e)
		c.Next()
	}
}
