package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		propagator := otel.GetTextMapPropagator()
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		spanName := c.Request.Method + " " + c.Request.URL.Path
		ctx, span := otel.Tracer("signaling-server").Start(ctx, spanName)
		defer span.End()

		// Add trace context to request
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Add trace attributes
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.Int64("http.status_code", int64(c.Writer.Status())),
		)
	}
}
