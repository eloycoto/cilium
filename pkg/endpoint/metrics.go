package endpoint

import (
	"math"
	"time"

	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/metrics"
	"github.com/cilium/cilium/pkg/spanstat"
)

type regenerationStatistics struct {
	success                bool
	totalTime              spanstat.SpanStat
	waitingForLock         spanstat.SpanStat
	waitingForCTClean      spanstat.SpanStat
	policyCalculation      spanstat.SpanStat
	proxyConfiguration     spanstat.SpanStat
	proxyPolicyCalculation spanstat.SpanStat
	proxyWaitForAck        spanstat.SpanStat
	bpfCompilation         spanstat.SpanStat
	mapSync                spanstat.SpanStat
	prepareBuild           spanstat.SpanStat
}

// SendMetrics get the curret statistics and push the info to the given prometheus metrics.
func (s *regenerationStatistics) SendMetrics() {

	metrics.EndpointCountRegenerating.Dec()

	if !s.success {
		// Endpoint regeneration failed, increase on failed metrics
		metrics.EndpointRegenerationCount.WithLabelValues(metrics.LabelValueOutcomeFail).Inc()
		return
	}

	metrics.EndpointRegenerationCount.WithLabelValues(metrics.LabelValueOutcomeSuccess).Inc()
	regenerateTimeSec := s.totalTime.Total().Seconds()
	metrics.EndpointRegenerationTime.Add(regenerateTimeSec)
	metrics.EndpointRegenerationTimeSquare.Add(math.Pow(regenerateTimeSec, 2))

	for scope, value := range s.GetMap() {
		metrics.EndpointRegenerationTimeStats.WithLabelValues(scope).Observe(value.Seconds())
	}
}

// GetMap returns a map where the key is the stats name and the value is the duration of the stat.
func (s *regenerationStatistics) GetMap() map[string]time.Duration {
	return map[string]time.Duration{
		"waitingForLock":         s.waitingForLock.Total(),
		"waitingForCTClean":      s.waitingForCTClean.Total(),
		"policyCalculation":      s.policyCalculation.Total(),
		"proxyConfiguration":     s.proxyConfiguration.Total(),
		"proxyPolicyCalculation": s.proxyPolicyCalculation.Total(),
		"proxyWaitForAck":        s.proxyWaitForAck.Total(),
		"bpfCompilation":         s.bpfCompilation.Total(),
		"mapSync":                s.mapSync.Total(),
		"prepareBuild":           s.prepareBuild.Total(),
		logfields.BuildDuration:  s.totalTime.Total(),
	}
}
