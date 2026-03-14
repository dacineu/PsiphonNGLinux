package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/test.conf"

	validConfig := `{
  "data_directory": "/var/lib/psiphon",
  "log_file": "/var/log/psiphon/psiphond-ng.log",
  "log_level": "info",
  "propagation_channel_id": "FFFFFFFFFFFFFFFF",
  "sponsor_id": "FFFFFFFFFFFFFFFF",
  "client_version": "1",
  "client_platform": "linux",
  "tunnel_mode": "portforward",
  "local_socks_proxy_port": 1080,
  "local_http_proxy_port": 8080,
  "listen_interface": "127.0.0.1",
  "packet_tunnel_device_name": "tun-psiphon",
  "packet_tunnel_address_ipv4": "10.0.0.2/30",
  "packet_tunnel_dns_ipv4": "8.8.8.8",
  "packet_tunnel_gateway_ipv4": "10.0.0.1",
  "packet_tunnel_mtu": 1500,
  "metrics_enabled": false,
  "metrics_port": ":9100"
}`

	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.Equal(t, "/var/lib/psiphon", cfg.DataDirectory)
	require.Equal(t, "info", cfg.LogLevel)
	require.Equal(t, "FFFFFFFFFFFFFFFF", cfg.PropagationChannelId)
	require.Equal(t, "portforward", cfg.TunnelMode)
	require.Equal(t, 1080, cfg.LocalSocksProxyPort)
	require.False(t, cfg.MetricsEnabled)
	require.Equal(t, ":9100", cfg.MetricsPort)
}

func TestValidateConfig_MissingRequired(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PropagationChannelId = ""
	err := validateConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "propagation_channel_id is required")

	cfg = DefaultConfig()
	cfg.SponsorId = ""
	err = validateConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sponsor_id is required")
}

func TestValidateConfig_InvalidTunnelMode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TunnelMode = "invalid"
	err := validateConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid tunnel_mode")
}

func TestValidateConfig_InvalidPort(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LocalSocksProxyPort = 99999
	err := validateConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "local_socks_proxy_port must be between 0 and 65535")
}

func TestValidateConfig_InvalidDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TacticsWaitPeriod = "invalid"
	err := validateConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid tactics_wait_period")
}

func TestValidateConfig_InvalidSplitTunnelCIDR(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SplitTunnelInclude = []string{"not a cidr"}
	err := validateConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid split_tunnel_include CIDR")
}

func TestValidateConfig_InvalidInproxyClient(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InproxyMode = "client"
	// No broker addresses
	err := validateConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "inproxy_broker_server_addresses is required")
}

func TestValidateConfig_InvalidInproxyProxy(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DataDirectory = t.TempDir()
	cfg.DisableRemoteServerListFetcher = true
	cfg.InproxyMode = "proxy"
	cfg.InproxySessionPrivateKey = "key"
	cfg.InproxySessionRootObfuscationSecret = "secret"
	// Missing egress interface
	err := validateConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "inproxy_egress_interface is required")
}

func TestValidateConfig_ValidInproxy(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InproxyMode = "client"
	cfg.InproxyBrokerServerAddresses = []string{"example.com:1234"}
	err := validateConfig(cfg)
	require.NoError(t, err)
}

func TestValidateConfig_PacketTunnelMissingFields(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TunnelMode = "packet"
	cfg.PacketTunnelDeviceName = ""
	cfg.PacketTunnelAddressIPv4 = ""
	cfg.PacketTunnelGatewayIPv4 = ""
	cfg.PacketTunnelDNSIPv4 = ""
	err := validateConfig(cfg)
	require.Error(t, err)
	// The validation stops at the first error
	require.Contains(t, err.Error(), "packet_tunnel_device_name is required")
}

func TestValidateConfig_PacketTunnelInvalidMTU(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TunnelMode = "packet"
	cfg.PacketTunnelDeviceName = "tun0"
	cfg.PacketTunnelAddressIPv4 = "10.0.0.2/30"
	cfg.PacketTunnelGatewayIPv4 = "10.0.0.1"
	cfg.PacketTunnelDNSIPv4 = "8.8.8.8"
	cfg.PacketTunnelMTU = 0
	err := validateConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "packet_tunnel_mtu must be between 1 and 9000")
}

func TestValidateConfig_PacketTunnelValid(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TunnelMode = "packet"
	cfg.PacketTunnelDeviceName = "tun0"
	cfg.PacketTunnelAddressIPv4 = "10.0.0.2/30"
	cfg.PacketTunnelGatewayIPv4 = "10.0.0.1"
	cfg.PacketTunnelDNSIPv4 = "8.8.8.8"
	cfg.PacketTunnelMTU = 1500
	err := validateConfig(cfg)
	require.NoError(t, err)
}

func TestBuildPsiphonConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PropagationChannelId = "CHANNEL123"
	cfg.SponsorId = "SPONSOR456"
	cfg.ClientVersion = "2"
	cfg.DataDirectory = "/tmp/psiphon"
	// Disable remote server list fetching to avoid signature requirement
	cfg.DisableRemoteServerListFetcher = true

	psiphonCfg, err := BuildPsiphonConfig(cfg)
	require.NoError(t, err)
	require.Equal(t, "CHANNEL123", psiphonCfg.PropagationChannelId)
	require.Equal(t, "SPONSOR456", psiphonCfg.SponsorId)
	require.Equal(t, "2", psiphonCfg.ClientVersion)
	require.Equal(t, "/tmp/psiphon", psiphonCfg.DataRootDirectory)
}

func TestBuildPsiphonConfig_MapsAllFields(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PropagationChannelId = "TESTCHANNEL"
	cfg.SponsorId = "TESTSPONSOR"
	cfg.ClientVersion = "3"
	cfg.ClientPlatform = "linux"
	cfg.DataDirectory = t.TempDir()
	cfg.EgressRegion = "us"
	cfg.SplitTunnelOwnRegion = true
	cfg.SplitTunnelRegions = []string{"us", "ca"}
	cfg.ListenInterface = "0.0.0.0"
	cfg.LocalSocksProxyPort = 1081
	cfg.LocalHttpProxyPort = 8081
	cfg.DisableLocalSocksProxy = true
	cfg.DisableLocalHTTPProxy = true
	cfg.PacketTunnelDNSIPv4 = "1.1.1.1"
	cfg.UpstreamProxyURL = "http://proxy:8080"
	cfg.CustomHeaders = map[string]string{"X-Custom": "value"}
	cfg.EstablishTunnelTimeoutSeconds = 120
	cfg.ConnectionWorkerPoolSize = 5
	cfg.TunnelPoolSize = 2
	cfg.DisableRemoteServerListFetcher = true
	cfg.ObfuscatedServerListDownloadDirectory = "custom_obfs"
	cfg.DisableTactics = true
	cfg.LimitCPUThreads = true
	cfg.LimitRelayBufferSizes = true

	psiphonCfg, err := BuildPsiphonConfig(cfg)
	require.NoError(t, err)
	require.Equal(t, "TESTCHANNEL", psiphonCfg.PropagationChannelId)
	require.Equal(t, "TESTSPONSOR", psiphonCfg.SponsorId)
	require.Equal(t, "3", psiphonCfg.ClientVersion)
	require.Equal(t, "linux", psiphonCfg.ClientPlatform)
	require.Equal(t, cfg.DataDirectory, psiphonCfg.DataRootDirectory)
	require.Equal(t, "us", psiphonCfg.EgressRegion)
	require.True(t, psiphonCfg.SplitTunnelOwnRegion)
	require.Equal(t, []string{"us", "ca"}, psiphonCfg.SplitTunnelRegions)
	require.Equal(t, "0.0.0.0", psiphonCfg.ListenInterface)
	require.Equal(t, 1081, psiphonCfg.LocalSocksProxyPort)
	require.Equal(t, 8081, psiphonCfg.LocalHttpProxyPort)
	require.True(t, psiphonCfg.DisableLocalSocksProxy)
	require.True(t, psiphonCfg.DisableLocalHTTPProxy)
	require.Equal(t, "1.1.1.1", psiphonCfg.PacketTunnelTransparentDNSIPv4Address)
	require.Equal(t, "http://proxy:8080", psiphonCfg.UpstreamProxyURL)
	require.Contains(t, psiphonCfg.CustomHeaders, "X-Custom")
	require.Equal(t, 120, *psiphonCfg.EstablishTunnelTimeoutSeconds)
	require.Equal(t, 5, psiphonCfg.ConnectionWorkerPoolSize)
	require.Equal(t, 2, psiphonCfg.TunnelPoolSize)
	require.True(t, psiphonCfg.DisableRemoteServerListFetcher)
	require.Contains(t, psiphonCfg.ObfuscatedServerListDownloadDirectory, "custom_obfs")
	require.True(t, psiphonCfg.DisableTactics)
	require.True(t, psiphonCfg.LimitCPUThreads)
	require.True(t, psiphonCfg.LimitRelayBufferSizes)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.Equal(t, "/var/lib/psiphon", cfg.DataDirectory)
	require.Equal(t, "/var/log/psiphon/psiphond-ng.log", cfg.LogFile)
	require.Equal(t, "portforward", cfg.TunnelMode)
	require.Equal(t, 1080, cfg.LocalSocksProxyPort)
	require.Equal(t, 8080, cfg.LocalHttpProxyPort)
	require.False(t, cfg.MetricsEnabled)
	require.Equal(t, ":9100", cfg.MetricsPort)
}

func TestConfigWatcher(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/watch.conf"

	// Write initial config
	initial := `{
  "data_directory": "/tmp/data1",
  "log_file": "/tmp/log1",
  "propagation_channel_id": "FFFFFFFFFFFFFFFF",
  "sponsor_id": "FFFFFFFFFFFFFFFF",
  "client_version": "1",
  "client_platform": "linux",
  "tunnel_mode": "portforward",
  "local_socks_proxy_port": 1080,
  "local_http_proxy_port": 8080,
  "packet_tunnel_mtu": 1500,
  "metrics_enabled": false,
  "metrics_port": ":9100"
}`
	err := os.WriteFile(configPath, []byte(initial), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)

	watcher := NewConfigWatcher(configPath, cfg)
	err = watcher.Start()
	require.NoError(t, err)
	defer watcher.Stop()

	// Wait a bit for watcher to be ready
	time.Sleep(200 * time.Millisecond)

	// Modify the config file (change port from 1080 to 1082)
	modified := strings.Replace(initial, "1080", "1082", 1)
	err = os.WriteFile(configPath, []byte(modified), 0644)
	require.NoError(t, err)

	// Wait for reload to be detected
	select {
	case newCfg := <-watcher.ReloadChan():
		require.NotNil(t, newCfg)
		require.Equal(t, 1082, newCfg.LocalSocksProxyPort)
		require.Equal(t, "/tmp/data1", newCfg.DataDirectory)
	case <-time.After(5 * time.Second):
		t.Fatal("Config reload not detected within timeout")
	}
}

func TestValidatePort(t *testing.T) {
	// Test edge cases for port validation indirectly through validateConfig
	cfg := DefaultConfig()
	cfg.LocalSocksProxyPort = 0
	require.NoError(t, validateConfig(cfg)) // port 0 is allowed (0 means disabled in some contexts)

	cfg.LocalSocksProxyPort = 65535
	require.NoError(t, validateConfig(cfg))

	cfg.LocalSocksProxyPort = 65536
	require.Error(t, validateConfig(cfg))

	cfg.LocalSocksProxyPort = -1
	require.Error(t, validateConfig(cfg))
}

func TestValidateConfig_DataDirAndLogFile(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DataDirectory = ""
	require.Error(t, validateConfig(cfg))

	cfg = DefaultConfig()
	cfg.LogFile = ""
	require.Error(t, validateConfig(cfg))
}

func TestValidateConfig_TunnelPoolSize(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TunnelPoolSize = 0
	require.NoError(t, validateConfig(cfg))

	cfg.TunnelPoolSize = 32
	require.NoError(t, validateConfig(cfg))

	cfg.TunnelPoolSize = 33
	require.Error(t, validateConfig(cfg))

	cfg.TunnelPoolSize = -1
	require.Error(t, validateConfig(cfg))
}

func TestValidateConfig_EstablishTimeouts(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EstablishTunnelTimeoutSeconds = 0
	require.Error(t, validateConfig(cfg))

	cfg = DefaultConfig()
	cfg.EstablishTunnelPausePeriodSeconds = 0
	require.Error(t, validateConfig(cfg))
}

func TestValidateConfig_InproxyCompartmentID(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InproxyMode = "client"
	cfg.InproxyBrokerServerAddresses = []string{"example.com:1234"}
	cfg.InproxyCompartmentID = strings.Repeat("a", 101)
	require.Error(t, validateConfig(cfg))

	cfg.InproxyCompartmentID = "valid"
	require.NoError(t, validateConfig(cfg))
}

func TestValidateConfig_InproxyRateLimiting(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InproxyMode = "client"
	cfg.InproxyBrokerServerAddresses = []string{"example.com:1234"}
	cfg.InproxyRateLimiting = true
	cfg.LimitUpstreamBytesPerSecond = 0
	cfg.LimitDownstreamBytesPerSecond = 0
	require.Error(t, validateConfig(cfg))

	cfg.LimitUpstreamBytesPerSecond = 1000
	cfg.LimitDownstreamBytesPerSecond = 1000
	require.NoError(t, validateConfig(cfg))
}

func TestValidateConfig_InvalidInproxyApprovalURL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InproxyMode = "client"
	cfg.InproxyApprovalWebSocketURL = "not a valid url"
	require.Error(t, validateConfig(cfg))
}

func TestValidateConfig_PacketTunnelIPv6(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TunnelMode = "packet"
	cfg.PacketTunnelDeviceName = "tun0"
	cfg.PacketTunnelAddressIPv6 = "2001:db8::1/64"
	cfg.PacketTunnelGatewayIPv6 = "2001:db8::1"
	cfg.PacketTunnelDNSIPv6 = "2001:4860:4860::8888"
	cfg.PacketTunnelMTU = 1500
	cfg.PacketTunnelGatewayIPv4 = "10.0.0.1"
	cfg.PacketTunnelDNSIPv4 = "8.8.8.8"
	require.NoError(t, validateConfig(cfg))

	// Invalid IPv6 CIDR
	cfg.PacketTunnelAddressIPv6 = "invalid"
	require.Error(t, validateConfig(cfg))
}

func TestValidateConfig_PacketTunnelBothFamilies(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TunnelMode = "packet"
	cfg.PacketTunnelDeviceName = "tun0"
	cfg.PacketTunnelAddressIPv4 = "10.0.0.2/30"
	cfg.PacketTunnelGatewayIPv4 = "10.0.0.1"
	cfg.PacketTunnelDNSIPv4 = "8.8.8.8"
	// Missing IPv6 is fine since we have IPv4
	cfg.PacketTunnelMTU = 1500
	require.NoError(t, validateConfig(cfg))
}

func TestBuildPsiphonConfig_NoApprovalCallback(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DataDirectory = t.TempDir()
	cfg.DisableRemoteServerListFetcher = true
	cfg.InproxyMode = "proxy"
	cfg.InproxyApprovalWebSocketURL = ""
	psiphonCfg, err := BuildPsiphonConfig(cfg)
	require.NoError(t, err)
	require.Nil(t, psiphonCfg.InproxyApproveClientConnection)

	cfg2 := DefaultConfig()
	cfg2.DataDirectory = t.TempDir()
	cfg2.DisableRemoteServerListFetcher = true
	cfg2.InproxyMode = "client"
	cfg2.InproxyApprovalWebSocketURL = "ws://example.com"
	psiphonCfg2, err := BuildPsiphonConfig(cfg2)
	require.NoError(t, err)
	require.Nil(t, psiphonCfg2.InproxyApproveClientConnection)
}


func TestBuildPsiphonConfig_ApprovalStrictness(t *testing.T) {
	t.Skip("not implemented yet")
}
