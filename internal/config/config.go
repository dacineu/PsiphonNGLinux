package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon"
	"github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon/common/errors"
	"github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon/common/parameters"
)

// Config represents the PsiphonNGLinux configuration
type Config struct {
	// DataDirectory is the root directory for persistent data
	DataDirectory string `json:"data_directory"`

	// Logging
	LogFile  string `json:"log_file"`
	LogLevel string `json:"log_level"` // "debug", "info", "warning", "error"

	// Network identifiers
	PropagationChannelId string `json:"propagation_channel_id"`
	SponsorId            string `json:"sponsor_id"`
	ClientVersion        string `json:"client_version,omitempty"`
	ClientPlatform       string `json:"client_platform,omitempty"`

	// Tunnel mode
	TunnelMode string `json:"tunnel_mode"` // "portforward" or "packet"

	// Port forward settings
	LocalSocksProxyPort    int    `json:"local_socks_proxy_port"`
	LocalHttpProxyPort     int    `json:"local_http_proxy_port"`
	ListenInterface        string `json:"listen_interface,omitempty"`
	DisableLocalSocksProxy bool   `json:"disable_local_socks_proxy"`
	DisableLocalHTTPProxy  bool   `json:"disable_local_http_proxy"`

	// Packet tunnel (TUN) settings
	PacketTunnelDeviceName            string `json:"packet_tunnel_device_name"`
	PacketTunnelAddressIPv4           string `json:"packet_tunnel_address_ipv4"`
	PacketTunnelAddressIPv6           string `json:"packet_tunnel_address_ipv6"`
	PacketTunnelDNSIPv4               string `json:"packet_tunnel_dns_ipv4"`
	PacketTunnelDNSIPv6               string `json:"packet_tunnel_dns_ipv6"`
	PacketTunnelGatewayIPv4           string `json:"packet_tunnel_gateway_ipv4"`
	PacketTunnelGatewayIPv6           string `json:"packet_tunnel_gateway_ipv6"`
	PacketTunnelMTU                   int    `json:"packet_tunnel_mtu"`
	PacketTunnelPersist               bool   `json:"packet_tunnel_persist"`
	PacketTunnelOffenseMode           bool   `json:"packet_tunnel_offense_mode"`
	PacketTunnelTransparentDNSIPv4    bool   `json:"packet_tunnel_transparent_dns_ipv4,omitempty"`
	PacketTunnelTransparentDNSIPv6    bool   `json:"packet_tunnel_transparent_dns_ipv6,omitempty"`
	PacketTunnelDetectDNSIPv4         bool   `json:"packet_tunnel_detect_dns_ipv4,omitempty"`
	PacketTunnelDetectDNSIPv6         bool   `json:"packet_tunnel_detect_dns_ipv6,omitempty"`
	PacketTunnelIncludeIPv4Routes     bool   `json:"packet_tunnel_include_ipv4_routes,omitempty"`
	PacketTunnelIncludeIPv6Routes     bool   `json:"packet_tunnel_include_ipv6_routes,omitempty"`
	PacketTunnelExcludeIPv4Routes     bool   `json:"packet_tunnel_exclude_ipv4_routes,omitempty"`
	PacketTunnelExcludeIPv6Routes     bool   `json:"packet_tunnel_exclude_ipv6_routes,omitempty"`
	PacketTunnelReplaceDefaultIPv4    bool   `json:"packet_tunnel_replace_default_ipv4_route,omitempty"`
	PacketTunnelReplaceDefaultIPv6    bool   `json:"packet_tunnel_replace_default_ipv6_route,omitempty"`

	// Region selection
	EgressRegion string `json:"egress_region"`

	// Split tunneling
	SplitTunnelOwnRegion   bool     `json:"split_tunnel_own_region"`
	SplitTunnelRegions     []string `json:"split_tunnel_regions"`
	SplitTunnelInclude     []string `json:"split_tunnel_include"`
	SplitTunnelExclude     []string `json:"split_tunnel_exclude"`

	// In-proxy (broker) mode
	InproxyMode                          string   `json:"inproxy_mode"` // "client", "proxy", or ""
	InproxyBrokerServerAddresses         []string `json:"inproxy_broker_server_addresses"`
	InproxyClientLimits                  int      `json:"inproxy_client_limits,omitempty"`
	InproxyCompartmentID                 string   `json:"inproxy_compartment_id"`
	InproxySessionPrivateKey             string   `json:"inproxy_session_private_key,omitempty"`
	InproxySessionRootObfuscationSecret  string   `json:"inproxy_session_root_obfuscation_secret,omitempty"`
	InproxyEgressInterface               string   `json:"inproxy_egress_interface,omitempty"`
	InproxyMTU                           int      `json:"inproxy_mtu,omitempty"`
	InproxyMaxCommonClients              int      `json:"inproxy_max_common_clients"`
	InproxyMaxPersonalClients            int      `json:"inproxy_max_personal_clients"`
	InproxyMaxCommonClientCompartmentIDs []string `json:"inproxy_max_common_client_compartment_ids,omitempty"`
	InproxyMaxPersonalClientCompartmentIDs []string `json:"inproxy_max_personal_client_compartment_ids,omitempty"`
	InproxyRateLimiting                  bool     `json:"inproxy_rate_limiting"`
	InproxyApprovalWebSocketURL          string   `json:"inproxy_approval_websocket_url,omitempty"`
	InproxyApprovalTimeout               string   `json:"inproxy_approval_timeout,omitempty"`
	InproxyClientTrafficUDPPort          int      `json:"inproxy_client_traffic_udp_port,omitempty"`

	// Rate limiting (client-side, for packet tunnel mode)
	LimitUpstreamBytesPerSecond   int `json:"limit_upstream_bytes_per_second"`
	LimitDownstreamBytesPerSecond int `json:"limit_downstream_bytes_per_second"`

	// Connection parameters
	EstablishTunnelTimeoutSeconds          int `json:"establish_tunnel_timeout_seconds"`
	EstablishTunnelPausePeriodSeconds      int `json:"establish_tunnel_pause_period_seconds"`
	ConnectionWorkerPoolSize               int `json:"connection_worker_pool_size"`
	StaggerConnectionWorkersMilliseconds   int `json:"stagger_connection_workers_milliseconds"`
	LimitIntensiveConnectionWorkers        int `json:"limit_intensive_connection_workers"`
	TunnelPoolSize                         int `json:"tunnel_pool_size"`

	// Remote server list
	DisableRemoteServerListFetcher        bool     `json:"disable_remote_server_list_fetcher"`
	RemoteServerListURLs                  []string `json:"remote_server_list_urls"`
	RemoteServerListSignaturePublicKey    string   `json:"remote_server_list_signature_public_key"`
	RemoteServerListDownloadFilename      string   `json:"remote_server_list_download_filename"`
	ObfuscatedServerListRootURLs          []string `json:"obfuscated_server_list_root_urls"`
	ObfuscatedServerListDownloadDirectory string   `json:"obfuscated_server_list_download_directory"`

	// Tactics
	DisableTactics               bool   `json:"disable_tactics"`
	TacticsWaitPeriod            string `json:"tactics_wait_period"`
	TacticsRetryPeriod           string `json:"tactics_retry_period"`
	TacticsRetryPeriodJitter    string `json:"tactics_retry_period_jitter"`
	UseDeviceDetectCode          bool   `json:"use_device_detect_code"`
	UseRegionDetectionCode       bool   `json:"use_region_detection_code"`

	// API parameters
	PsiphonAPIRequestTimeout                string `json:"psiphon_api_request_timeout"`
	PsiphonAPIStatusRequestPeriodMin        string `json:"psiphon_api_status_request_period_min"`
	PsiphonAPIStatusRequestPeriodMax        string `json:"psiphon_api_status_request_period_max"`
	PsiphonAPIConnectedRequestPeriod        string `json:"psiphon_api_connected_request_period"`
	PsiphonAPIPersistentStatsMaxCount       int    `json:"psiphon_api_persistent_stats_max_count"`
	PsiphonAPIKey                           string `json:"psiphon_api_key,omitempty"`
	PsiphonAPIServerAddress                 string `json:"psiphon_api_server_address,omitempty"`
	SkipServerSelectionOnAPI                 bool   `json:"skip_server_selection_on_api,omitempty"`
	EgressSubnets                            []string `json:"egress_subnets,omitempty"`

	// Network
	UpstreamProxyURL            string            `json:"upstream_proxy_url,omitempty"`
	CustomHeaders               map[string]string `json:"custom_headers,omitempty"`
	LimitTunnelProtocols        []string          `json:"limit_tunnel_protocols,omitempty"`
	InitialLimitTunnelProtocols []string          `json:"initial_limit_tunnel_protocols,omitempty"`
	InitialLimitTunnelProtocolsCandidateCount int    `json:"initial_limit_tunnel_protocols_candidate_count"`

	// Resources
	LimitCPUThreads              bool   `json:"limit_cpu_threads"`
	LimitRelayBufferSizes        bool   `json:"limit_relay_buffer_sizes"`
	LimitMeekBufferSizes         bool   `json:"limit_meek_buffer_sizes"`
	SpeedTestPaddingMinBytes     int    `json:"speed_test_padding_min_bytes"`
	SpeedTestPaddingMaxBytes     int    `json:"speed_test_padding_max_bytes"`
	SpeedTestMaxSampleCount      int    `json:"speed_test_max_sample_count"`

	// Diagnostics
	NetworkIDCallback func() (string, error) `json:"-"`
	GetNetworkID      func(ctx context.Context) (string, error) `json:"-"`

	// Other
	ClientFeatures []string `json:"client_features,omitempty"`
	BuildInfo      string   `json:"build_info,omitempty"`

	// Metrics
	MetricsEnabled bool   `json:"metrics_enabled"`
	MetricsPort    string `json:"metrics_port"`

	// ApprovalLogging controls what metadata is sent to the WSS approval server
	// for each connection approval request. This data is used by the server for
	// audit logging, statistics, and decision-making.
	ApprovalLogging ApprovalLoggingConfig `json:"approval_logging,omitempty"`

	// ApprovalNotification controls the local notification server for GUI alerts.
	ApprovalNotificationEnabled bool   `json:"approval_notification_enabled,omitempty"`
	ApprovalNotificationPort    string `json:"approval_notification_port,omitempty"`

	// Config file path (not serialized)
	configPath string `json:"-"`
}

// ApprovalLoggingConfig defines which fields are included in approval requests.
type ApprovalLoggingConfig struct {
	// IncludeTimestamp adds a UTC timestamp to the approval request.
	IncludeTimestamp bool `json:"include_timestamp,omitempty"`
	// IncludeDaemonInfo includes the daemon version and platform.
	IncludeDaemonInfo bool `json:"include_daemon_info,omitempty"`
	// IncludeRawClientInfo includes the raw ClientConnectionInfo fields.
	// This is typically always true as it contains essential connection details.
	IncludeRawClientInfo bool `json:"include_raw_client_info,omitempty"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DataDirectory: filepath.Join(os.Getenv("HOME"), ".local/var/lib/psiphon"),
		LogFile:       filepath.Join(os.Getenv("HOME"), ".local/var/log/psiphon/psiphond-ng.log"),
		LogLevel:      "notice",

		PropagationChannelId: "FFFFFFFFFFFFFFFF",
		SponsorId:            "FFFFFFFFFFFFFFFF",
		ClientVersion:        "1",
		ClientPlatform:       "linux",

		TunnelMode: "portforward",

		LocalSocksProxyPort: 1080,
		LocalHttpProxyPort:  8080,
		ListenInterface:     "",

		PacketTunnelDeviceName:     "tun-psiphon",
		PacketTunnelAddressIPv4:    "10.0.0.2/30",
		PacketTunnelDNSIPv4:        "8.8.8.8",
		PacketTunnelGatewayIPv4:    "10.0.0.1",
		PacketTunnelMTU:            1500,
		PacketTunnelPersist:        false,
		PacketTunnelOffenseMode:    false,

		EgressRegion: "",

		SplitTunnelOwnRegion: false,
		SplitTunnelRegions:   []string{},
		SplitTunnelInclude:   []string{},
		SplitTunnelExclude:   []string{},

		InproxyMode:                          "",
		InproxyBrokerServerAddresses:         []string{},
		InproxyCompartmentID:                 "",
		InproxyMaxCommonClients:              10,
		InproxyMaxPersonalClients:            5,
		InproxyApprovalTimeout:               "10s",
		InproxyClientTrafficUDPPort:          0,

		LimitUpstreamBytesPerSecond:   0,
		LimitDownstreamBytesPerSecond: 0,

		EstablishTunnelTimeoutSeconds:          300,
		EstablishTunnelPausePeriodSeconds:      5,
		ConnectionWorkerPoolSize:               0,
		StaggerConnectionWorkersMilliseconds:   0,
		LimitIntensiveConnectionWorkers:        0,
		TunnelPoolSize:                         1,

		DisableRemoteServerListFetcher: false,
		RemoteServerListURLs:                []string{"https://s3.amazonaws.com/psiphon/web/mjr4-p23r-puwl/server_list_compressed"},
		RemoteServerListSignaturePublicKey:  "MIICIDANBgkqhkiG9w0BAQEFAAOCAg0AMIICCAKCAgEAt7Ls+/39r+T6zNW7GiVpJfzq/xvL9SBH5rIFnk0RXYEYavax3WS6HOD35eTAqn8AniOwiH+DOkvgSKF2caqk/y1dfq47Pdymtwzp9ikpB1C5OfAysXzBiwVJlCdajBKvBZDerV1cMvRzCKvKwRmvDmHgphQQ7WfXIGbRbmmk6opMBh3roE42KcotLFtqp0RRwLtcBRNtCdsrVsjiI1Lqz/lH+T61sGjSjQ3CHMuZYSQJZo/KrvzgQXpkaCTdbObxHqb6/+i1qaVOfEsvjoiyzTxJADvSytVtcTjijhPEV6XskJVHE1Zgl+7rATr/pDQkw6DPCNBS1+Y6fy7GstZALQXwEDN/qhQI9kWkHijT8ns+i1vGg00Mk/6J75arLhqcodWsdeG/M/moWgqQAnlZAGVtJI1OgeF5fsPpXu4kctOfuZlGjVZXQNW34aOzm8r8S0eVZitPlbhcPiR4gT/aSMz/wd8lZlzZYsje/Jr8u/YtlwjjreZrGRmG8KMOzukV3lLmMppXFMvl4bxv6YFEmIuTsOhbLTwFgh7KYNjodLj/LsqRVfwz31PgWQFTEPICV7GCvgVlPRxnofqKSjgTWI4mxDhBpVcATvaoBl1L/6WLbFvBsoAUBItWwctO2xalKxF5szhGm8lccoc5MZr8kfE0uxMgsxz4er68iCID+rsCAQM=",
		RemoteServerListDownloadFilename:   "remote_server_list",
		ObfuscatedServerListRootURLs:        []string{"https://s3.amazonaws.com/psiphon/web/mjr4-p23r-puwl/obfuscated_server_lists"},
		ObfuscatedServerListDownloadDirectory: "obfuscated_server_lists",

		DisableTactics:               false,
		TacticsWaitPeriod:            "10s",
		TacticsRetryPeriod:           "1m",
		TacticsRetryPeriodJitter:    "0s",
		UseDeviceDetectCode:          false,
		UseRegionDetectionCode:       true,

		PsiphonAPIRequestTimeout:                "30s",
		PsiphonAPIStatusRequestPeriodMin:        "30s",
		PsiphonAPIStatusRequestPeriodMax:        "60s",
		PsiphonAPIConnectedRequestPeriod:        "5m",
		PsiphonAPIPersistentStatsMaxCount:       100,

		UpstreamProxyURL:            "",
		CustomHeaders:               map[string]string{},
		LimitTunnelProtocols:        []string{},
		InitialLimitTunnelProtocols: []string{},
		InitialLimitTunnelProtocolsCandidateCount: 0,

		LimitCPUThreads:           false,
		LimitRelayBufferSizes:     false,
		LimitMeekBufferSizes:      false,
		SpeedTestPaddingMinBytes:  10000,
		SpeedTestPaddingMaxBytes:  50000,
		SpeedTestMaxSampleCount:   5,

		ClientFeatures: []string{},

		MetricsEnabled: false,
		MetricsPort:    ":9100",

		ApprovalLogging: ApprovalLoggingConfig{
			IncludeTimestamp:      true,
			IncludeDaemonInfo:     true,
			IncludeRawClientInfo: true,
		},

		ApprovalNotificationEnabled: false,
		ApprovalNotificationPort:    "",
	}
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Tracef("failed to read config file: %s", err)
	}

	// Parse JSON
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Tracef("failed to parse config JSON: %s", err)
	}

	// Validate
	if err := validateConfig(config); err != nil {
		return nil, errors.Trace(err)
	}

	// Set derived paths
	if config.RemoteServerListDownloadFilename == "" {
		config.RemoteServerListDownloadFilename = "remote_server_list"
	}
	if config.ObfuscatedServerListDownloadDirectory == "" {
		config.ObfuscatedServerListDownloadDirectory = filepath.Join(config.DataDirectory, "obfuscated_server_lists")
	}

	return config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.Trace(err)
	}
	return os.WriteFile(path, data, 0600)
}

// validateConfig performs comprehensive validation of the configuration
func validateConfig(config *Config) error {
	// Required fields
	if config.PropagationChannelId == "" {
		return errors.TraceNew("propagation_channel_id is required")
	}
	if config.SponsorId == "" {
		return errors.TraceNew("sponsor_id is required")
	}

	// Validate tunnel mode
	if config.TunnelMode != "portforward" && config.TunnelMode != "packet" {
		return errors.Tracef("invalid tunnel_mode: %s (must be 'portforward' or 'packet')", config.TunnelMode)
	}

	// Validate in-proxy mode
	if config.InproxyMode != "" && config.InproxyMode != "client" && config.InproxyMode != "proxy" {
		return errors.Tracef("invalid inproxy_mode: %s (must be '', 'client', or 'proxy')", config.InproxyMode)
	}

	// Validate port numbers
	if err := validatePort(config.LocalSocksProxyPort, "local_socks_proxy_port"); err != nil {
		return err
	}
	if err := validatePort(config.LocalHttpProxyPort, "local_http_proxy_port"); err != nil {
		return err
	}

	// Validate durations (parse to ensure valid)
	durationFields := []struct {
		field string
		value string
	}{
		{"tactics_wait_period", config.TacticsWaitPeriod},
		{"tactics_retry_period", config.TacticsRetryPeriod},
		{"tactics_retry_period_jitter", config.TacticsRetryPeriodJitter},
		{"psiphon_api_request_timeout", config.PsiphonAPIRequestTimeout},
		{"psiphon_api_status_request_period_min", config.PsiphonAPIStatusRequestPeriodMin},
		{"psiphon_api_status_request_period_max", config.PsiphonAPIStatusRequestPeriodMax},
		{"psiphon_api_connected_request_period", config.PsiphonAPIConnectedRequestPeriod},
		{"inproxy_approval_timeout", config.InproxyApprovalTimeout},
	}
	for _, df := range durationFields {
		if df.value != "" {
			if _, err := time.ParseDuration(df.value); err != nil {
				return errors.Tracef("invalid %s: %v", df.field, err)
			}
		}
	}

	// Validate integer bounds
	if config.EstablishTunnelTimeoutSeconds <= 0 {
		return errors.TraceNew("establish_tunnel_timeout_seconds must be positive")
	}
	if config.EstablishTunnelPausePeriodSeconds <= 0 {
		return errors.TraceNew("establish_tunnel_pause_period_seconds must be positive")
	}
	if config.TunnelPoolSize < 0 || config.TunnelPoolSize > 32 {
		return errors.TraceNew("tunnel_pool_size must be between 0 and 32")
	}

	// Validate split tunneling CIDR ranges
	for _, cidr := range config.SplitTunnelInclude {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return errors.Tracef("invalid split_tunnel_include CIDR '%s': %v", cidr, err)
		}
	}
	for _, cidr := range config.SplitTunnelExclude {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return errors.Tracef("invalid split_tunnel_exclude CIDR '%s': %v", cidr, err)
		}
	}

	// Validate in-proxy configuration if mode is set
	if config.InproxyMode != "" {
		if err := validateInproxyConfig(config); err != nil {
			return err
		}
	}

	// Validate TUN configuration if packet mode
	if config.TunnelMode == "packet" {
		if err := validateTUNConfig(config); err != nil {
			return err
		}
	}

	// Validate data directory
	if config.DataDirectory == "" {
		return errors.TraceNew("data_directory is required")
	}

	// Validate log file path
	if config.LogFile == "" {
		return errors.TraceNew("log_file is required")
	}

	return nil
}

// validatePort checks if a port is in valid range (1-65535)
func validatePort(port int, fieldName string) error {
	if port < 0 || port > 65535 {
		return errors.Tracef("%s must be between 0 and 65535 (got %d)", fieldName, port)
	}
	return nil
}

// validateInproxyConfig validates in-proxy specific settings
func validateInproxyConfig(config *Config) error {
	if config.InproxyMode == "client" {
		if len(config.InproxyBrokerServerAddresses) == 0 {
			return errors.TraceNew("inproxy_broker_server_addresses is required when inproxy_mode is 'client'")
		}
		// Validate broker address format (host:port)
		for _, addr := range config.InproxyBrokerServerAddresses {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				return errors.Tracef("invalid broker server address '%s': %v", addr, err)
			}
		}
	}

	if config.InproxyMode == "proxy" {
		if config.InproxySessionPrivateKey == "" {
			return errors.TraceNew("inproxy_session_private_key is required when inproxy_mode is 'proxy'")
		}
		if config.InproxySessionRootObfuscationSecret == "" {
			return errors.TraceNew("inproxy_session_root_obfuscation_secret is required when inproxy_mode is 'proxy'")
		}
		if config.InproxyEgressInterface == "" {
			return errors.TraceNew("inproxy_egress_interface is required when inproxy_mode is 'proxy'")
		}
	}

	// Validate compartment ID if set
	if config.InproxyCompartmentID != "" && (len(config.InproxyCompartmentID) < 1 || len(config.InproxyCompartmentID) > 100) {
		return errors.TraceNew("inproxy_compartment_id must be between 1 and 100 characters")
	}

	// Validate rate limits if enabled
	if config.InproxyRateLimiting && (config.LimitUpstreamBytesPerSecond <= 0 || config.LimitDownstreamBytesPerSecond <= 0) {
		return errors.TraceNew("when inproxy_rate_limiting is true, limit_upstream_bytes_per_second and limit_downstream_bytes_per_second must be positive")
	}

	// Validate approval URL if set
	if config.InproxyApprovalWebSocketURL != "" {
		if _, err := url.ParseRequestURI(config.InproxyApprovalWebSocketURL); err != nil {
			return errors.Tracef("invalid inproxy_approval_websocket_url: %v", err)
		}
	}

	return nil
}

// validateTUNConfig validates packet tunnel (TUN) specific settings
func validateTUNConfig(config *Config) error {
	if config.PacketTunnelDeviceName == "" {
		return errors.TraceNew("packet_tunnel_device_name is required when tunnel_mode is 'packet'")
	}
	if config.PacketTunnelAddressIPv4 == "" && config.PacketTunnelAddressIPv6 == "" {
		return errors.TraceNew("at least one of packet_tunnel_address_ipv4 or packet_tunnel_address_ipv6 is required when tunnel_mode is 'packet'")
	}
	if config.PacketTunnelGatewayIPv4 == "" && config.PacketTunnelGatewayIPv6 == "" {
		return errors.TraceNew("at least one of packet_tunnel_gateway_ipv4 or packet_tunnel_gateway_ipv6 is required when tunnel_mode is 'packet'")
	}
	if config.PacketTunnelDNSIPv4 == "" && config.PacketTunnelDNSIPv6 == "" {
		return errors.TraceNew("at least one of packet_tunnel_dns_ipv4 or packet_tunnel_dns_ipv6 is required when tunnel_mode is 'packet'")
	}
	if config.PacketTunnelMTU <= 0 || config.PacketTunnelMTU > 9000 {
		return errors.TraceNew("packet_tunnel_mtu must be between 1 and 9000")
	}

	// Validate IPv4 CIDR if provided
	if config.PacketTunnelAddressIPv4 != "" {
		if _, _, err := net.ParseCIDR(config.PacketTunnelAddressIPv4); err != nil {
			return errors.Tracef("invalid packet_tunnel_address_ipv4 CIDR '%s': %v", config.PacketTunnelAddressIPv4, err)
		}
	}
	if config.PacketTunnelGatewayIPv4 != "" {
		if ip := net.ParseIP(strings.TrimSuffix(config.PacketTunnelGatewayIPv4, "/")); ip == nil {
			return errors.Tracef("invalid packet_tunnel_gateway_ipv4 '%s'", config.PacketTunnelGatewayIPv4)
		}
	}
	if config.PacketTunnelDNSIPv4 != "" {
		if ip := net.ParseIP(config.PacketTunnelDNSIPv4); ip == nil {
			return errors.Tracef("invalid packet_tunnel_dns_ipv4 '%s'", config.PacketTunnelDNSIPv4)
		}
	}

	// IPv6 validation
	if config.PacketTunnelAddressIPv6 != "" {
		if _, _, err := net.ParseCIDR(config.PacketTunnelAddressIPv6); err != nil {
			return errors.Tracef("invalid packet_tunnel_address_ipv6 CIDR '%s': %v", config.PacketTunnelAddressIPv6, err)
		}
	}
	if config.PacketTunnelGatewayIPv6 != "" {
		if ip := net.ParseIP(strings.TrimSuffix(config.PacketTunnelGatewayIPv6, "/")); ip == nil {
			return errors.Tracef("invalid packet_tunnel_gateway_ipv6 '%s'", config.PacketTunnelGatewayIPv6)
		}
	}
	if config.PacketTunnelDNSIPv6 != "" {
		if ip := net.ParseIP(config.PacketTunnelDNSIPv6); ip == nil {
			return errors.Tracef("invalid packet_tunnel_dns_ipv6 '%s'", config.PacketTunnelDNSIPv6)
		}
	}

	return nil
}

// buildPsiphonConfig converts PsiphonNGLinux config to psiphon.Config
func BuildPsiphonConfig(ngConfig *Config) (*psiphon.Config, error) {
	cfg := &psiphon.Config{
		// Data storage
		DataRootDirectory: ngConfig.DataDirectory,

		// Identifiers
		PropagationChannelId: ngConfig.PropagationChannelId,
		SponsorId:            ngConfig.SponsorId,
		ClientVersion:        ngConfig.ClientVersion,
		ClientPlatform:       ngConfig.ClientPlatform,
		ClientFeatures:       ngConfig.ClientFeatures,

		// Region
		EgressRegion: ngConfig.EgressRegion,

		// Split tunnel
		SplitTunnelOwnRegion: ngConfig.SplitTunnelOwnRegion,
		SplitTunnelRegions:   ngConfig.SplitTunnelRegions,

		// Proxies
		ListenInterface:            ngConfig.ListenInterface,
		DisableLocalSocksProxy:     ngConfig.DisableLocalSocksProxy,
		LocalSocksProxyPort:        ngConfig.LocalSocksProxyPort,
		DisableLocalHTTPProxy:      ngConfig.DisableLocalHTTPProxy,
		LocalHttpProxyPort:         ngConfig.LocalHttpProxyPort,
		PacketTunnelTransparentDNSIPv4Address: ngConfig.PacketTunnelDNSIPv4,
		PacketTunnelTransparentDNSIPv6Address: ngConfig.PacketTunnelDNSIPv6,

		// Network
		UpstreamProxyURL: ngConfig.UpstreamProxyURL,
		CustomHeaders:    make(http.Header),

		// Protocols
		LimitTunnelProtocols:                    ngConfig.LimitTunnelProtocols,
		InitialLimitTunnelProtocols:             ngConfig.InitialLimitTunnelProtocols,
		InitialLimitTunnelProtocolsCandidateCount: ngConfig.InitialLimitTunnelProtocolsCandidateCount,

		// Timeouts
		EstablishTunnelTimeoutSeconds:   &ngConfig.EstablishTunnelTimeoutSeconds,
		EstablishTunnelPausePeriodSeconds: &ngConfig.EstablishTunnelPausePeriodSeconds,

		// Workers
		ConnectionWorkerPoolSize:         ngConfig.ConnectionWorkerPoolSize,
		StaggerConnectionWorkersMilliseconds: ngConfig.StaggerConnectionWorkersMilliseconds,
		LimitIntensiveConnectionWorkers:  ngConfig.LimitIntensiveConnectionWorkers,

		// Tunnels
		TunnelPoolSize: ngConfig.TunnelPoolSize,

		// Resource limits (in-proxy only)
		InproxyLimitUpstreamBytesPerSecond:  ngConfig.LimitUpstreamBytesPerSecond,
		InproxyLimitDownstreamBytesPerSecond: ngConfig.LimitDownstreamBytesPerSecond,
		LimitCPUThreads:                     ngConfig.LimitCPUThreads,
		LimitRelayBufferSizes:               ngConfig.LimitRelayBufferSizes,
		LimitMeekBufferSizes:                ngConfig.LimitMeekBufferSizes,

	// Remote server list
	DisableRemoteServerListFetcher: ngConfig.DisableRemoteServerListFetcher,
	RemoteServerListDownloadFilename: filepath.Join(ngConfig.DataDirectory, ngConfig.RemoteServerListDownloadFilename),
	RemoteServerListSignaturePublicKey: ngConfig.RemoteServerListSignaturePublicKey,
	RemoteServerListURLs:                convertStringSliceToTransferURLs(ngConfig.RemoteServerListURLs),
	ObfuscatedServerListRootURLs:        convertStringSliceToTransferURLs(ngConfig.ObfuscatedServerListRootURLs),
	ObfuscatedServerListDownloadDirectory: ngConfig.ObfuscatedServerListDownloadDirectory,

		// Tactics
		DisableTactics: ngConfig.DisableTactics,

		// API: use defaults; can be overridden via parameters if needed

		// Split tunnel DNS (already set above)
	}

	// Convert custom headers
	for k, v := range ngConfig.CustomHeaders {
		cfg.CustomHeaders.Add(k, v)
	}

	if err := setupApprovalCallback(ngConfig, cfg); err != nil {
		return nil, errors.Trace(err)
	}

	// Commit the configuration to finalize and validate
	if err := cfg.Commit(false); err != nil {
		return nil, errors.Trace(err)
	}

	return cfg, nil
}

// convertStringSliceToTransferURLs converts a slice of URLs to parameters.TransferURLs.
// The URLs are base64-encoded as required by the TransferURL.URL field.
// All URLs have OnlyAfterAttempts=0 to be eligible from the first attempt.
func convertStringSliceToTransferURLs(urls []string) parameters.TransferURLs {
	result := make(parameters.TransferURLs, 0, len(urls))
	for _, url := range urls {
		if url != "" {
			// The TransferURL.URL field expects base64-encoded content to obfuscate
			// the URL in the binary. The DecodeAndValidate method will decode it.
			encoded := base64.StdEncoding.EncodeToString([]byte(url))
			result = append(result, &parameters.TransferURL{
				URL:             encoded,
				OnlyAfterAttempts: 0,
			})
		}
	}
	return result
}

// ConfigWatcher watches config file for changes and triggers reloads
type ConfigWatcher struct {
	configPath string
	config     *Config
	mu         sync.RWMutex
	reloadChan chan *Config
	watcher    *fsnotify.Watcher
}

func NewConfigWatcher(configPath string, initialConfig *Config) *ConfigWatcher {
	return &ConfigWatcher{
		configPath: configPath,
		config:     initialConfig,
		reloadChan: make(chan *Config, 1),
	}
}

func (cw *ConfigWatcher) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Trace(err)
	}
	cw.watcher = watcher

	// Watch directory containing config file
	dir := filepath.Dir(cw.configPath)
	if err := watcher.Add(dir); err != nil {
		watcher.Close()
		return errors.Trace(err)
	}

	go func() {
		defer watcher.Close()
		for event := range watcher.Events {
			if event.Name == cw.configPath && (event.Op&fsnotify.Write != 0 || event.Op&fsnotify.Create != 0) {
				// Debounce: wait a bit for file to be fully written
				time.Sleep(500 * time.Millisecond)

				// Reload config
				newConfig, err := LoadConfig(cw.configPath)
				if err != nil {
					log.Printf("Failed to reload config: %v", err)
					continue
				}

				cw.mu.Lock()
				old := cw.config
				cw.config = newConfig
				cw.mu.Unlock()

				select {
				case cw.reloadChan <- newConfig:
					log.Println("Configuration reloaded successfully")
				default:
					// Channel full, drop but still log
					log.Println("Configuration reloaded (notification dropped)")
				}

				// TODO: trigger controller param update or restart based on changes
				_ = old
			}
		}
	}()

	return nil
}

func (cw *ConfigWatcher) Stop() {
	if cw.watcher != nil {
		cw.watcher.Close()
	}
}

func (cw *ConfigWatcher) Get() *Config {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.config
}

func (cw *ConfigWatcher) ReloadChan() <-chan *Config {
	return cw.reloadChan
}

