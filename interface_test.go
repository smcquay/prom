package prom

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestGauginess(t *testing.T) {
	g := NewGauge(prometheus.GaugeOpts{
		Name: "test",
		Help: "test help",
	}, 3)

	switch g.(type) {
	case prometheus.Gauge:
	default:
		t.Fatalf("FixedPrecision is not a prometheus.Gauge")
	}
}

func TestCounteriness(t *testing.T) {
	c := NewCounter(prometheus.CounterOpts{
		Name: "test",
		Help: "test help",
	}, 3)

	switch c.(type) {
	case prometheus.Counter:
	default:
		t.Fatalf("FixedPrecision is not a prometheus.Counter")
	}
}
