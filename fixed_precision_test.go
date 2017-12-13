package prom

import (
	"math"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestNoPrecisionGauge(t *testing.T) {
	c := NewFixedPrecisionGauge(prometheus.GaugeOpts{
		Name: "test",
		Help: "test help",
	}, 0)

	c.Inc()
	var want float64 = 1
	if expected, got := want, c.Value(); expected != got {
		t.Errorf("Expected %v, got %v.", expected, got)
	}
	c.Add(0.9999999999999999999)
	want = 2
	if expected, got := want, c.Value(); expected != got {
		t.Errorf("Expected %v, got %v.", expected, got)
	}

	c.Sub(0.9999999999999999999 * 3)
	want = -1
	if expected, got := want, c.Value(); expected != got {
		t.Errorf("Expected %v, got %v.", expected, got)
	}

	m := &dto.Metric{}
	c.Write(m)

	if expected, got := `counter:<value:-1 > `, m.String(); expected != got {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFixedPrecisionAdd(t *testing.T) {
	c := NewFixedPrecisionGauge(prometheus.GaugeOpts{
		Name: "test",
		Help: "test help",
	}, 3)

	c.Inc()
	want := 1.0
	if expected, got := want, c.Value(); expected != got {
		t.Errorf("Expected %f, got %f.", expected, got)
	}
	c.Add(42.3)
	want = 43.3
	if expected, got := want, c.Value(); expected != got {
		t.Errorf("Expected %f, got %f.", expected, got)
	}

	m := &dto.Metric{}
	c.Write(m)

	if expected, got := `counter:<value:43.3 > `, m.String(); expected != got {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFixedPrecisionSub(t *testing.T) {
	c := NewFixedPrecisionGauge(prometheus.GaugeOpts{
		Name: "test",
		Help: "test help",
	}, 3)

	c.Dec()
	var want float64 = -1
	if expected, got := want, c.Value(); expected != got {
		t.Errorf("Expected %f, got %f.", expected, got)
	}
	c.Sub(42.3)
	want = -43.3
	if expected, got := want, c.Value(); expected != got {
		t.Errorf("Expected %f, got %f.", expected, got)
	}

	m := &dto.Metric{}
	c.Write(m)

	if expected, got := `counter:<value:-43.3 > `, m.String(); expected != got {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSetToCurrentTime(t *testing.T) {
	precs := []uint{}
	var i uint
	for i = 0; i < 10; i++ {
		precs = append(precs, i)
	}

	for _, prec := range precs {
		c := NewFixedPrecisionGauge(prometheus.GaugeOpts{
			Name: "test",
			Help: "test help",
		}, prec)

		c.SetToCurrentTime()
		n := time.Now()

		delta := math.Abs(c.Value() - float64(n.Unix()))
		if !(delta <= 1) {
			t.Fatalf("SetToCurrentTime at %v precision was off from time.Now(): got: %v, want: <= 1", prec, delta)
		}
	}
}

func TestCounterDirection(t *testing.T) {
	defer func() {
		if e := recover(); e == nil {
			t.Fatalf("did not panic and shold have")
		}
	}()

	c := NewFixedPrecisionCounter(prometheus.CounterOpts{
		Name: "test",
		Help: "test help",
	}, 3)
	c.Add(1)
	var want float64 = 1
	if expected, got := want, c.Value(); expected != got {
		t.Errorf("Expected %f, got %f.", expected, got)
	}
	c.Add(-1)
}
