package clients

import (
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

type PhaseName string

const (
	StepNameDns  PhaseName = "DNS"
	StepNameConn PhaseName = "CONN"
	StepNameTls  PhaseName = "TLS"
)

type PhaseStats struct {
	StepName string
	Start    time.Time
	Done     time.Time
	Error    error
}

func (p *PhaseStats) Duration() time.Duration {
	return p.Done.Sub(p.Start)
}

type HttpClientTraceWrapper struct {
	phaseStats map[PhaseName]*PhaseStats
}

func (h *HttpClientTraceWrapper) startPhase(stepName PhaseName) {
	h.phaseStats[stepName] = &PhaseStats{
		StepName: string(stepName),
		Start:    time.Now(),
	}
}

func (h *HttpClientTraceWrapper) donePhase(stepName PhaseName) {
	h.phaseStats[stepName].Done = time.Now()
}

func (h *HttpClientTraceWrapper) errorPhase(stepName PhaseName, err error) {
	h.phaseStats[stepName].Error = err
}

func (h *HttpClientTraceWrapper) Stats() map[PhaseName]PhaseStats {
	stats := make(map[PhaseName]PhaseStats)

	for _, value := range h.phaseStats {
		if value != nil {
			stats[PhaseName(value.StepName)] = *value
		}
	}

	return stats
}

func (h *HttpClientTraceWrapper) NewClientTrace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			h.startPhase(StepNameDns)
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			h.donePhase(StepNameDns)
		},
		GetConn: func(hostPort string) {
			h.startPhase(StepNameConn)
		},
		ConnectDone: func(network, addr string, err error) {
			h.donePhase(StepNameConn)
			if err != nil {
				h.errorPhase(StepNameConn, err)
			}
		},
		TLSHandshakeStart: func() {
			h.startPhase(StepNameTls)
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			h.donePhase(StepNameTls)
			if err != nil {
				h.errorPhase(StepNameTls, err)
			}
		},
	}
}

func NewHttpClientTraceWrapper() HttpClientTraceWrapper {
	return HttpClientTraceWrapper{
		phaseStats: make(map[PhaseName]*PhaseStats),
	}
}
