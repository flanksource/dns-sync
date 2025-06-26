package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
)

// Config represents the main configuration structure
type Config struct {
	// Log level (debug, info, warn, error)
	LogLevel string `yaml:"log_level" json:"log_level"`

	// Log format (text, json)
	LogFormat string `yaml:"log_format" json:"log_format"`

	// Metrics server address
	MetricsAddress string `yaml:"metrics_address" json:"metrics_address"`

	// Request timeout for external API calls
	RequestTimeout time.Duration `yaml:"request_timeout" json:"request_timeout"`
	// Target providers configuration
	Zones []*ZoneConfig `yaml:"zones" json:"zones"`

	// Synchronization settings
	Sync SyncConfig `yaml:"sync" json:"sync"`
}

type Spec Config

// SyncConfig contains synchronization settings
type SyncConfig struct {
	// How often to perform full synchronization
	Interval time.Duration `yaml:"interval" json:"interval"`

	// Enable DNS NOTIFY support for real-time updates
	EnableNotify bool `yaml:"enable_notify" json:"enable_notify"`

	// Port to listen for DNS NOTIFY messages
	NotifyPort int `yaml:"notify_port" json:"notify_port"`

	// Test mode - shows what would be changed without making changes
	DryRun bool `yaml:"dry_run" json:"dry_run"`

	// Remove records from target that don't exist in source
	DeleteOrphaned bool `yaml:"delete_orphaned" json:"delete_orphaned"`

	// Override TTL for all records (0 = use source TTL)
	RecordTTL uint32 `yaml:"record_ttl" json:"record_ttl"`
}

// SourceConfig defines the source DNS server configuration
type SourceConfig struct {
	ProviderConfig `yaml:",inline" json:",inline"`

	// Domain filtering configuration
	DomainFilter endpoint.DomainFilter `yaml:"domainFilter" json:"domainFilter"`
	RecordFilter provider.ZoneIDFilter `yaml:"recordFilter" json:"recordFilter"`
}

// TargetConfig defines target DNS provider configuration
type TargetConfig struct {
	ProviderConfig `yaml:",inline" json:",inline"`
}

// ProviderConfig contains provider-specific configurations
type ProviderConfig struct {
	AWS          *AWSProviderConfig          `yaml:"aws,omitempty" json:"aws,omitempty"`
	Azure        *AzureProviderConfig        `yaml:"azure,omitempty" json:"azure,omitempty"`
	Cloudflare   *CloudflareProviderConfig   `yaml:"cloudflare,omitempty" json:"cloudflare,omitempty"`
	Google       *GoogleProviderConfig       `yaml:"google,omitempty" json:"google,omitempty"`
	Akamai       *AkamaiProviderConfig       `yaml:"akamai,omitempty" json:"akamai,omitempty"`
	OCI          *OCIProviderConfig          `yaml:"oci,omitempty" json:"oci,omitempty"`
	OVH          *OVHProviderConfig          `yaml:"ovh,omitempty" json:"ovh,omitempty"`
	PowerDNS     *PowerDNSProviderConfig     `yaml:"powerdns,omitempty" json:"powerdns,omitempty"`
	NS1          *NS1ProviderConfig          `yaml:"ns1,omitempty" json:"ns1,omitempty"`
	DigitalOcean *DigitalOceanProviderConfig `yaml:"digitalocean,omitempty" json:"digitalocean,omitempty"`
	IBMCloud     *IBMCloudProviderConfig     `yaml:"ibmcloud,omitempty" json:"ibmcloud,omitempty"`
	GoDaddy      *GoDaddyProviderConfig      `yaml:"godaddy,omitempty" json:"godaddy,omitempty"`
	Exoscale     *ExoscaleProviderConfig     `yaml:"exoscale,omitempty" json:"exoscale,omitempty"`
	RFC2136      *RFC2136ProviderConfig      `yaml:"rfc2136,omitempty" json:"rfc2136,omitempty"`
	AlibabaCloud *AlibabaCloudProviderConfig `yaml:"alibabacloud,omitempty" json:"alibabacloud,omitempty"`
	TencentCloud *TencentCloudProviderConfig `yaml:"tencentcloud,omitempty" json:"tencentcloud,omitempty"`
	CloudFoundry *CloudFoundryProviderConfig `yaml:"cloudfoundry,omitempty" json:"cloudfoundry,omitempty"`
	CoreDNS      *CoreDNSProviderConfig      `yaml:"coredns,omitempty" json:"coredns,omitempty"`
	TransIP      *TransIPProviderConfig      `yaml:"transip,omitempty" json:"transip,omitempty"`
	Pihole       *PiholeProviderConfig       `yaml:"pihole,omitempty" json:"pihole,omitempty"`
	Plural       *PluralProviderConfig       `yaml:"plural,omitempty" json:"plural,omitempty"`
	Webhook      *WebhookProviderConfig      `yaml:"webhook,omitempty" json:"webhook,omitempty"`
	InMemory     *InMemoryProviderConfig     `yaml:"inmemory,omitempty" json:"inmemory,omitempty"`
	File         *FileProviderConfig         `yaml:"file,omitempty" json:"file,omitempty"`
}

func (p ProviderConfig) String() string {
	if p.AWS != nil {
		return "AWS"
	} else if p.Azure != nil {
		return fmt.Sprintf("Azure{sub=%s}", p.Azure.SubscriptionID)
	} else if p.DigitalOcean != nil {
		return "DigitalOcean{}"
	} else if p.IBMCloud != nil {
		return "IBMCloud"
	} else if p.GoDaddy != nil {
		return "GoDaddy"
	} else if p.Exoscale != nil {
		return "Exoscale"
	} else if p.RFC2136 != nil {
		return fmt.Sprintf("RFC2136{host=%s,tsig=%s}", p.RFC2136.Host, p.RFC2136.TSIGKeyName)
	} else if p.AlibabaCloud != nil {
		return "AlibabaCloud"
	} else if p.TencentCloud != nil {
		return "TencentCloud"
	} else if p.CloudFoundry != nil {
		return "CloudFoundry"
	} else if p.CoreDNS != nil {
		return "CoreDNS"
	} else if p.Cloudflare != nil {
		return "Cloudflare"
	} else if p.TransIP != nil {
		return "TransIP"
	} else if p.Pihole != nil {
		return "Pihole"
	} else if p.Plural != nil {
		return "Plural"
	} else if p.Webhook != nil {
		return "Webhook"
	} else if p.InMemory != nil {
		return "InMemory"
	} else if p.File != nil {
		return fmt.Sprintf("File{%s}", p.File.Path)
	}
	return "Unknown"
}

type FileProviderConfig struct {
	// Path to the file containing DNS records
	Path string `yaml:"path" json:"path"`
}

// RecordFilterConfig defines which records to sync
type RecordFilterConfig struct {
	// Only sync these record types
	IncludeTypes []string `yaml:"include_types" json:"include_types"`

	// Skip these record types
	ExcludeTypes []string `yaml:"exclude_types" json:"exclude_types"`

	// Only sync records matching these patterns
	IncludeNames []string `yaml:"include_names" json:"include_names"`

	// Skip records matching these patterns
	ExcludeNames []string `yaml:"exclude_names" json:"exclude_names"`
}

// ZoneConfig defines a zone to synchronize (legacy format)
type ZoneConfig struct {
	// Zone name
	Name string `yaml:"name" json:"name"`

	// Source DNS server configuration
	Source SourceConfig `yaml:"source" json:"source"`

	// Target providers
	Targets []TargetConfig `yaml:"targets" json:"targets"`

	// Record filtering configuration
	RecordFilter RecordFilterConfig `yaml:"record_filter" json:"record_filter"`
}

// Load reads and parses the configuration file
func Load(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	d, _ := yaml.Marshal(config)
	fmt.Println(string(d))

	// Set defaults
	if err := setDefaults(&config); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	// // Validate configuration
	// if err := validate(&config); err != nil {
	// 	return nil, fmt.Errorf("invalid configuration: %w", err)
	// }

	return &config, nil
}

// setDefaults applies default values to the configuration
func setDefaults(config *Config) error {
	// Core defaults
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.LogFormat == "" {
		config.LogFormat = "text"
	}
	if config.MetricsAddress == "" {
		config.MetricsAddress = ":7979"
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 30 * time.Second
	}

	// Sync defaults
	if config.Sync.Interval == 0 {
		config.Sync.Interval = 5 * time.Minute
	}
	if config.Sync.NotifyPort == 0 {
		config.Sync.NotifyPort = 5353
	}
	if config.Sync.RecordTTL == 0 {
		config.Sync.RecordTTL = 300
	}

	return nil
}
