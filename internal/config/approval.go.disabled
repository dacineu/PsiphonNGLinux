package config

import (
	"context"
	"reflect"
	"time"

	"github.com/gorilla/websocket"
	"github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon"
	"github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon/common/errors"
	"github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon/common/inproxy"
	"github.com/dacineu/PsiphonNGLinux/internal/notification"
)

// approvalRequest extends inproxy.ClientConnectionInfo with additional metadata
// for logging and audit at the approval server.
type approvalRequest struct {
	inproxy.ClientConnectionInfo
	Timestamp     string `json:"timestamp,omitempty"`
	DaemonVersion string `json:"daemon_version,omitempty"`
	DaemonPlatform string `json:"daemon_platform,omitempty"`
}

// approvalResponse represents the server's reply to an approval request.
type approvalResponse struct {
	Approved      bool     `json:"approved"`
	StrictFields  []string `json:"strict_fields,omitempty"`   // current list of required fields
	MissingStrict []string `json:"missing_strict,omitempty"` // fields that are strict but not sent by client
}

// approvalState tracks per-configuration state for the approval callback.
type approvalState struct {
	lastStrictSet map[string]struct{}
}

// setupApprovalCallback configures the InproxyApproveClientConnection callback
// if the configuration requires approval for in-proxy mode.
func setupApprovalCallback(ngConfig *Config, cfg *psiphon.Config) error {
	if ngConfig.InproxyMode != "proxy" || ngConfig.InproxyApprovalWebSocketURL == "" {
		return nil
	}

	approvalTimeout, err := time.ParseDuration(ngConfig.InproxyApprovalTimeout)
	if err != nil {
		return errors.Trace(err)
	}

	state := &approvalState{
		lastStrictSet: make(map[string]struct{}),
	}

	cfg.InproxyApproveClientConnection = func(info inproxy.ClientConnectionInfo) (bool, error) {
		ctx, cancel := context.WithTimeout(context.Background(), approvalTimeout)
		defer cancel()

		dialer := websocket.Dialer{}
		ws, _, err := dialer.DialContext(ctx, ngConfig.InproxyApprovalWebSocketURL, nil)
		if err != nil {
			psiphon.NoticeWarning("approval: WebSocket dial failed for connection %s: %v", info.ConnectionID, err)
			return false, nil
		}
		defer ws.Close()

		// Build request payload based on logging configuration
		req := &approvalRequest{}
		if ngConfig.ApprovalLogging.IncludeRawClientInfo {
			req.ClientConnectionInfo = info
		} else {
			req.ConnectionID = info.ConnectionID
		}
		if ngConfig.ApprovalLogging.IncludeTimestamp {
			req.Timestamp = time.Now().UTC().Format(time.RFC3339)
		}
		if ngConfig.ApprovalLogging.IncludeDaemonInfo {
			req.DaemonVersion = ngConfig.ClientVersion
			req.DaemonPlatform = ngConfig.ClientPlatform
		}

		// Set write deadline and send request
		ws.SetWriteDeadline(time.Now().Add(approvalTimeout))
		if err := ws.WriteJSON(req); err != nil {
			psiphon.NoticeWarning("approval: WebSocket write failed for connection %s: %v", info.ConnectionID, err)
			return false, nil
		}

		// Set read deadline and receive response
		ws.SetReadDeadline(time.Now().Add(approvalTimeout))
		var resp approvalResponse
		if err := ws.ReadJSON(&resp); err != nil {
			psiphon.NoticeWarning("approval: WebSocket read failed for connection %s: %v", info.ConnectionID, err)
			return false, nil
		}

		// Detect strict fields changes
		currentStrict := make(map[string]struct{}, len(resp.StrictFields))
		for _, f := range resp.StrictFields {
			currentStrict[f] = struct{}{}
		}
		if !reflect.DeepEqual(currentStrict, state.lastStrictSet) {
			// Send notification to GUI
			notification.Send(map[string]interface{}{
				"type":     "strictness_changed",
				"previous": state.lastStrictSet,
				"current":  currentStrict,
			})
			state.lastStrictSet = currentStrict
		}

		// Notify missing strict fields (server indicates which fields it required but didn't receive)
		if len(resp.MissingStrict) > 0 {
			// Log local warning as well
			psiphon.NoticeWarning("approval: missing strict fields for connection %s: %v", info.ConnectionID, resp.MissingStrict)
			notification.Send(map[string]interface{}{
				"type":          "missing_strict_fields",
				"connection_id": info.ConnectionID,
				"missing":       resp.MissingStrict,
			})
		}

		return resp.Approved, nil
	}
	return nil
}
