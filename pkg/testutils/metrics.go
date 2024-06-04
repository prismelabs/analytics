package testutils

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

// FindMetric gather and returns metric of the given type, name and with the
// given labels from the registry.
func FindMetric(
	t *testing.T,
	registry *prometheus.Registry,
	metricType io_prometheus_client.MetricType,
	name string,
	labels prometheus.Labels,
) *io_prometheus_client.Metric {
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	for _, metricFamily := range metricFamilies {
		// Check if counter.
		if metricFamily.GetType() != metricType {
			continue
		}

		// Check name.
		if metricFamily.GetName() != name {
			continue
		}

		for _, metric := range metricFamily.Metric {
			labelFound := 0

			for _, label := range metric.Label {
				if labels[label.GetName()] == label.GetValue() {
					labelFound++
				}
			}

			// Some label didn't match.
			if labelFound != len(labels) {
				continue
			}

			return metric
		}
	}

	return nil
}

// CounterValue gather and return value of counter metric with given name and
// labels.
func CounterValue(t *testing.T, registry *prometheus.Registry, name string, labels prometheus.Labels) float64 {
	metric := FindMetric(t, registry, io_prometheus_client.MetricType_COUNTER, name, labels)
	if metric != nil {
		return *metric.Counter.Value
	}

	return 0
}

// HistogramSumValue gather and return sample sum of histogram metric with given
// name and labels.
func HistogramSumValue(t *testing.T, registry *prometheus.Registry, name string, labels prometheus.Labels) float64 {
	metric := FindMetric(t, registry, io_prometheus_client.MetricType_HISTOGRAM, name, labels)
	if metric != nil {
		return *metric.Histogram.SampleSum
	}

	return 0
}

func HistogramBucketValue(t *testing.T, registry *prometheus.Registry, name string, labels prometheus.Labels, upperBound float64) uint64 {
	metric := FindMetric(t, registry, io_prometheus_client.MetricType_HISTOGRAM, name, labels)
	if metric == nil {
		return 0
	}

	for _, bucket := range metric.Histogram.Bucket {
		if bucket.GetUpperBound() == upperBound {
			return bucket.GetCumulativeCount()
		}
	}

	return 0
}

func GaugeValue(t *testing.T, registry *prometheus.Registry, name string, labels prometheus.Labels) float64 {
	metric := FindMetric(t, registry, io_prometheus_client.MetricType_GAUGE, name, labels)
	if metric != nil {
		return *metric.Gauge.Value
	}

	return 0
}
