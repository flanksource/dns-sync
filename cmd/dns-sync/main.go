package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/flanksource/dns-sync/config"
	"github.com/flanksource/dns-sync/sync"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	var configFile = flag.String("config", "config.yaml", "Configuration file path")
	var logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	var showVersion = flag.Bool("version", false, "Show version information")
	var dryRun = flag.Bool("dry-run", false, "Enable dry run mode (no changes made)")
	var once = flag.Bool("once", false, "Run synchronization once and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("dns-sync version %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	if dryRun != nil {
		cfg.Sync.DryRun = *dryRun
	}

	// Setup logging
	setupLogging(*logLevel)

	// Initialize synchronizer
	syncer := sync.NewSynchronizer(*cfg)

	// Start the synchronizer
	log.Println("Starting DNS synchronizer...")
	if once != nil && *once {
		if err, _ := syncer.Once(context.Background()); err != nil {
			log.Fatalf("Synchronizer failed: %v", err)
		}
	} else {
		// Setup graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle OS signals
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigCh
			log.Printf("Received signal %v, shutting down gracefully...", sig)
			cancel()
		}()
		if err := syncer.Start(ctx); err != nil {
			log.Fatalf("Synchronizer failed: %v", err)
		}
		log.Println("DNS synchronizer stopped")
	}

}

func setupLogging(level string) {
	// Configure logging based on level
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	switch level {
	case "debug":
		// Enable debug logging
	case "info":
		// Default info logging
	case "warn":
		// Warning and above
	case "error":
		// Error only
	default:
		log.Printf("Unknown log level: %s, using info", level)
	}
}
