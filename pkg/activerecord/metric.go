package activerecord

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type ARMertics struct {
	timing     *prometheus.HistogramVec
	count      *prometheus.CounterVec
	errcount   *prometheus.CounterVec
	cumulative CumulativeMetricsInterface
}

type CumulativeMetricsInterface interface {
	TimeCumulativeMetric(ctx context.Context, name string, dur time.Duration)
	IncCumulativeMetric(ctx context.Context, name string, diff int32)
}

func NewARMetrics(metrics *prometheus.Registry, cumulativeMetrics CumulativeMetricsInterface) *ARMertics {
	timing := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "active_record",
		Name:      "request_duration_seconds",
		Help:      "The latency of the database requests",
		Buckets:   prometheus.DefBuckets,
	}, []string{"storage", "entity", "metric"})

	metrics.MustRegister(timing)

	count := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "active_record",
		Name:      "request_counters",
		Help:      "Any count of the database process requests",
	}, []string{"storage", "entity", "metric"})

	metrics.MustRegister(count)

	errcount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "active_record",
		Name:      "error_counters",
		Help:      "Any count of the database process requests",
	}, []string{"storage", "entity", "metric"})

	metrics.MustRegister(errcount)

	return &ARMertics{
		timing:     timing,
		count:      count,
		errcount:   errcount,
		cumulative: cumulativeMetrics,
	}
}

func (m *ARMertics) Timer(storage, entity string) MetricTimerInterface {
	now := time.Now()

	return &ARMetricTimer{
		backend:    m.timing,
		storage:    storage,
		entity:     entity,
		start:      now,
		last:       now,
		cumulative: m.cumulative,
	}
}

func (m *ARMertics) StatCount(storage, entity string) MetricStatCountInterface {
	return &ARMetricCount{
		backend: m.count,
		storage: storage,
		entity:  entity,
	}
}

func (m *ARMertics) ErrorCount(storage, entity string) MetricStatCountInterface {
	return &ARMetricCount{
		backend: m.errcount,
		storage: storage,
		entity:  entity,
	}
}

type ARMetricCount struct {
	backend *prometheus.CounterVec
	storage string
	entity  string
}

func (m *ARMetricCount) Inc(ctx context.Context, name string, val float64) {
	m.backend.WithLabelValues(m.storage, m.entity, name).Add(val)
}

type ARMetricTimer struct {
	backend    *prometheus.HistogramVec
	storage    string
	entity     string
	start      time.Time
	last       time.Time
	cumulative CumulativeMetricsInterface
}

func (m *ARMetricTimer) Timing(ctx context.Context, name string) {
	if !m.last.IsZero() {
		m.backend.WithLabelValues(m.storage, m.entity, name).Observe(float64(time.Since(m.last).Seconds()))
	}

	m.last = time.Now()
}

func (m *ARMetricTimer) Finish(ctx context.Context, name string) {
	dur := time.Since(m.start)

	m.backend.WithLabelValues(m.storage, m.entity, name).Observe(float64(dur.Seconds()))

	if m.cumulative != nil {
		m.cumulative.TimeCumulativeMetric(ctx, name, dur)
		m.cumulative.IncCumulativeMetric(ctx, name, 1)
	}
}
