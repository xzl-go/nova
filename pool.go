package core

import (
	"net/http"
	"sync"
	"time"
)

// 对象池
var (
	contextPool = sync.Pool{
		New: func() interface{} {
			return &Context{
				Params: make(map[string]string),
				Errors: make([]error, 0),
			}
		},
	}

	responseWriterPool = sync.Pool{
		New: func() interface{} {
			return &ResponseWriter{}
		},
	}
)

// GetContext 从对象池获取 Context
func GetContext(w http.ResponseWriter, r *http.Request) *Context {
	c := contextPool.Get().(*Context)
	c.Request = r
	c.Response = w
	c.Writer = GetResponseWriter(w)
	c.start = time.Now()
	c.Index = -1
	return c
}

// PutContext 将 Context 放回对象池
func PutContext(c *Context) {
	c.reset()
	PutResponseWriter(c.Writer)
	contextPool.Put(c)
}

// GetResponseWriter 从对象池获取 ResponseWriter
func GetResponseWriter(w http.ResponseWriter) *ResponseWriter {
	rw := responseWriterPool.Get().(*ResponseWriter)
	rw.ResponseWriter = w
	rw.Status = http.StatusOK
	return rw
}

// PutResponseWriter 将 ResponseWriter 放回对象池
func PutResponseWriter(rw *ResponseWriter) {
	rw.ResponseWriter = nil
	rw.Status = 0
	responseWriterPool.Put(rw)
}
