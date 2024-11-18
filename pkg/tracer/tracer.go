package tracer

import (
	"bytes"
	"context"
	"crypto/tls"
	"github.com/ZhanibekTau/go-sdk/pkg/config"
	"github.com/ZhanibekTau/go-sdk/pkg/exception"
	span2 "github.com/ZhanibekTau/go-sdk/pkg/tracer/span"
	"github.com/ZhanibekTau/go-sdk/pkg/tracer/structure"
	"github.com/gin-gonic/gin"
	"github.com/gookit/goutil/netutil/httpctype"
	"github.com/gookit/goutil/netutil/httpheader"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	trace2 "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"io"
	"net/http"
	"strings"
)

// TraceClient - данная глобальная переменная нужна для пакета http-билдера в основном.
var TraceClient *Tracer

type Tracer struct {
	tp          *tracesdk.TracerProvider
	cfg         *structure.TraceConfig
	IsEnabled   bool
	ServiceName string
}

// InitTraceClient - создание клиента трассировки
func InitTraceClient() (*Tracer, error) {
	t := &Tracer{}
	// config init
	if err := t.initTraceConfig(); err != nil {
		return nil, err
	}

	if !t.cfg.IsTraceEnabled {
		return t, nil
	}

	// отключил провеку сертификата, так как на тесте были ошибки "x509: certificate signed by unknown authority error"
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// Create the Jaeger exporter
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(t.cfg.Url),
			jaeger.WithHTTPClient(&http.Client{Transport: transport}),
		),
	)

	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(t.cfg.ServiceName),
			attribute.String("environment", "development"),
			attribute.Int64("ID", 1),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(b3.New()))

	t.tp = tp
	TraceClient = t

	return t, nil
}

// Shutdown -
func (t *Tracer) Shutdown(ctx context.Context) error {
	return t.tp.Shutdown(ctx)
}

// InjectHttpTraceId -  записывает  trace id  в запрос, требует  *http.Request
func (t *Tracer) InjectHttpTraceId(ctx context.Context, req *http.Request) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
}

// MiddleWareExtractTraceId -  мидлвар который записывает трассировку
func (t *Tracer) MiddleWareExtractTraceId() gin.HandlerFunc {
	return func(c *gin.Context) {
		if t == nil || !t.cfg.IsTraceEnabled {
			c.Next()

			return
		}

		parentCtx, span := t.CreateSpan(c.Request.Context(), "["+c.Request.Method+"] "+c.FullPath())
		defer span.End()

		// парсинг body
		if t.cfg.IsHttpBodyEnabled {
			// нет смысла копировать тело запроса при наличии файла
			if !strings.HasPrefix(c.GetHeader(httpheader.ContentType), httpctype.MIMEDataForm) {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				span.SetAttributes(attribute.String(span2.AttributeReqBody, string(bodyBytes)))
			}
		}

		c.Request = c.Request.WithContext(parentCtx)
		c.Next()

		// парсинг ошибок
		{
			excep := c.Keys["exception"]

			switch v := excep.(type) {
			case *exception.AppException:
				span.SetAttributes(attribute.Int(span2.AttributeRespHttpCode, v.Code))
				if v.Error != nil {
					span.SetAttributes(attribute.String(span2.AttributeRespErrMsg, v.Error.Error()))
				}
			default:
				span.SetAttributes(attribute.Int(span2.AttributeRespHttpCode, c.Writer.Status()))
			}
		}
	}
}

// CreateSpan - Создает родительский спан,и возвращает контекст, этот контекст нужен для дочернего спана.
// В случае если в ctx нет контекста родителя то создается контекст родителя
// Не забыть вызывать span.End()
func (t *Tracer) CreateSpan(ctx context.Context, name string) (context.Context, trace2.Span) {
	if t == nil || t.tp == nil {
		return context.Background(), noop.Span{}
	}

	return t.tp.Tracer(t.ServiceName).Start(ctx, name)
}

// CreateSpanWithCustomTraceId -  экспериментальный метод, создаем спан на основе кастомного трайс айди
func (t *Tracer) CreateSpanWithCustomTraceId(ctx context.Context, traceId, name string) (context.Context, trace2.Span, error) {
	tId, err := trace2.TraceIDFromHex(traceId)

	if err != nil {
		return nil, noop.Span{}, err
	}

	spanContext := trace2.NewSpanContext(trace2.SpanContextConfig{
		TraceID: tId,
	})

	ctx1 := trace2.ContextWithSpanContext(ctx, spanContext)
	ctx1, span := t.tp.Tracer(t.ServiceName).Start(ctx1, name)

	return ctx1, span, nil
}

// initTraceConfig -  инициализирует конфиг трассировки, читает  из файла  .env переменки
func (t *Tracer) initTraceConfig() error {
	if err := config.ReadEnv(); err != nil {
		return err
	}

	traceCfg := &structure.TraceConfig{}
	err := config.InitConfig(traceCfg)

	if err != nil {
		return err
	}

	t.cfg = traceCfg
	t.ServiceName = traceCfg.ServiceName
	t.IsEnabled = traceCfg.IsTraceEnabled

	return nil
}
