package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

// Schema GraphQL模式
type Schema struct {
	schema *graphql.Schema
}

// NewSchema 创建新的GraphQL模式
func NewSchema() *Schema {
	return &Schema{}
}

// AddType 添加类型
func (s *Schema) AddType(name string, fields graphql.Fields) {
	// 创建对象类型
	objectType := graphql.NewObject(graphql.ObjectConfig{
		Name:   name,
		Fields: fields,
	})

	// 添加到模式
	if s.schema == nil {
		s.schema = &graphql.Schema{}
	}
}

// AddQuery 添加查询
func (s *Schema) AddQuery(name string, field *graphql.Field) {
	if s.schema == nil {
		s.schema = &graphql.Schema{}
	}

	// 创建查询类型
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: graphql.Fields{name: field},
	})

	// 更新模式
	s.schema, _ = graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
}

// AddMutation 添加变更
func (s *Schema) AddMutation(name string, field *graphql.Field) {
	if s.schema == nil {
		s.schema = &graphql.Schema{}
	}

	// 创建变更类型
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Mutation",
		Fields: graphql.Fields{name: field},
	})

	// 更新模式
	s.schema, _ = graphql.NewSchema(graphql.SchemaConfig{
		Query:    s.schema.QueryType(),
		Mutation: mutationType,
	})
}

// AddSubscription 添加订阅
func (s *Schema) AddSubscription(name string, field *graphql.Field) {
	if s.schema == nil {
		s.schema = &graphql.Schema{}
	}

	// 创建订阅类型
	subscriptionType := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Subscription",
		Fields: graphql.Fields{name: field},
	})

	// 更新模式
	s.schema, _ = graphql.NewSchema(graphql.SchemaConfig{
		Query:        s.schema.QueryType(),
		Mutation:     s.schema.MutationType(),
		Subscription: subscriptionType,
	})
}

// Handler 创建HTTP处理器
func (s *Schema) Handler() http.Handler {
	return handler.New(&handler.Config{
		Schema:   s.schema,
		Pretty:   true,
		GraphiQL: true,
	})
}

// Execute 执行查询
func (s *Schema) Execute(query string, variables map[string]interface{}) *graphql.Result {
	return graphql.Do(graphql.Params{
		Schema:         *s.schema,
		RequestString:  query,
		VariableValues: variables,
	})
}

// Resolver 解析器接口
type Resolver interface {
	Resolve(p graphql.ResolveParams) (interface{}, error)
}

// Field 创建字段
func Field(name string, resolver Resolver, args graphql.FieldConfigArgument) *graphql.Field {
	return &graphql.Field{
		Name: name,
		Type: graphql.String, // 默认类型，应该根据实际情况设置
		Args: args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return resolver.Resolve(p)
		},
	}
}

// Context 上下文
type Context struct {
	context.Context
	Request *http.Request
}

// NewContext 创建新的上下文
func NewContext(ctx context.Context, req *http.Request) *Context {
	return &Context{
		Context: ctx,
		Request: req,
	}
}

// WithContext 添加上下文
func WithContext(ctx context.Context, req *http.Request) *http.Request {
	return req.WithContext(NewContext(ctx, req))
}

// ParseBody 解析请求体
func ParseBody(r *http.Request) (string, map[string]interface{}, error) {
	var body struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return "", nil, fmt.Errorf("failed to parse request body: %v", err)
	}

	return body.Query, body.Variables, nil
}

// Error 错误
type Error struct {
	Message   string   `json:"message"`
	Locations []string `json:"locations,omitempty"`
	Path      []string `json:"path,omitempty"`
}

// NewError 创建新的错误
func NewError(message string) *Error {
	return &Error{
		Message: message,
	}
}

// WithLocations 添加位置
func (e *Error) WithLocations(locations ...string) *Error {
	e.Locations = locations
	return e
}

// WithPath 添加路径
func (e *Error) WithPath(path ...string) *Error {
	e.Path = path
	return e
}

// MarshalJSON 实现json.Marshaler接口
func (e *Error) MarshalJSON() ([]byte, error) {
	type Alias Error
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	})
}

// Response 响应
type Response struct {
	Data   interface{} `json:"data,omitempty"`
	Errors []*Error    `json:"errors,omitempty"`
}

// NewResponse 创建新的响应
func NewResponse(data interface{}) *Response {
	return &Response{
		Data: data,
	}
}

// WithErrors 添加错误
func (r *Response) WithErrors(errors ...*Error) *Response {
	r.Errors = errors
	return r
}

// MarshalJSON 实现json.Marshaler接口
func (r *Response) MarshalJSON() ([]byte, error) {
	type Alias Response
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}
