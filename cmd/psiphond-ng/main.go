package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/dacineu/PsiphonNGLinux/internal/config"
	"github.com/dacineu/PsiphonNGLinux/internal/metrics"
	"github.com/dacineu/PsiphonNGLinux/internal/notification"
	"github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon"
	"github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon/common/errors"
)

// PsiphonController wraps the psiphon Controller
type PsiphonController struct {
	controller    *psiphon.Controller
	ngConfig      *config.Config
	psiphonConfig *psiphon.Config
	stopChan      chan struct{}
	cancelFunc    context.CancelFunc
	metricsServer *metrics.Server
}

// NewPsiphonController creates a new controller
func NewPsiphonController(ngConfig *config.Config) (*PsiphonController, error) {
	// Build psiphon config
	psiphonCfg, err := config.BuildPsiphonConfig(ngConfig)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// Create controller
	controller, err := psiphon.NewController(psiphonCfg)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &PsiphonController{
		controller:    controller,
		ngConfig:      ngConfig,
		psiphonConfig: psiphonCfg,
		stopChan:      make(chan struct{}),
	}, nil
}

// Start begins tunnel operation
func (pc *PsiphonController) Start() error {
	// Initialize data store
	if err := psiphon.OpenDataStore(pc.psiphonConfig); err != nil {
		return errors.Trace(err)
	}

	// Start tunnel loop with a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	pc.cancelFunc = cancel
	go func() {
		pc.controller.Run(ctx)
		close(pc.stopChan)
	}()

	return nil
}

// Stop gracefully shuts down the controller
func (pc *PsiphonController) Stop() {
	if pc.cancelFunc != nil {
		pc.cancelFunc()
	}
	psiphon.CloseDataStore()
}

// Wait blocks until controller stops
func (pc *PsiphonController) Wait() {
	<-pc.stopChan
}

// setupLogging configures logging based on config and returns a writer
// that receives Psiphon notices, along with the log file handle for closing.
func setupLogging(ngConfig *config.Config) (io.Writer, *os.File, error) {
	// Ensure log directory exists
	logDir := filepath.Dir(ngConfig.LogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, errors.Tracef("failed to create log directory: %s", err)
	}

	// Open log file
	logFile, err := os.OpenFile(ngConfig.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, errors.Tracef("failed to open log file: %s", err)
	}

	// Also log to stdout/stderr
	multiWriter := io.MultiWriter(logFile, os.Stdout)

	return multiWriter, logFile, nil
}

// Service template for systemd user unit.
// The placeholder %BIN% will be replaced with the actual binary path during installation.
const serviceTemplate = `[Unit]
Description=PsiphonNG Tunnel Daemon
After=network.target

[Service]
Type=simple
ExecStart=%s %s
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
`

// copyFile copies src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

// isTerminal checks if stdin is a terminal.
func isTerminal() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// installService installs PsiphonNG as a user systemd service.
func installService() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Define default paths
	defaultBinDir := filepath.Join(home, ".local/bin")
	configDir := filepath.Join(home, ".config/psiphond-ng")
	systemdDir := filepath.Join(home, ".config/systemd/user")
	dataDir := filepath.Join(home, ".local/var/lib/psiphon")
	logDir := filepath.Join(home, ".local/var/log/psiphon")

	// Determine bin directory with the following precedence:
	// 1. Command-line flag --bin-dir (if present)
	// 2. Environment variable PSIPHON_BIN_DIR
	// 3. Interactive prompt (if terminal)
	// 4. Default: ~/.local/bin
	binDir := defaultBinDir

	// Check for --bin-dir flag in arguments after "service"
	for i := 1; i < len(os.Args)-1; i++ {
		if os.Args[i] == "--bin-dir" {
			binDir = filepath.Clean(os.Args[i+1])
			break
		}
	}
	// If not set via flag, check environment variable
	if binDir == defaultBinDir {
		if envDir := os.Getenv("PSIPHON_BIN_DIR"); envDir != "" {
			binDir = filepath.Clean(envDir)
		}
	}
	// If still default and interactive, prompt
	if binDir == defaultBinDir && isTerminal() {
		fmt.Printf("Where should the binary be installed? [default: %s]: ", defaultBinDir)
		var answer string
		fmt.Scanln(&answer)
		if strings.TrimSpace(answer) != "" {
			expanded := os.ExpandEnv(answer)
			if strings.HasPrefix(expanded, "~/") || strings.HasPrefix(expanded, "~\\") {
				expanded = filepath.Join(home, expanded[2:])
			}
			binDir = filepath.Clean(expanded)
		}
	}

	// Create necessary directories
	for _, dir := range []string{binDir, configDir, systemdDir, dataDir, logDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Install binary: copy the running executable to binDir
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	destBin := filepath.Join(binDir, "psiphond-ng")
	// Check if destination binary already exists
	if _, err := os.Stat(destBin); err == nil {
		if isTerminal() {
			fmt.Printf("Binary already exists at: %s\n", destBin)
			fmt.Print("Choose action: [O]verwrite, [B]ackup then overwrite, [C]ancel? [O] ")
			var answer string
			fmt.Scanln(&answer)
			ans := strings.ToUpper(strings.TrimSpace(answer))
			switch ans {
			case "B":
				backupPath := destBin + ".old"
				os.Remove(backupPath) // ignore
				if err := os.Rename(destBin, backupPath); err != nil {
					return fmt.Errorf("failed to backup existing binary: %w", err)
				}
				log.Printf("Backed up existing binary to %s", backupPath)
			case "C":
				return fmt.Errorf("installation cancelled by user")
			case "O", "":
				// Just overwrite (default)
				log.Println("Overwriting existing binary.")
			}
		} else {
			// Non-interactive: default to overwrite with log
			log.Printf("Binary exists at %s; overwriting (non-interactive mode)", destBin)
		}
	}
	// Copy binary
	if err := copyFile(exePath, destBin); err != nil {
		return fmt.Errorf("failed to copy binary to %s: %w", destBin, err)
	}
	if err := os.Chmod(destBin, 0755); err != nil {
		return fmt.Errorf("failed to set binary permissions: %w", err)
	}

	// If the source and destination are different files (e.g., user installed from Downloads to ~/.local/bin)
	// and we're in an interactive terminal, ask if they want to delete the original.
	if exePath != destBin && isTerminal() {
		fmt.Printf("Binary installed to: %s\n", destBin)
		fmt.Printf("Original location: %s\n", exePath)
		fmt.Print("Delete original binary? [Y/n] ")
		var answer string
		fmt.Scanln(&answer)
		ans := strings.ToLower(strings.TrimSpace(answer))
		if ans == "" || ans[0] == 'y' {
			if err := os.Remove(exePath); err != nil {
				log.Printf("Warning: failed to delete original binary: %v", err)
			} else {
				log.Println("Original binary deleted.")
			}
		}
	}

	// Install configuration
	configPath := filepath.Join(configDir, "psiphond-ng.conf")
	// Backup existing config if present
	if _, err := os.Stat(configPath); err == nil {
		backupPath := configPath + ".bkp"
		os.Remove(backupPath) // ignore
		if err := os.Rename(configPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup config: %w", err)
		}
		log.Printf("Backed up existing config to %s", backupPath)
	}

	// Generate configuration using DefaultConfig
	ngConfig := config.DefaultConfig()
	// Override paths for user install
	ngConfig.DataDirectory = dataDir
	ngConfig.LogFile = filepath.Join(logDir, "psiphond-ng.log")
	// Set approval server integration (optional but convenient)
	ngConfig.InproxyApprovalWebSocketURL = "ws://localhost:8443/approval"
	ngConfig.InproxyApprovalTimeout = "10s"
	// Save configuration
	if err := config.SaveConfig(ngConfig, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	log.Printf("Created configuration at %s", configPath)

	// Install systemd user service
	servicePath := filepath.Join(systemdDir, "psiphond-ng.service")
	// Build service file content with actual paths
	serviceContent := fmt.Sprintf(serviceTemplate, destBin, configPath)
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}
	log.Printf("Created service file at %s", servicePath)

	// Reload systemd user daemon (best effort)
	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		log.Printf("Warning: failed to reload systemd daemon: %v", err)
		log.Println("You may need to run: systemctl --user daemon-reload")
	} else {
		// Enable and start the service
		if err := exec.Command("systemctl", "--user", "enable", "--now", "psiphond-ng.service").Run(); err != nil {
			log.Printf("Warning: failed to enable/start service: %v", err)
			log.Println("You may need to start it manually: systemctl --user start psiphond-ng")
		} else {
			log.Println("Service enabled and started successfully.")
		}
	}

	// Print summary
	fmt.Println("\n===========================================")
	fmt.Println("PsiphonNG installed as a user service!")
	fmt.Println("===========================================")
	fmt.Printf("Binary:       %s\n", destBin)
	fmt.Printf("Config:       %s\n", configPath)
	fmt.Printf("Service:      %s\n", servicePath)
	fmt.Printf("Data dir:     %s\n", dataDir)
	fmt.Printf("Log dir:      %s\n", logDir)
	fmt.Println("\nControl commands:")
	fmt.Println("  systemctl --user status psiphond-ng")
	fmt.Println("  journalctl --user-unit psiphond-ng -f")
	fmt.Println("  systemctl --user stop psiphond-ng")
	fmt.Println("  systemctl --user start psiphond-ng")
	fmt.Println("\nTo run in foreground (client mode):")
	fmt.Printf("  %s\n", configPath)
	fmt.Println("===========================================")

	return nil
}

// createDefaultConfig generates a fresh default configuration at the specified path.
func createDefaultConfig(configPath string) error {
	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Build default config with user-writable paths
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	ngConfig := config.DefaultConfig()

	// Set user-writable paths
	ngConfig.DataDirectory = filepath.Join(home, ".local/var/lib/psiphon")
	ngConfig.LogFile = filepath.Join(home, ".local/var/log/psiphon/psiphond-ng.log")

	// Enable DOIP approval integration by default
	ngConfig.InproxyApprovalWebSocketURL = "ws://localhost:8443/approval"
	ngConfig.InproxyApprovalTimeout = "10s"

	// Incorporate defaults from PsiphonLinux reference config for real network connectivity
	ngConfig.LocalSocksProxyPort = 1081
	ngConfig.LocalHttpProxyPort = 8081
	ngConfig.EgressRegion = "US"
	ngConfig.PropagationChannelId = "FFFFFFFFFFFFFFFF"
	ngConfig.SponsorId = "FFFFFFFFFFFFFFFF"

	// Enable remote server list fetching with PKI signature verification
	// Using production Psiphon server list endpoints
	ngConfig.DisableRemoteServerListFetcher = false
	ngConfig.RemoteServerListSignaturePublicKey = "MIICIDANBgkqhkiG9w0BAQEFAAOCAg0AMIICCAKCAgEAt7Ls+/39r+T6zNW7GiVpJfzq/xvL9SBH5rIFnk0RXYEYavax3WS6HOD35eTAqn8AniOwiH+DOkvgSKF2caqk/y1dfq47Pdymtwzp9ikpB1C5OfAysXzBiwVJlCdajBKvBZDerV1cMvRzCKvKwRmvDmHgphQQ7WfXIGbRbmmk6opMBh3roE42KcotLFtqp0RRwLtcBRNtCdsrVsjiI1Lqz/lH+T61sGjSjQ3CHMuZYSQJZo/KrvzgQXpkaCTdbObxHqb6/+i1qaVOfEsvjoiyzTxJADvSytVtcTjijhPEV6XskJVHE1Zgl+7rATr/pDQkw6DPCNBS1+Y6fy7GstZALQXwEDN/qhQI9kWkHijT8ns+i1vGg00Mk/6J75arLhqcodWsdeG/M/moWgqQAnlZAGVtJI1OgeF5fsPpXu4kctOfuZlGjVZXQNW34aOzm8r8S0eVZitPlbhcPiR4gT/aSMz/wd8lZlzZYsje/Jr8u/YtlwjjreZrGRmG8KMOzukV3lLmMppXFMvl4bxv6YFEmIuTsOhbLTwFgh7KYNjodLj/LsqRVfwz31PgWQFTEPICV7GCvgVlPRxnofqKSjgTWI4mxDhBpVcATvaoBl1L/6WLbFvBsoAUBItWwctO2xalKxF5szhGm8lccoc5MZr8kfE0uxMgsxz4er68iCID+rsCAQM="
	ngConfig.RemoteServerListURLs = []string{"https://s3.amazonaws.com/psiphon/web/mjr4-p23r-puwl/server_list_compressed"}
	ngConfig.ObfuscatedServerListRootURLs = []string{"https://s3.amazonaws.com/psiphon/web/mjr4-p23r-puwl/obfuscated_server_lists"}

	// Save configuration
	if err := config.SaveConfig(ngConfig, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

func main() {
	// Handle "service" command to install as systemd user service
	if len(os.Args) > 1 && os.Args[1] == "service" {
		if err := installService(); err != nil {
			log.Fatalf("Installation failed: %v", err)
		}
		return
	}

	// Determine config path
	var configPath string
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	} else {
		// Default: user config location
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".config/psiphond-ng/psiphond-ng.conf")
	}

	// Handle config file: if missing, create default; if present but invalid, backup and create new default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Configuration file not found at %s; creating default configuration...", configPath)
		if err := createDefaultConfig(configPath); err != nil {
			log.Fatalf("Failed to create default config: %v", err)
		}
	} else {
		// Try to load existing config
		_, err := config.LoadConfig(configPath)
		if err != nil {
			log.Printf("Warning: failed to load existing config: %v", err)
			log.Println("Configuration format may have changed. Backing up and creating new default config...")
			// Backup old config
			backupPath := configPath + ".bak_" + time.Now().Format("20060102-150405")
			if err := os.Rename(configPath, backupPath); err != nil {
				log.Printf("Failed to backup old config: %v", err)
				// Continue anyway, will overwrite
			} else {
				log.Printf("Old config backed up to: %s", backupPath)
			}
			// Create new default config
			if err := createDefaultConfig(configPath); err != nil {
				log.Fatalf("Failed to create new default config: %v", err)
			}
			log.Println("New default configuration created. Please review and adjust settings as needed.")
		}
	}

	// Load configuration
	log.Printf("Loading configuration from %s", configPath)
	ngConfig, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure required directories exist (DataDirectory and LogFile parent)
	dirs := []string{
		ngConfig.DataDirectory,
		filepath.Dir(ngConfig.LogFile),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Setup logging and notice handling
	noticeWriter, logFile, err := setupLogging(ngConfig)
	if err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}
	defer logFile.Close()

	// Initialize metrics if enabled
	var metricsInstance *metrics.Metrics
	var metricsServer *metrics.Server
	if ngConfig.MetricsEnabled {
		metricsInstance = metrics.NewMetrics()
		// Wrap the notice writer to collect metrics
		noticeWriter = metrics.NewNoticeCollector(metricsInstance, noticeWriter)
		// Start metrics HTTP server
		metricsServer = metrics.NewServer(metricsInstance, ngConfig.MetricsPort)
		metricsServer.Start()
		log.Printf("Metrics server started on %s", ngConfig.MetricsPort)
	}

	// Start approval notification server if enabled
	if ngConfig.ApprovalNotificationEnabled {
		if ngConfig.ApprovalNotificationPort == "" {
			ngConfig.ApprovalNotificationPort = ":9101"
		}
		if err := notification.Start(ngConfig.ApprovalNotificationPort); err != nil {
			log.Fatalf("Failed to start notification server: %v", err)
		}
		log.Printf("Approval notification server started on %s", ngConfig.ApprovalNotificationPort)
	}

	// Set Psiphon notice writer (after possibly wrapping with metrics collector)
	if err := psiphon.SetNoticeWriter(noticeWriter); err != nil {
		log.Fatalf("Failed to set notice writer: %v", err)
	}

	log.Println("Starting PsiphonNGLinux daemon")
	log.Printf("Version: %s", ngConfig.ClientVersion)
	log.Printf("Platform: %s", ngConfig.ClientPlatform)
	log.Printf("Data directory: %s", ngConfig.DataDirectory)
	log.Printf("Tunnel mode: %s", ngConfig.TunnelMode)

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Create controller
	controller, err := NewPsiphonController(ngConfig)
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}
	// Attach metrics server for graceful shutdown
	controller.metricsServer = metricsServer

	// Start config watcher for hot-reload
	configWatcher := config.NewConfigWatcher(configPath, ngConfig)
	if err := configWatcher.Start(); err != nil {
		log.Fatalf("Failed to start config watcher: %v", err)
	}
	log.Println("Configuration hot-reload enabled")

	// Start controller
	if err := controller.Start(); err != nil {
		log.Fatalf("Failed to start controller: %v", err)
	}

	log.Println("Controller started successfully")

	// Monitor for signals, controller exit, and config reloads
	go func() {
		defer notification.Stop()
		for {
			select {
			case sig := <-sigChan:
				log.Printf("Received signal %s, shutting down", sig)
				controller.Stop()
				if controller.metricsServer != nil {
					shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					controller.metricsServer.Stop(shutdownCtx)
				}
				notification.Stop()
				return

			case newConfig := <-configWatcher.ReloadChan():
				log.Println("Processing configuration reload...")
				// For now, restart the controller to apply new config
				// In future, we can implement selective parameter updates
				log.Println("Restarting controller to apply new configuration")
				controller.Stop()
				controller.Wait()

				// Create new controller with updated config
				newController, err := NewPsiphonController(newConfig)
				if err != nil {
					log.Printf("Failed to create new controller after config reload: %v", err)
					// Try to restart with old config
					oldController, _ := NewPsiphonController(ngConfig)
					if oldController != nil {
						_ = oldController.Start()
					}
					return
				}
				controller = newController
				controller.metricsServer = metricsServer

				if err := controller.Start(); err != nil {
					log.Fatalf("Failed to restart controller after config reload: %v", err)
				}
				log.Println("Controller restarted with new configuration")

				// Update reference for next reload
				ngConfig = newConfig

			case <-controller.stopChan:
				log.Println("Controller stopped")
				if controller.metricsServer != nil {
					shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					controller.metricsServer.Stop(shutdownCtx)
				}
				return
			}
		}
	}()

	// Wait for shutdown
	controller.Wait()
	log.Println("PsiphonNGLinux daemon exiting")
}
