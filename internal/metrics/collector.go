package metrics

import (
	"encoding/json"
	"io"
	"strings"
)

// NoticeCollector is an io.Writer that intercepts Psiphon notices and updates metrics
type NoticeCollector struct {
	metrics *Metrics
	inner   io.Writer
}

// NewNoticeCollector creates a new collector that wraps the given writer
func NewNoticeCollector(metrics *Metrics, inner io.Writer) *NoticeCollector {
	return &NoticeCollector{
		metrics: metrics,
		inner:   inner,
	}
}

// Write implements io.Writer. It intercepts JSON notices, updates metrics, and forwards to inner writer.
func (nc *NoticeCollector) Write(p []byte) (n int, err error) {
	// First, forward to inner writer
	if nc.inner != nil {
		n, _ = nc.inner.Write(p)
	}

	// Try to parse the notice and update metrics
	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var notice struct {
			NoticeType string          `json:"noticeType"`
			Data       json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal([]byte(line), &notice); err != nil {
			continue // not a valid JSON notice, skip
		}
		nc.processNotice(notice.NoticeType, notice.Data)
	}
	return n, err
}

// processNotice updates metrics based on the notice type and data
func (nc *NoticeCollector) processNotice(noticeType string, data json.RawMessage) {
	switch noticeType {
	case "Tunnels":
		var payload struct {
			Count int `json:"count"`
		}
		if err := json.Unmarshal(data, &payload); err == nil {
			nc.metrics.ActiveTunnels.Set(float64(payload.Count))
		}

	case "BytesTransferred":
		var payload struct {
			Sent     int64 `json:"sent"`
			Received int64 `json:"received"`
		}
		if err := json.Unmarshal(data, &payload); err == nil {
			nc.metrics.BytesSentTotal.Add(float64(payload.Sent))
			nc.metrics.BytesReceivedTotal.Add(float64(payload.Received))
		}

	case "TotalBytesTransferred":
		var payload struct {
			Sent     int64 `json:"sent"`
			Received int64 `json:"received"`
		}
		if err := json.Unmarshal(data, &payload); err == nil {
			nc.metrics.BytesSentTotal.Add(float64(payload.Sent))
			nc.metrics.BytesReceivedTotal.Add(float64(payload.Received))
		}

	case "Error":
		nc.metrics.ErrorsTotal.Inc()

	case "ConnectingServer":
		nc.metrics.ConnectionAttempts.Inc()
	}
}

// Ensure NoticeCollector implements io.Writer
var _ io.Writer = (*NoticeCollector)(nil)
