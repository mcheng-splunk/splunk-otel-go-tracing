package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/signalfx/splunk-otel-go/distro"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// newExporter returns a console exporter.
func newExporter(w io.Writer) (trace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

// newResource returns a resource describing this application.
func newResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("fib"),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)
	return r
}

func init() {

	// Configured to send our trace to Otel Agent residing in our network.
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=go-Fibonacci,service.version=1.0.0,deployment.environment=development")
	os.Setenv("OTEL_LOG_LEVEL", "debug")
	os.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", "http://192.168.20.34:14268/api/traces")
	os.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk")
}

func main() {
	l := log.New(os.Stdout, "", 0)

	// *************************************
	// Splunk Otel
	sdk, err := distro.Run()
	if err != nil {
		panic(err)
	}

	// Ensure all spans are flushed before the application exits.
	defer func() {
		if err := sdk.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}()

	OTEL_LOG_LEVEL := os.Getenv("OTEL_LOG_LEVEL")
	OTEL_EXPORTER_OTLP_TRACES_ENDPOINT := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	fmt.Printf("Display environment variables OTEL_LOG_LEVEL: %s, OTEL_EXPORTER_OTLP_TRACES_ENDPOINT: %s\n", OTEL_LOG_LEVEL, OTEL_EXPORTER_OTLP_TRACES_ENDPOINT)

	// *************************************

	// // Write telemetry data to a file.
	// f, err := os.Create("traces.txt")
	// if err != nil {
	// 	l.Fatal(err)
	// }
	// defer f.Close()

	// exp, err := newExporter(f)
	// if err != nil {
	// 	l.Fatal(err)
	// }

	// tp := trace.NewTracerProvider(
	// 	trace.WithBatcher(exp),
	// 	trace.WithResource(newResource()),
	// )
	// defer func() {
	// 	if err := tp.Shutdown(context.Background()); err != nil {
	// 		l.Fatal(err)
	// 	}
	// }()
	// otel.SetTracerProvider(tp)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	errCh := make(chan error)
	app := NewApp(os.Stdin, l)
	go func() {
		errCh <- app.Run(context.Background())
	}()

	select {
	case <-sigCh:
		l.Println("\ngoodbye")
		return
	case err := <-errCh:
		if err != nil {
			l.Fatal(err)
		}
	}
}
