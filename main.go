package main

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "go.opentelemetry.io/contrib/bridges/otelslog"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
    "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
    "go.opentelemetry.io/otel/log/global"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/log"
    "go.opentelemetry.io/otel/sdk/trace"
)

type contextKey string

const (
    loggerKey contextKey = "logger"
)

func main() {
    if err := run(); err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}

func run() error {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()
    
    shutdown, err := setupTracing()
    if err != nil {
        return err
    }
    defer func() {
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        shutdown(shutdownCtx)
    }()
    
    baseLogger := otelslog.NewLogger("app")
    
    r := chi.NewRouter()
    r.Use(RequestLoggerMiddleware(baseLogger))
    r.Get("/", homeHandler)
    
    srv := &http.Server{
        Addr:    ":3000",
        Handler: r,
    }
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Printf("Server error: %v\n", err)
        }
    }()
    
    <-ctx.Done()
    
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return srv.Shutdown(shutdownCtx)
}

func setupTracing() (func(context.Context) error, error) {
    // Trace exporter
    traceExporter, err := stdouttrace.New()
    if err != nil {
        return nil, err
    }

    // Log exporter
    logExporter, err := stdoutlog.New()
    if err != nil {
        return nil, err
    }

    // Trace provider
    tracerProvider := trace.NewTracerProvider(
        trace.WithBatcher(traceExporter),
    )

    // Log provider
    loggerProvider := log.NewLoggerProvider(
        log.WithProcessor(log.NewBatchProcessor(logExporter)),
    )

    otel.SetTracerProvider(tracerProvider)
    global.SetLoggerProvider(loggerProvider)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    // Combined shutdown
    shutdown := func(ctx context.Context) error {
        if err := tracerProvider.Shutdown(ctx); err != nil {
            return err
        }
        return loggerProvider.Shutdown(ctx)
    }

    return shutdown, nil
}

func RequestLoggerMiddleware(baseLogger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := r.Context()
            
            tracer := otel.Tracer("app")
            ctx, span := tracer.Start(ctx, r.URL.Path)
            defer span.End()
            
            requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
            traceID := span.SpanContext().TraceID().String()
            spanID := span.SpanContext().SpanID().String()
            
            logger := baseLogger.With(
                "request_id", requestID,
                "trace_id", traceID,
                "span_id", spanID,
                "method", r.Method,
                "path", r.URL.Path,
            )
            
            ctx = context.WithValue(ctx, loggerKey, logger)
            
            logger.InfoContext(ctx, "request started")
            next.ServeHTTP(w, r.WithContext(ctx))
            logger.InfoContext(ctx, "request finished")
        })
    }
}

func GetLogger(ctx context.Context) *slog.Logger {
    if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return logger
    }
    return slog.Default()
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger := GetLogger(ctx)
    logger.InfoContext(ctx, "processing home request")
    w.Write([]byte("Hello World"))
}