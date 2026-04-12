// Copyright (c) 2026 PhysicsCopilot contributors. All rights reserved.
// SPDX-License-Identifier: MIT

package metrics

import (
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
)

// RuntimeCollector is a prometheus.Collector that exports Go runtime metrics:
// goroutine count, heap objects, GC pause time, and memory allocations.
// Register it once at startup via prometheus.MustRegister(NewRuntimeCollector()).
type RuntimeCollector struct {
	goroutines   *prometheus.Desc
	heapObjects  *prometheus.Desc
	gcPauseTotal *prometheus.Desc
	allocBytes   *prometheus.Desc
}

// NewRuntimeCollector returns an initialised RuntimeCollector ready for
// registration with a Prometheus registry.
func NewRuntimeCollector() *RuntimeCollector {
	return &RuntimeCollector{
		goroutines:   prometheus.NewDesc("go_goroutines_custom", "Number of goroutines.", nil, nil),
		heapObjects:  prometheus.NewDesc("go_heap_objects_custom", "Number of heap objects.", nil, nil),
		gcPauseTotal: prometheus.NewDesc("go_gc_pause_total_ns_custom", "Total GC pause time in nanoseconds.", nil, nil),
		allocBytes:   prometheus.NewDesc("go_alloc_bytes_custom", "Bytes of allocated heap objects.", nil, nil),
	}
}

// Describe sends the descriptors of each metric to the channel.
// It implements prometheus.Collector.
func (c *RuntimeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.goroutines
	ch <- c.heapObjects
	ch <- c.gcPauseTotal
	ch <- c.allocBytes
}

// Collect reads current runtime stats and sends metric values to the channel.
// It implements prometheus.Collector.
func (c *RuntimeCollector) Collect(ch chan<- prometheus.Metric) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	ch <- prometheus.MustNewConstMetric(c.goroutines, prometheus.GaugeValue, float64(runtime.NumGoroutine()))
	ch <- prometheus.MustNewConstMetric(c.heapObjects, prometheus.GaugeValue, float64(m.HeapObjects))
	ch <- prometheus.MustNewConstMetric(c.gcPauseTotal, prometheus.CounterValue, float64(m.PauseTotalNs))
	ch <- prometheus.MustNewConstMetric(c.allocBytes, prometheus.GaugeValue, float64(m.Alloc))
}
