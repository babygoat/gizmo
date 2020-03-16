package observe

import (
	"os"

	datadog "github.com/DataDog/opencensus-go-exporter-datadog"
	"github.com/pkg/errors"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

// RegisterAndObserveDatadog will initiate and register Datadog metrics and tracing in environments
// that pass the tests in IsDatadogEnabled function
func RegisterAndObserveDatadog(onError func(error)) error {
	if SkipObserve() {
		return nil
	}

	exp, err := NewDatadogExporter(onError)
	if err != nil {
		return errors.Wrap(err, "unable to initiate Datadog tracing exporter")
	}
	trace.RegisterExporter(exp)
	view.RegisterExporter(exp)

	return nil
}

// NewDatadogExporter will return the tracing and metrics through
// the stack driver exporter, if exists in the underlying platform.
// If exporter is registered, it returns the exporter so you can register
// it and ensure to call Flush on termination.
func NewDatadogExporter(onErr func(error)) (*datadog.Exporter, error) {
	_, service, version := GetServiceInfo()

	opts := getDatadogOpts(service, version, getDatadogAddr(), onErr)
	if opts == nil {
		return nil, nil
	}

	return datadog.NewExporter(*opts)
}

func getDatadogAddr() string {
	return os.Getenv("DATADOG_ADDR")
}

// getDatadogOpts returns Datadog Options that you can pass directly
// to the OpenCensus exporter or other libraries.
func getDatadogOpts(service, version, datadogAddress string, onErr func(err error)) *datadog.Options {

	canExport := IsDatadogEnabled()
	if !canExport {
		return nil
	}

	return &datadog.Options{
		// Namespace specifies the namespaces to which metric keys are appended.
		// TODO: Figure out what the namespace should be. Can be either a projectID or something else.
		Namespace: "opencensus",

		// Service specifies the service name used for tracing.
		Service: service,

		// TraceAddr specifies the host[:port] address of the Datadog Trace Agent.
		// It defaults to localhost:8126.
		TraceAddr: datadogAddress,

		// StatsAddr specifies the host[:port] address for DogStatsD. It defaults
		// to localhost:8125.
		StatsAddr: datadogAddress,

		// OnError specifies a function that will be called if an error occurs during
		// processing stats or metrics.
		OnError: onErr,

		// // Tags specifies a set of global tags to attach to each metric.
		// Tags []string

		// GlobalTags holds a set of tags that will automatically be applied to all
		// exported spans.
		GlobalTags: map[string]interface{}{
			"service": service,
			"version": version,
		},
		// // DisableCountPerBuckets specifies whether to emit count_per_bucket metrics
		// DisableCountPerBuckets bool
	}
}
