package config

import (
	"regexp"
	"time"
)

// AWSProviderConfig contains AWS Route53 specific configuration
type AWSProviderConfig struct {
	// When using the AWS provider, filter for zones of this type (optional, options: public, private)
	ZoneType string `yaml:"zone_type" json:"zone_type"`

	// When using the AWS provider, filter for zones with these tags
	ZoneTagFilter []string `yaml:"zone_tag_filter" json:"zone_tag_filter"`

	// When using the AWS API, name of the profile to use
	Profiles []string `yaml:"profiles" json:"profiles"`

	// When using the AWS API, assume this IAM role. Useful for hosted zones in another AWS account.
	// Specify the full ARN, e.g. `arn:aws:iam::123455567:role/external-dns` (optional)
	AssumeRole string `yaml:"assume_role" json:"assume_role"`

	// When using the AWS API and assuming a role then specify this external ID (optional)
	AssumeRoleExternalID string `yaml:"assume_role_external_id" json:"assume_role_external_id" secure:"yes"`

	// When using the AWS provider, set the maximum number of changes that will be applied in each batch
	BatchChangeSize int `yaml:"batch_change_size" json:"batch_change_size"`

	// When using the AWS provider, set the maximum byte size that will be applied in each batch
	BatchChangeSizeBytes int `yaml:"batch_change_size_bytes" json:"batch_change_size_bytes"`

	// When using the AWS provider, set the maximum total record values that will be applied in each batch
	BatchChangeSizeValues int `yaml:"batch_change_size_values" json:"batch_change_size_values"`

	// When using the AWS provider, set the interval between batch changes
	BatchChangeInterval time.Duration `yaml:"batch_change_interval" json:"batch_change_interval"`

	// When using the AWS provider, set whether to evaluate the health of a DNS target (default: enabled)
	EvaluateTargetHealth bool `yaml:"evaluate_target_health" json:"evaluate_target_health"`

	// When using the AWS API, set the maximum number of retries before giving up
	APIRetries int `yaml:"api_retries" json:"api_retries"`

	// When using the AWS provider, prefer using CNAME instead of ALIAS (default: disabled)
	PreferCNAME bool `yaml:"prefer_cname" json:"prefer_cname"`

	// When using the AWS provider, set the zones list cache TTL (0s to disable)
	ZoneCacheDuration time.Duration `yaml:"zone_cache_duration" json:"zone_cache_duration"`

	// When using the AWS CloudMap provider, delete empty Services without endpoints (default: disabled)
	SDServiceCleanup bool `yaml:"sd_service_cleanup" json:"sd_service_cleanup"`

	// When using the AWS CloudMap provider, add tag to created services
	SDCreateTag map[string]string `yaml:"sd_create_tag" json:"sd_create_tag"`

	// Expand limit possible target by sub-domains (default: disabled)
	ZoneMatchParent bool `yaml:"zone_match_parent" json:"zone_match_parent"`

	// DynamoDB region for storing records
	DynamoDBRegion string `yaml:"dynamodb_region" json:"dynamodb_region"`

	// DynamoDB table name for storing records
	DynamoDBTable string `yaml:"dynamodb_table" json:"dynamodb_table"`
}

// AzureProviderConfig contains Azure DNS specific configuration
type AzureProviderConfig struct {
	// When using the Azure provider, specify the Azure configuration file (required when --provider=azure)
	ConfigFile string `yaml:"config_file" json:"config_file"`

	// When using the Azure provider, override the Azure resource group to use (optional)
	ResourceGroup string `yaml:"resource_group" json:"resource_group"`

	// When using the Azure provider, override the Azure subscription to use (optional)
	SubscriptionID string `yaml:"subscription_id" json:"subscription_id"`

	// When using the Azure provider, override the client id of user assigned identity in config file (optional)
	ClientID string `yaml:"client_id" json:"client_id"`

	// Azure Active Directory authority host
	ActiveDirectoryAuthorityHost string `yaml:"active_directory_authority_host" json:"active_directory_authority_host"`

	// When using the Azure provider, set the zones list cache TTL (0s to disable)
	ZonesCacheDuration time.Duration `yaml:"zones_cache_duration" json:"zones_cache_duration"`
}

// CloudflareProviderConfig contains Cloudflare specific configuration
type CloudflareProviderConfig struct {
	// When using the Cloudflare provider, specify if the proxy mode must be enabled (default: disabled)
	Proxied bool `yaml:"proxied" json:"proxied"`

	// When using the Cloudflare provider, specify if the Custom Hostnames feature will be used.
	// Requires "Cloudflare for SaaS" enabled (default: disabled)
	CustomHostnames bool `yaml:"custom_hostnames" json:"custom_hostnames"`

	// When using the Cloudflare provider with the Custom Hostnames, specify which Minimum TLS Version
	// will be used by default (default: 1.0, options: 1.0, 1.1, 1.2, 1.3)
	CustomHostnamesMinTLSVersion string `yaml:"custom_hostnames_min_tls_version" json:"custom_hostnames_min_tls_version"`

	// When using the Cloudflare provider with the Custom Hostnames, specify which Certificate Authority
	// will be used by default (default: google, options: google, ssl_com, lets_encrypt)
	CustomHostnamesCertificateAuthority string `yaml:"custom_hostnames_certificate_authority" json:"custom_hostnames_certificate_authority"`

	// When using the Cloudflare provider, specify how many DNS records listed per page, max possible 5,000 (default: 100)
	DNSRecordsPerPage int `yaml:"dns_records_per_page" json:"dns_records_per_page"`

	// When using the Cloudflare provider, specify the region (default: earth)
	RegionKey string `yaml:"region_key" json:"region_key"`
}

// GoogleProviderConfig contains Google Cloud DNS specific configuration
type GoogleProviderConfig struct {
	// When using the Google provider, current project is auto-detected, when running on GCP.
	// Specify other project with this. Must be specified when running outside GCP
	Project string `yaml:"project" json:"project"`

	// When using the Google provider, set the maximum number of changes that will be applied in each batch
	BatchChangeSize int `yaml:"batch_change_size" json:"batch_change_size"`

	// When using the Google provider, set the interval between batch changes
	BatchChangeInterval time.Duration `yaml:"batch_change_interval" json:"batch_change_interval"`

	// When using the Google provider, filter for zones with this visibility (optional, options: public, private)
	ZoneVisibility string `yaml:"zone_visibility" json:"zone_visibility"`
}

// AkamaiProviderConfig contains Akamai EdgeDNS specific configuration
type AkamaiProviderConfig struct {
	// When using the Akamai provider, specify the base URL (required when edgerc-path not specified)
	ServiceConsumerDomain string `yaml:"service_consumer_domain" json:"service_consumer_domain"`

	// When using the Akamai provider, specify the client token (required when edgerc-path not specified)
	ClientToken string `yaml:"client_token" json:"client_token"`

	// When using the Akamai provider, specify the client secret (required when edgerc-path not specified)
	ClientSecret string `yaml:"client_secret" json:"client_secret"`

	// When using the Akamai provider, specify the access token (required when edgerc-path not specified)
	AccessToken string `yaml:"access_token" json:"access_token"`

	// When using the Akamai provider, specify the .edgerc file path.
	// Path must be reachable from invocation environment
	EdgercPath string `yaml:"edgerc_path" json:"edgerc_path"`

	// When using the Akamai provider, specify the .edgerc file section (optional when edgerc-path is specified)
	EdgercSection string `yaml:"edgerc_section" json:"edgerc_section"`
}

// OCIProviderConfig contains Oracle Cloud Infrastructure DNS specific configuration
type OCIProviderConfig struct {
	// When using the OCI provider, specify the OCI configuration file (required when --provider=oci)
	ConfigFile string `yaml:"config_file" json:"config_file"`

	// When using the OCI provider, specify the OCID of the OCI compartment containing all managed zones and records.
	// Required when using OCI IAM instance principal authentication
	CompartmentOCID string `yaml:"compartment_ocid" json:"compartment_ocid"`

	// When using the OCI provider, specify whether OCI IAM instance principal authentication should be used
	// (instead of key-based auth via the OCI config file)
	AuthInstancePrincipal bool `yaml:"auth_instance_principal" json:"auth_instance_principal"`

	// When using OCI provider, filter for zones with this scope (optional, options: GLOBAL, PRIVATE).
	// Defaults to GLOBAL, setting to empty value will target both
	ZoneScope string `yaml:"zone_scope" json:"zone_scope"`

	// When using the OCI provider, set the zones list cache TTL (0s to disable)
	ZoneCacheDuration time.Duration `yaml:"zone_cache_duration" json:"zone_cache_duration"`
}

// OVHProviderConfig contains OVH DNS specific configuration
type OVHProviderConfig struct {
	// When using the OVH provider, specify the endpoint (default: ovh-eu)
	Endpoint string `yaml:"endpoint" json:"endpoint"`

	// When using the OVH provider, specify the API request rate limit, X operations by seconds (default: 20)
	APIRateLimit int `yaml:"api_rate_limit" json:"api_rate_limit"`

	// When using the OVH provider, specify if CNAME should be treated as relative on target without final dot (default: false)
	EnableCNAMERelative bool `yaml:"enable_cname_relative" json:"enable_cname_relative"`
}

// PowerDNSProviderConfig contains PowerDNS specific configuration
type PowerDNSProviderConfig struct {
	TLSConfig `yaml:",inline" json:",inline"`

	// When using the PowerDNS/PDNS provider, specify the URL to the pdns server (required)
	Server string `yaml:"server" json:"server"`

	// When using the PowerDNS/PDNS provider, specify the id of the server to retrieve.
	// Should be `localhost` except when the server is behind a proxy (default: localhost)
	ServerID string `yaml:"server_id" json:"server_id"`

	// When using the PowerDNS/PDNS provider, specify the API key to use to authorize requests (required)
	APIKey string `yaml:"api_key" json:"api_key" secure:"yes"`

	// When using the PowerDNS/PDNS provider, disable verification of any TLS certificates (default: false)
	SkipTLSVerify bool `yaml:"skip_tls_verify" json:"skip_tls_verify"`
}

// NS1ProviderConfig contains NS1 DNS specific configuration
type NS1ProviderConfig struct {
	// When using the NS1 provider, specify the URL of the API endpoint to target (default: https://api.nsone.net/v1/)
	Endpoint string `yaml:"endpoint" json:"endpoint"`

	// When using the NS1 provider, specify whether to verify the SSL certificate (default: false)
	IgnoreSSL bool `yaml:"ignore_ssl" json:"ignore_ssl"`

	// Minimal TTL (in seconds) for records. This value will be used if the provided TTL for a service/ingress is lower than this
	MinTTLSeconds int `yaml:"min_ttl_seconds" json:"min_ttl_seconds"`
}

// DigitalOceanProviderConfig contains DigitalOcean DNS specific configuration
type DigitalOceanProviderConfig struct {
	// Configure the page size used when querying the DigitalOcean API
	APIPageSize int `yaml:"api_page_size" json:"api_page_size"`
}

// IBMCloudProviderConfig contains IBM Cloud DNS specific configuration
type IBMCloudProviderConfig struct {
	// When using the IBM Cloud provider, specify the IBM Cloud configuration file (required)
	ConfigFile string `yaml:"config_file" json:"config_file"`

	// When using the IBM provider, specify if the proxy mode must be enabled (default: disabled)
	Proxied bool `yaml:"proxied" json:"proxied"`
}

// GoDaddyProviderConfig contains GoDaddy DNS specific configuration
type GoDaddyProviderConfig struct {
	// When using the GoDaddy provider, specify the API Key (required)
	APIKey string `yaml:"api_key" json:"api_key" secure:"yes"`

	// When using the GoDaddy provider, specify the API secret (required)
	SecretKey string `yaml:"secret_key" json:"secret_key" secure:"yes"`

	// TTL (in seconds) for records. This value will be used if the provided TTL for a service/ingress is not provided
	TTL int64 `yaml:"ttl" json:"ttl"`

	// When using the GoDaddy provider, use OTE api (optional, default: false)
	OTE bool `yaml:"ote" json:"ote"`
}

// ExoscaleProviderConfig contains Exoscale DNS specific configuration
type ExoscaleProviderConfig struct {
	// When using Exoscale provider, specify the API environment (optional)
	APIEnvironment string `yaml:"api_environment" json:"api_environment"`

	// When using Exoscale provider, specify the API Zone (optional)
	APIZone string `yaml:"api_zone" json:"api_zone"`

	// Provide your API Key for the Exoscale provider
	APIKey string `yaml:"api_key" json:"api_key" secure:"yes"`

	// Provide your API Secret for the Exoscale provider
	APISecret string `yaml:"api_secret" json:"api_secret" secure:"yes"`

	// When using Exoscale provider, specify the endpoint (optional)
	Endpoint string `yaml:"endpoint" json:"endpoint"`
}

// RFC2136ProviderConfig contains RFC2136 DNS specific configuration
type RFC2136ProviderConfig struct {
	// When using the RFC2136 provider, specify the host of the DNS server
	// (optionally specify multiple times when using load balancing strategy)
	Host []string `yaml:"host" json:"host"`

	NotifyServer bool `yaml:"notify_server" json:"notify_server"`
	NotifyPort   int  `yaml:"notify_port" json:"notify_port"`

	// When using the RFC2136 provider, specify the port of the DNS server
	Port int `yaml:"port" json:"port"`

	// When using the RFC2136 provider, specify zone entry of the DNS server to use (can be specified multiple times)
	Zone []string `yaml:"zone" json:"zone"`

	// When using the RFC2136 provider, disable secure communication (default: false)
	Insecure bool `yaml:"insecure" json:"insecure"`

	// When using the RFC2136 provider, enable GSS-TSIG authentication
	GSSTSIG bool `yaml:"gss_tsig" json:"gss_tsig"`

	// When using the RFC2136 provider, create PTR records automatically
	CreatePTR bool `yaml:"create_ptr" json:"create_ptr"`

	// When using the RFC2136 provider, specify the Kerberos realm
	KerberosRealm string `yaml:"kerberos_realm" json:"kerberos_realm"`

	// When using the RFC2136 provider, specify the Kerberos username
	KerberosUsername string `yaml:"kerberos_username" json:"kerberos_username"`

	// When using the RFC2136 provider, specify the Kerberos password
	KerberosPassword string `yaml:"kerberos_password" json:"kerberos_password" secure:"yes"`

	// When using the RFC2136 provider, specify the TSIG key name
	TSIGKeyName string `yaml:"tsig_key_name" json:"tsig_key_name"`

	// When using the RFC2136 provider, specify the TSIG secret
	TSIGSecret string `yaml:"tsig_secret" json:"tsig_secret" secure:"yes"`

	// When using the RFC2136 provider, specify the TSIG secret algorithm
	TSIGSecretAlg string `yaml:"tsig_secret_alg" json:"tsig_secret_alg"`

	// When using the RFC2136 provider, enable AXFR zone transfers
	TAXFR bool `yaml:"taxfr" json:"taxfr"`

	// When using the RFC2136 provider, specify the minimum TTL for records
	MinTTL time.Duration `yaml:"min_ttl" json:"min_ttl"`

	// When using the RFC2136 provider, specify the load balancing strategy (default: disabled)
	LoadBalancingStrategy string `yaml:"load_balancing_strategy" json:"load_balancing_strategy"`

	// When using the RFC2136 provider, set the maximum number of changes that will be applied in each batch
	BatchChangeSize int `yaml:"batch_change_size" json:"batch_change_size"`

	// When using the RFC2136 provider, use TLS for communication
	UseTLS bool `yaml:"use_tls" json:"use_tls"`

	// When using the RFC2136 provider, skip TLS certificate verification
	SkipTLSVerify bool `yaml:"skip_tls_verify" json:"skip_tls_verify"`
	TLSConfig     `yaml:",inline" json:",inline"`
}

// AlibabaCloudProviderConfig contains Alibaba Cloud DNS specific configuration
type AlibabaCloudProviderConfig struct {
	// When using the Alibaba Cloud provider, specify the Alibaba Cloud configuration file (required)
	ConfigFile string `yaml:"config_file" json:"config_file"`

	// When using the Alibaba Cloud provider, filter for zones of this type (optional, options: public, private)
	ZoneType string `yaml:"zone_type" json:"zone_type"`
}

// TencentCloudProviderConfig contains Tencent Cloud DNS specific configuration
type TencentCloudProviderConfig struct {
	// When using the Tencent Cloud provider, specify the Tencent Cloud configuration file (required)
	ConfigFile string `yaml:"config_file" json:"config_file"`

	// When using the Tencent Cloud provider, filter for zones with visibility (optional, options: public, private)
	ZoneType string `yaml:"zone_type" json:"zone_type"`
}

// CloudFoundryProviderConfig contains Cloud Foundry DNS specific configuration
type CloudFoundryProviderConfig struct {
	// The fully-qualified domain name of the cloud foundry instance you are targeting
	APIEndpoint string `yaml:"api_endpoint" json:"api_endpoint"`

	// The username to log into the cloud foundry API
	Username string `yaml:"username" json:"username"`

	// The password to log into the cloud foundry API
	Password string `yaml:"password" json:"password"`
}

// CoreDNSProviderConfig contains CoreDNS specific configuration
type CoreDNSProviderConfig struct {
	// When using the CoreDNS provider, specify the prefix name
	Prefix string `yaml:"prefix" json:"prefix"`
}

// TransIPProviderConfig contains TransIP DNS specific configuration
type TransIPProviderConfig struct {
	// When using the TransIP provider, specify the account name
	AccountName string `yaml:"account_name" json:"account_name"`

	// When using the TransIP provider, specify the private key file path
	PrivateKeyFile string `yaml:"private_key_file" json:"private_key_file"`
}

// PiholeProviderConfig contains Pi-hole DNS specific configuration
type PiholeProviderConfig struct {
	// When using the Pi-hole provider, specify the Pi-hole server URL
	Server string `yaml:"server" json:"server"`

	// When using the Pi-hole provider, specify the admin password
	Password string `yaml:"password" json:"password" secure:"yes"`

	// When using the Pi-hole provider, skip TLS certificate verification
	TLSInsecureSkipVerify bool `yaml:"tls_insecure_skip_verify" json:"tls_insecure_skip_verify"`

	// When using the Pi-hole provider, specify the API version (default: 5)
	APIVersion string `yaml:"api_version" json:"api_version"`
}

// PluralProviderConfig contains Plural DNS specific configuration
type PluralProviderConfig struct {
	// When using the Plural provider, specify the cluster name
	Cluster string `yaml:"cluster" json:"cluster"`

	// When using the Plural provider, specify the provider name
	Provider string `yaml:"provider" json:"provider"`
}

// WebhookProviderConfig contains webhook DNS provider specific configuration
type WebhookProviderConfig struct {
	// When using the webhook provider, specify the URL of the webhook endpoint
	URL string `yaml:"url" json:"url"`

	// When using the webhook provider, specify the read timeout
	ReadTimeout time.Duration `yaml:"read_timeout" json:"read_timeout"`

	// When using the webhook provider, specify the write timeout
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`

	// Enable webhook server mode
	Server bool `yaml:"server" json:"server"`
}

// InMemoryProviderConfig contains in-memory DNS provider specific configuration
type InMemoryProviderConfig struct {
	// Provide a list of pre-configured zones for the inmemory provider (optional)
	Zones []string `yaml:"zones" json:"zones"`
}

// TLSConfig contains TLS communication configuration
type TLSConfig struct {
	// When using TLS communication, the path to the certificate authority to verify server communications
	CA string `yaml:"ca" json:"ca"`

	// When using TLS communication, the path to the certificate to present as a client
	ClientCert string `yaml:"client_cert" json:"client_cert"`

	// When using TLS communication, the path to the certificate key to use with the client certificate
	ClientCertKey string `yaml:"client_cert_key" json:"client_cert_key"`
}

// DomainFilterConfig contains domain filtering configuration
type DomainFilterConfig struct {
	// Limit possible target zones by a domain suffix; specify multiple times for multiple domains (optional)
	DomainFilter []string `yaml:"domain_filter" json:"domain_filter"`

	// Exclude subdomains (optional)
	ExcludeDomains []string `yaml:"exclude_domains" json:"exclude_domains"`

	// Limit possible domains and target zones by a Regex filter; Overrides domain-filter (optional)
	RegexDomainFilter *regexp.Regexp `yaml:"regex_domain_filter" json:"regex_domain_filter"`

	// Regex filter that excludes domains and target zones matched by regex-domain-filter (optional)
	RegexDomainExclusion *regexp.Regexp `yaml:"regex_domain_exclusion" json:"regex_domain_exclusion"`

	// Filter target zones by zone domain; specify multiple times for multiple zones (optional)
	ZoneNameFilter []string `yaml:"zone_name_filter" json:"zone_name_filter"`

	// Filter target zones by hosted zone id; specify multiple times for multiple zones (optional)
	ZoneIDFilter []string `yaml:"zone_id_filter" json:"zone_id_filter"`
}
