package trace

import (
	"context"
	"fmt"
	"github.com/xzl-go/nova"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Config 追踪配置
type Config struct {
	ServiceName string
	Endpoint    string // Jaeger/Zipkin 地址
	Env         string // 环境
}

var tp *sdktrace.TracerProvider

// Init 初始化全局 TracerProvider
func Init(cfg *Config) error {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.Endpoint)))
	if err != nil {
		return err
	}

	tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironment(cfg.Env),
		)),
	)
	otel.SetTracerProvider(tp)
	return nil
}

// Shutdown 关闭 TracerProvider
func Shutdown(ctx context.Context) error {
	if tp != nil {
		return tp.Shutdown(ctx)
	}
	return nil
}

// Tracing 追踪中间件
func Tracing(service string) nova.HandlerFunc {
	tracer := otel.Tracer(service)
	return func(c *nova.Context) {
		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+c.Request.URL.Path)
		defer span.End()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// StartSpan 手动创建业务埋点
func StartSpan(ctx context.Context, name string) (context.Context, oteltrace.Span) {
	return otel.Tracer("").Start(ctx, name)
}

// Tracer 追踪器
type Tracer struct {
	tracer oteltrace.Tracer
}

// NewTracer 创建追踪器
func NewTracer(cfg *Config) (*Tracer, error) {
	// 创建Jaeger导出器
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.Endpoint)))
	if err != nil {
		return nil, fmt.Errorf("failed to create jaeger exporter: %v", err)
	}

	// 创建资源
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.DeploymentEnvironment(cfg.Env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %v", err)
	}

	// 创建追踪提供者
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	// 设置全局追踪提供者
	otel.SetTracerProvider(tp)

	return &Tracer{
		tracer: tp.Tracer(cfg.ServiceName),
	}, nil
}

// StartSpan 开始新的span
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// EndSpan 结束span
func (t *Tracer) EndSpan(span oteltrace.Span) {
	span.End()
}

// AddEvent 添加事件
func (t *Tracer) AddEvent(span oteltrace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, oteltrace.WithAttributes(attrs...))
}

// SetAttributes 设置属性
func (t *Tracer) SetAttributes(span oteltrace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// SpanFromContext 从上下文获取span
func (t *Tracer) SpanFromContext(ctx context.Context) oteltrace.Span {
	return oteltrace.SpanFromContext(ctx)
}

// TraceFunc 追踪函数执行
func (t *Tracer) TraceFunc(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, span := t.StartSpan(ctx, name)
	defer t.EndSpan(span)

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	span.SetAttributes(
		attribute.Int64("duration_ms", duration.Milliseconds()),
	)

	return err
}
