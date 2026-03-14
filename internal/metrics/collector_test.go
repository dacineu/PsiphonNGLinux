package metrics

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoticeCollector_Tunnels(t *testing.T) {
	m := NewMetrics()
	inner := &bytes.Buffer{}
	collector := NewNoticeCollector(m, inner)

	notice := map[string]interface{}{
		"noticeType": "Tunnels",
		"data": map[string]interface{}{
			"count": 2,
		},
	}
	jsonBytes, _ := json.Marshal(notice)
	collector.Write(append(jsonBytes, '\n'))

	val, err := m.GetActiveTunnels()
	require.NoError(t, err)
	require.Equal(t, 2.0, val)
}

func TestNoticeCollector_BytesTransferred(t *testing.T) {
	m := NewMetrics()
	inner := &bytes.Buffer{}
	collector := NewNoticeCollector(m, inner)

	notice := map[string]interface{}{
		"noticeType": "BytesTransferred",
		"data": map[string]interface{}{
			"diagnosticID": "test",
			"sent":         100,
			"received":     200,
		},
	}
	jsonBytes, _ := json.Marshal(notice)
	collector.Write(append(jsonBytes, '\n'))

	sent, _ := m.GetBytesSentTotal()
	received, _ := m.GetBytesReceivedTotal()
	require.Equal(t, 100.0, sent)
	require.Equal(t, 200.0, received)
}

func TestNoticeCollector_Error(t *testing.T) {
	m := NewMetrics()
	inner := &bytes.Buffer{}
	collector := NewNoticeCollector(m, inner)

	notice := map[string]interface{}{
		"noticeType": "Error",
		"data": map[string]interface{}{
			"message": "something went wrong",
		},
	}
	jsonBytes, _ := json.Marshal(notice)
	collector.Write(append(jsonBytes, '\n'))

	errors, _ := m.GetErrorsTotal()
	require.Equal(t, 1.0, errors)
}

func TestNoticeCollector_ConnectingServer(t *testing.T) {
	m := NewMetrics()
	inner := &bytes.Buffer{}
	collector := NewNoticeCollector(m, inner)

	notice := map[string]interface{}{
		"noticeType": "ConnectingServer",
		"data": map[string]interface{}{
			"some": "data",
		},
	}
	jsonBytes, _ := json.Marshal(notice)
	collector.Write(append(jsonBytes, '\n'))

	attempts, _ := m.GetConnectionAttempts()
	require.Equal(t, 1.0, attempts)
}

func TestNoticeCollector_MultipleNotices(t *testing.T) {
	m := NewMetrics()
	inner := &bytes.Buffer{}
	collector := NewNoticeCollector(m, inner)

	notices := []interface{}{
		map[string]interface{}{
			"noticeType": "ConnectingServer",
			"data":       map[string]interface{}{},
		},
		map[string]interface{}{
			"noticeType": "ConnectingServer",
			"data":       map[string]interface{}{},
		},
		map[string]interface{}{
			"noticeType": "Tunnels",
			"data": map[string]interface{}{
				"count": 1,
			},
		},
		map[string]interface{}{
			"noticeType": "Error",
			"data":       map[string]interface{}{},
		},
	}

	for _, n := range notices {
		jsonBytes, _ := json.Marshal(n)
		collector.Write(append(jsonBytes, '\n'))
	}

	attempts, _ := m.GetConnectionAttempts()
	require.Equal(t, 2.0, attempts)

	tunnels, _ := m.GetActiveTunnels()
	require.Equal(t, 1.0, tunnels)

	errors, _ := m.GetErrorsTotal()
	require.Equal(t, 1.0, errors)
}

func TestNoticeCollector_InvalidJSON(t *testing.T) {
	m := NewMetrics()
	inner := &bytes.Buffer{}
	collector := NewNoticeCollector(m, inner)

	// Write invalid JSON - should be ignored
	collector.Write([]byte("not json\n"))

	tunnels, _ := m.GetActiveTunnels()
	require.Equal(t, 0.0, tunnels)

	errors, _ := m.GetErrorsTotal()
	require.Equal(t, 0.0, errors)
}
