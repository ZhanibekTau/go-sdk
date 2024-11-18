package plugin

import (
	"github.com/ZhanibekTau/go-sdk/pkg/tracer"
	span2 "github.com/ZhanibekTau/go-sdk/pkg/tracer/span"
	trace2 "go.opencensus.io/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

func GormPluginWithTrace() gorm.Plugin {
	return &PluginTrace{}
}

type PluginTrace struct {
}

func (p *PluginTrace) Name() string {
	return "plugin_trace"
}

func (p *PluginTrace) Initialize(db *gorm.DB) error {
	tracer := tracer.TraceClient

	if tracer == nil || !tracer.IsEnabled {
		return nil
	}

	// Before any operation callback
	db.Callback().Create().Before("gorm:before_create").Register("gormotel:before_create", p.before(tracer))
	db.Callback().Query().Before("gorm:before_query").Register("gormotel:before_query", p.before(tracer))
	db.Callback().Delete().Before("gorm:before_delete").Register("gormotel:before_delete", p.before(tracer))
	db.Callback().Update().Before("gorm:before_update").Register("gormotel:before_update", p.before(tracer))
	db.Callback().Row().Before("gorm:before_row").Register("gormotel:before_row", p.before(tracer))
	db.Callback().Raw().Before("gorm:before_raw").Register("gormotel:before_raw", p.before(tracer))

	// After any operation callback
	db.Callback().Create().After("gorm:after_create").Register("gormotel:after_create", p.after)
	db.Callback().Query().After("gorm:after_query").Register("gormotel:after_query", p.after)
	db.Callback().Delete().After("gorm:after_delete").Register("gormotel:after_delete", p.after)
	db.Callback().Update().After("gorm:after_update").Register("gormotel:after_update", p.after)
	db.Callback().Row().After("gorm:after_row").Register("gormotel:after_row", p.after)
	db.Callback().Raw().After("gorm:after_raw").Register("gormotel:after_raw", p.after)

	return nil
}

func (p *PluginTrace) before(tracer *tracer.Tracer) func(*gorm.DB) {
	return func(db *gorm.DB) {
		ctx, span := tracer.CreateSpan(db.Statement.Context, "[DB]")
		db.InstanceSet("otel:span", span)
		db.Statement.Context = ctx
	}
}

func (p *PluginTrace) after(db *gorm.DB) {
	if spanVal, ok := db.InstanceGet("otel:span"); ok {
		if span, ok := spanVal.(trace.Span); ok {
			defer span.End()

			span.SetAttributes(
				attribute.String(span2.AttributeDBStatement, db.Statement.SQL.String()),
				attribute.String(span2.AttributeDBTable, db.Statement.Table),
				attribute.Int64(span2.AttributeDbRowsAffected, db.RowsAffected),
			)

			if db.Error != nil {
				span.RecordError(db.Error)
				span.SetStatus(trace2.StatusCodeOK, db.Error.Error())
			}
		}
	}
}
