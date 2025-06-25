package providers

/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/flanksource/dns-sync/config"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
	awsProvider "sigs.k8s.io/external-dns/provider/aws"

	"sigs.k8s.io/external-dns/provider/akamai"
	"sigs.k8s.io/external-dns/provider/alibabacloud"
	"sigs.k8s.io/external-dns/provider/azure"
	"sigs.k8s.io/external-dns/provider/cloudflare"
	"sigs.k8s.io/external-dns/provider/coredns"
	"sigs.k8s.io/external-dns/provider/digitalocean"
	"sigs.k8s.io/external-dns/provider/exoscale"
	"sigs.k8s.io/external-dns/provider/godaddy"
	"sigs.k8s.io/external-dns/provider/google"
	"sigs.k8s.io/external-dns/provider/ibmcloud"
	"sigs.k8s.io/external-dns/provider/inmemory"
	"sigs.k8s.io/external-dns/provider/ns1"
	"sigs.k8s.io/external-dns/provider/oci"
	"sigs.k8s.io/external-dns/provider/ovh"
	"sigs.k8s.io/external-dns/provider/pdns"
	"sigs.k8s.io/external-dns/provider/pihole"
	"sigs.k8s.io/external-dns/provider/plural"
	"sigs.k8s.io/external-dns/provider/rfc2136"
	"sigs.k8s.io/external-dns/provider/tencentcloud"
	"sigs.k8s.io/external-dns/provider/transip"
)

func GetProvider(ctx context.Context, spec config.ProviderConfig, domainFilter endpoint.DomainFilter, zoneIDFilter provider.ZoneIDFilter, dryRun bool) (provider.Provider, error) {
	var p provider.Provider
	var err error

	// Helper function to create zone name filter
	zoneNameFilter := endpoint.NewDomainFilter([]string{})

	// AWS Provider
	if spec.AWS != nil {
		awsConfig := spec.AWS
		zoneTypeFilter := provider.NewZoneTypeFilter(awsConfig.ZoneType)
		zoneTagFilter := provider.NewZoneTagFilter(awsConfig.ZoneTagFilter)
		// zoneIdFilter := provider.NewZoneIDFilter(awsConf)

		// Create a simple AWS config map for now
		// This would need to be enhanced based on the actual AWS provider requirements
		configs := make(map[string]aws.Config)
		defaultConfig := aws.Config{}
		configs["default"] = defaultConfig

		clients := make(map[string]awsProvider.Route53API, len(configs))
		for profile, config := range configs {
			clients[profile] = route53.NewFromConfig(config)
		}

		p, err = awsProvider.NewAWSProvider(
			awsProvider.AWSConfig{
				DomainFilter: domainFilter,
				// ZoneIDFilter:          zoneIDFilter,
				ZoneTypeFilter:        zoneTypeFilter,
				ZoneTagFilter:         zoneTagFilter,
				ZoneMatchParent:       awsConfig.ZoneMatchParent,
				BatchChangeSize:       awsConfig.BatchChangeSize,
				BatchChangeSizeBytes:  awsConfig.BatchChangeSizeBytes,
				BatchChangeSizeValues: awsConfig.BatchChangeSizeValues,
				BatchChangeInterval:   awsConfig.BatchChangeInterval,
				EvaluateTargetHealth:  awsConfig.EvaluateTargetHealth,
				PreferCNAME:           awsConfig.PreferCNAME,
				DryRun:                dryRun,
				ZoneCacheDuration:     awsConfig.ZoneCacheDuration,
			},
			clients,
		)
		return p, err
	}

	// Azure Provider
	if spec.Azure != nil {
		azureConfig := spec.Azure
		p, err = azure.NewAzureProvider(
			azureConfig.ConfigFile,
			domainFilter,
			zoneNameFilter,
			zoneIDFilter,
			azureConfig.SubscriptionID,
			azureConfig.ResourceGroup,
			azureConfig.ClientID,
			azureConfig.ActiveDirectoryAuthorityHost,
			azureConfig.ZonesCacheDuration,
			dryRun,
		)
		return p, err
	}

	// Cloudflare Provider
	if spec.Cloudflare != nil {
		cfConfig := spec.Cloudflare
		p, err = cloudflare.NewCloudFlareProvider(
			domainFilter,
			zoneIDFilter,
			cfConfig.Proxied,
			dryRun,
			cfConfig.DNSRecordsPerPage,
			cfConfig.RegionKey,
			cloudflare.CustomHostnamesConfig{
				Enabled:              cfConfig.CustomHostnames,
				MinTLSVersion:        cfConfig.CustomHostnamesMinTLSVersion,
				CertificateAuthority: cfConfig.CustomHostnamesCertificateAuthority,
			},
		)
		return p, err
	}

	// Google Provider
	if spec.Google != nil {
		googleConfig := spec.Google
		p, err = google.NewGoogleProvider(
			ctx,
			googleConfig.Project,
			domainFilter,
			zoneIDFilter,
			googleConfig.BatchChangeSize,
			googleConfig.BatchChangeInterval,
			googleConfig.ZoneVisibility,
			dryRun,
		)
		return p, err
	}

	// Akamai Provider
	if spec.Akamai != nil {
		akamaiConfig := spec.Akamai
		p, err = akamai.NewAkamaiProvider(
			akamai.AkamaiConfig{
				DomainFilter:          domainFilter,
				ZoneIDFilter:          zoneIDFilter,
				ServiceConsumerDomain: akamaiConfig.ServiceConsumerDomain,
				ClientToken:           akamaiConfig.ClientToken,
				ClientSecret:          akamaiConfig.ClientSecret,
				AccessToken:           akamaiConfig.AccessToken,
				EdgercPath:            akamaiConfig.EdgercPath,
				EdgercSection:         akamaiConfig.EdgercSection,
				DryRun:                dryRun,
			}, nil)
		return p, err
	}

	// DigitalOcean Provider
	if spec.DigitalOcean != nil {
		doConfig := spec.DigitalOcean
		p, err = digitalocean.NewDigitalOceanProvider(
			ctx,
			domainFilter,
			dryRun,
			doConfig.APIPageSize,
		)
		return p, err
	}

	// OVH Provider
	if spec.OVH != nil {
		ovhConfig := spec.OVH
		p, err = ovh.NewOVHProvider(
			ctx,
			domainFilter,
			ovhConfig.Endpoint,
			ovhConfig.APIRateLimit,
			ovhConfig.EnableCNAMERelative,
			dryRun,
		)
		return p, err
	}

	// PowerDNS Provider
	if spec.PowerDNS != nil {
		pdnsConfig := spec.PowerDNS
		p, err = pdns.NewPDNSProvider(
			ctx,
			pdns.PDNSConfig{
				DomainFilter: domainFilter,
				DryRun:       dryRun,
				Server:       pdnsConfig.Server,
				ServerID:     pdnsConfig.ServerID,
				APIKey:       pdnsConfig.APIKey,
				TLSConfig: pdns.TLSConfig{
					SkipTLSVerify:         pdnsConfig.SkipTLSVerify,
					CAFilePath:            pdnsConfig.CA,
					ClientCertFilePath:    pdnsConfig.ClientCert,
					ClientCertKeyFilePath: pdnsConfig.ClientCertKey,
				},
			},
		)
		return p, err
	}

	// OCI Provider
	if spec.OCI != nil {
		ociConfig := spec.OCI
		var config *oci.OCIConfig

		if ociConfig.AuthInstancePrincipal {
			if len(ociConfig.CompartmentOCID) == 0 {
				return nil, fmt.Errorf("instance principal authentication requested, but no compartment OCID provided")
			}
			authConfig := oci.OCIAuthConfig{UseInstancePrincipal: true}
			config = &oci.OCIConfig{Auth: authConfig, CompartmentID: ociConfig.CompartmentOCID}
		} else {
			config, err = oci.LoadOCIConfig(ociConfig.ConfigFile)
			if err != nil {
				return nil, err
			}
		}
		config.ZoneCacheDuration = ociConfig.ZoneCacheDuration

		p, err = oci.NewOCIProvider(*config, domainFilter, zoneIDFilter, ociConfig.ZoneScope, dryRun)
		return p, err
	}

	// RFC2136 Provider
	if spec.RFC2136 != nil {
		rfc2136Config := spec.RFC2136
		tlsConfig := rfc2136.TLSConfig{
			UseTLS:                rfc2136Config.UseTLS,
			SkipTLSVerify:         rfc2136Config.SkipTLSVerify,
			CAFilePath:            rfc2136Config.CA,
			ClientCertFilePath:    rfc2136Config.ClientCert,
			ClientCertKeyFilePath: rfc2136Config.ClientCertKey,
		}
		p, err = rfc2136.NewRfc2136Provider(
			rfc2136Config.Host,
			rfc2136Config.Port,
			rfc2136Config.Zone,
			rfc2136Config.Insecure,
			rfc2136Config.TSIGKeyName,
			rfc2136Config.TSIGSecret,
			rfc2136Config.TSIGSecretAlg,
			rfc2136Config.TAXFR,
			domainFilter,
			dryRun,
			rfc2136Config.MinTTL,
			rfc2136Config.CreatePTR,
			rfc2136Config.GSSTSIG,
			rfc2136Config.KerberosUsername,
			rfc2136Config.KerberosPassword,
			rfc2136Config.KerberosRealm,
			rfc2136Config.BatchChangeSize,
			tlsConfig,
			rfc2136Config.LoadBalancingStrategy,
			nil,
		)
		return p, err
	}

	// NS1 Provider
	if spec.NS1 != nil {
		ns1Config := spec.NS1
		p, err = ns1.NewNS1Provider(
			ns1.NS1Config{
				DomainFilter:  domainFilter,
				ZoneIDFilter:  zoneIDFilter,
				NS1Endpoint:   ns1Config.Endpoint,
				NS1IgnoreSSL:  ns1Config.IgnoreSSL,
				DryRun:        dryRun,
				MinTTLSeconds: ns1Config.MinTTLSeconds,
			},
		)
		return p, err
	}

	// TransIP Provider
	if spec.TransIP != nil {
		transipConfig := spec.TransIP
		p, err = transip.NewTransIPProvider(
			transipConfig.AccountName,
			transipConfig.PrivateKeyFile,
			domainFilter,
			dryRun,
		)
		return p, err
	}

	// GoDaddy Provider
	if spec.GoDaddy != nil {
		godaddyConfig := spec.GoDaddy
		p, err = godaddy.NewGoDaddyProvider(
			ctx,
			domainFilter,
			godaddyConfig.TTL,
			godaddyConfig.APIKey,
			godaddyConfig.SecretKey,
			godaddyConfig.OTE,
			dryRun,
		)
		return p, err
	}

	// Exoscale Provider
	if spec.Exoscale != nil {
		exoscaleConfig := spec.Exoscale
		p, err = exoscale.NewExoscaleProvider(
			exoscaleConfig.APIEnvironment,
			exoscaleConfig.APIZone,
			exoscaleConfig.APIKey,
			exoscaleConfig.APISecret,
			dryRun,
			exoscale.ExoscaleWithDomain(domainFilter),
			exoscale.ExoscaleWithLogging(),
		)
		return p, err
	}

	// AlibabaCloud Provider
	if spec.AlibabaCloud != nil {
		alibabaConfig := spec.AlibabaCloud
		p, err = alibabacloud.NewAlibabaCloudProvider(
			alibabaConfig.ConfigFile,
			domainFilter,
			zoneIDFilter,
			alibabaConfig.ZoneType,
			dryRun,
		)
		return p, err
	}

	// TencentCloud Provider
	if spec.TencentCloud != nil {
		tencentConfig := spec.TencentCloud
		p, err = tencentcloud.NewTencentCloudProvider(
			domainFilter,
			zoneIDFilter,
			tencentConfig.ConfigFile,
			tencentConfig.ZoneType,
			dryRun,
		)
		return p, err
	}

	// IBMCloud Provider
	if spec.IBMCloud != nil {
		ibmConfig := spec.IBMCloud
		// Note: endpointsSource would need to be passed from outside or created here
		p, err = ibmcloud.NewIBMCloudProvider(
			ibmConfig.ConfigFile,
			domainFilter,
			zoneIDFilter,
			nil, // endpointsSource - this would need to be handled appropriately
			ibmConfig.Proxied,
			dryRun,
		)
		return p, err
	}

	// CoreDNS Provider
	if spec.CoreDNS != nil {
		corednsConfig := spec.CoreDNS
		p, err = coredns.NewCoreDNSProvider(domainFilter, corednsConfig.Prefix, dryRun)
		return p, err
	}

	// Pihole Provider
	if spec.Pihole != nil {
		piholeConfig := spec.Pihole
		p, err = pihole.NewPiholeProvider(
			pihole.PiholeConfig{
				Server:                piholeConfig.Server,
				Password:              piholeConfig.Password,
				TLSInsecureSkipVerify: piholeConfig.TLSInsecureSkipVerify,
				DomainFilter:          domainFilter,
				DryRun:                dryRun,
				APIVersion:            piholeConfig.APIVersion,
			},
		)
		return p, err
	}

	// Plural Provider
	if spec.Plural != nil {
		pluralConfig := spec.Plural
		p, err = plural.NewPluralProvider(pluralConfig.Cluster, pluralConfig.Provider)
		return p, err
	}

	// InMemory Provider
	if spec.InMemory != nil {
		inmemoryConfig := spec.InMemory
		p, err = inmemory.NewInMemoryProvider(
			inmemory.InMemoryInitZones(inmemoryConfig.Zones),
			inmemory.InMemoryWithDomain(domainFilter),
			inmemory.InMemoryWithLogging(),
		), nil
		return p, err
	}

	// File Provider
	if spec.File != nil {
		fileConfig := spec.File
		p = NewFileProvider(*fileConfig, domainFilter)
		return p, nil
	}

	return nil, fmt.Errorf("no valid provider configuration found")
}
