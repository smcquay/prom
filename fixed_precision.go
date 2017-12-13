// Package prom exports a collection of prometheus.Metrics that use int64 and
// atomic store and add for maximum speed.
//
// In testing it provides a 10x speedup compared to
// github.com/prometheus/client_golang/prometheus.Metrics.
package prom

import (
	"math"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// NewCounter returns prometheus.Counter backed by a fixed-precision int64 that
// uses atomic operations.
func NewCounter(opts prometheus.CounterOpts, prec uint) prometheus.Counter {
	return NewFixedPrecisionCounter(opts, prec)
}

// NewGauge returns prometheus.Gauge backed by a fixed-precision int64 that
// uses atomic operations.
func NewGauge(opts prometheus.GaugeOpts, prec uint) prometheus.Gauge {
	return NewFixedPrecisionGauge(opts, prec)
}

// FixedPrecisionGauge implements a prometheus Gauge/Counter metric that uses atomic
// adds and stores for speed.
type FixedPrecisionGauge struct {
	val  int64
	prec uint

	desc *prometheus.Desc
}

// NewFixedPrecisionGauge returns a populated fixed-precision counter.
func NewFixedPrecisionGauge(opts prometheus.GaugeOpts, prec uint) *FixedPrecisionGauge {
	desc := prometheus.NewDesc(
		prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		nil,
		opts.ConstLabels,
	)
	return &FixedPrecisionGauge{
		desc: desc,
		prec: uint(math.Pow10(int(prec))),
	}
}

// Set stores the value in the counter.
func (fpg *FixedPrecisionGauge) Set(val float64) {
	atomic.StoreInt64(&fpg.val, int64(val)*int64(fpg.prec))
}

// add maps delta into the appropriate precision and adds it to val.
func (fpg *FixedPrecisionGauge) add(delta int64) {
	atomic.AddInt64(&fpg.val, delta*int64(fpg.prec))
}

// Inc adds 1 to the counter.
func (fpg *FixedPrecisionGauge) Inc() {
	fpg.add(1)
}

// Dec decrements 1 from the counter.
func (fpg *FixedPrecisionGauge) Dec() {
	fpg.add(-1)
}

// Add generically adds delta to the value stored by counter.
func (fpg *FixedPrecisionGauge) Add(delta float64) {
	atomic.AddInt64(&fpg.val, int64(delta*float64(fpg.prec)))
}

// Sub is the inverse of Add.
func (fpg *FixedPrecisionGauge) Sub(val float64) {
	fpg.Add(val * -1)
}

// Write is implemented to be useful as a prometheus counter.
func (fpg *FixedPrecisionGauge) Write(out *dto.Metric) error {
	f := float64(atomic.LoadInt64(&fpg.val)) / float64(fpg.prec)
	out.Counter = &dto.Counter{Value: &f}
	return nil
}

// Value returns a float64 representation of the current value stored.
func (fpg *FixedPrecisionGauge) Value() float64 {
	return float64(atomic.LoadInt64(&fpg.val)) / float64(fpg.prec)
}

// The following three methods exist to make this behave with Prometheus

// Desc returns this FixedPrecisionGauge's prometheus description.
func (fpg *FixedPrecisionGauge) Desc() *prometheus.Desc {
	return fpg.desc
}

// Describe sends the counter's description to the chan
func (fpg *FixedPrecisionGauge) Describe(dc chan<- *prometheus.Desc) {
	dc <- fpg.desc
}

// Collect sends the counter value to the chan
func (fpg *FixedPrecisionGauge) Collect(mc chan<- prometheus.Metric) {
	mc <- fpg
}

// SetToCurrentTime sets the Gauge to the current Unix time in seconds.
//
// Beware that if precision is set too high (greater than 9) it can overflow
// the underlying int64.
func (fpg *FixedPrecisionGauge) SetToCurrentTime() {
	fpg.Set(float64(time.Now().Unix()))
}

// FixedPrecisionCounter embeds FixedPrecisionGauge and enforces the same
// guarantees as a prometheus.Counter where negative adds panic.
type FixedPrecisionCounter struct {
	FixedPrecisionGauge
}

// NewFixedPrecisionCounter creates a FixedPrecisionCounter based on the
// provided Opts. It matches prometheus.Counter behavior by panicking if adding
// negative numbers.
func NewFixedPrecisionCounter(opts prometheus.CounterOpts, prec uint) *FixedPrecisionCounter {
	desc := prometheus.NewDesc(
		prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		nil,
		opts.ConstLabels,
	)
	return &FixedPrecisionCounter{
		FixedPrecisionGauge: FixedPrecisionGauge{
			desc: desc,
			prec: uint(math.Pow10(int(prec))),
		},
	}

}

// Add adds the given value to the counter. It matches prometheus.Counter
// behavior by panicking if the value is < 0.
func (fpc *FixedPrecisionCounter) Add(v float64) {
	if v < 0 {
		panic("counter cannot decrease in value")
	}
	fpc.FixedPrecisionGauge.Add(v)
}
