package testutils

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func CounterValue(t *testing.T, registry *prometheus.Registry, name string, labels prometheus.Labels) float64 {
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	for _, metricFamily := range metricFamilies {
		// Check if counter.
		if metricFamily.GetType() != io_prometheus_client.MetricType_COUNTER {
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
				t.Log("label found doesn't match, found", labelFound, "expected, ", len(labels))
				continue
			}

			return *metric.Counter.Value
		}
	}

	t.Log("counter metric not found")
	return 0
}
