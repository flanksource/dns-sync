package sync

import (
	"context"
	"log"
	"time"

	"github.com/flanksource/dns-sync/config"
	"github.com/flanksource/dns-sync/config/providers"
	"github.com/pkg/errors"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

// Synchronizer manages DNS zone synchronization
type Synchronizer struct {
	config config.Config
}

func NewSynchronizer(config config.Config) *Synchronizer {
	for _, zone := range config.Zones {
		if len(zone.RecordFilter.IncludeTypes) == 0 {
			zone.RecordFilter.IncludeTypes = []string{"A", "AAAA", "CNAME", "MX", "NS", "PTR", "SRV", "TXT"}
		}
	}
	return &Synchronizer{
		config: config,
	}
}

// Start starts the synchronizer
func (s *Synchronizer) Start(ctx context.Context) error {
	// // Start notify server if enabled
	// if s.config.Source.NotifyServer {
	// 	s.notifyServer = zone.NewNotifyServer(s.zoneClient, s.RFC.NotifyPort)

	// 	// Register notify handlers for each zone
	// 	for _, zoneConfig := range s.zones {
	// 		zoneName := zoneConfig.Name
	// 		s.notifyServer.RegisterHandler(zoneName, func(ctx context.Context, zone string) error {
	// 			return s.syncZone(ctx, s.findZoneConfig(zone))
	// 		})
	// 	}

	// 	// Start notify server in a goroutine
	// 	go func() {
	// 		if err := s.notifyServer.Start(ctx); err != nil {
	// 			log.Printf("Notify server error: %v", err)
	// 		}
	// 	}()
	// }

	// Start periodic sync
	ticker := time.NewTicker(s.config.Sync.Interval)
	defer ticker.Stop()

	// Perform initial sync
	if err, _ := s.syncAllZones(ctx); err != nil {
		log.Printf("Initial sync failed: %v", err)
	}

	// Main sync loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err, _ := s.syncAllZones(ctx); err != nil {
				log.Printf("Periodic sync failed: %v", err)
			}
		}
	}
}

func (s *Synchronizer) Once(ctx context.Context) (map[string]map[config.TargetConfig]*plan.Changes, error) {
	// Perform a single synchronization of all zones
	return s.syncAllZones(ctx)
}

// SetZones sets the zones to synchronize
func (s *Synchronizer) SetZones(zones []*config.ZoneConfig) {
	s.config.Zones = zones
}

// syncAllZones synchronizes all configured zones
func (s *Synchronizer) syncAllZones(ctx context.Context) (map[string]map[config.TargetConfig]*plan.Changes, error) {
	changes := make(map[string]map[config.TargetConfig]*plan.Changes)
	for _, zoneConfig := range s.config.Zones {
		if chg, err := s.syncZone(ctx, zoneConfig); err != nil {
			log.Printf("Failed to sync zone %s: %v", zoneConfig.Name, err)
			// Continue with other zones
		} else {
			changes[zoneConfig.Name] = chg
		}
	}
	return changes, nil
}

// syncZone synchronizes a single zone
func (s *Synchronizer) syncZone(ctx context.Context, zoneConfig *config.ZoneConfig) (map[config.TargetConfig]*plan.Changes, error) {
	log.Printf("Starting sync for zone: %s", zoneConfig.Name)

	changes := make(map[config.TargetConfig]*plan.Changes)

	source, _ := providers.GetProvider(ctx, zoneConfig.Source.ProviderConfig, zoneConfig.Source.DomainFilter, zoneConfig.Source.RecordFilter, s.config.Sync.DryRun)

	desired, err := s.listRecords(ctx, source, *zoneConfig)
	if err != nil {
		return nil, err
	}
	for _, targetConfig := range zoneConfig.Targets {
		target, err := providers.GetProvider(ctx, targetConfig.ProviderConfig, zoneConfig.Source.DomainFilter, zoneConfig.Source.RecordFilter, s.config.Sync.DryRun)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get target provider for %s", targetConfig.ProviderConfig.String())
		}
		current, err := s.listRecords(ctx, target, *zoneConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get current records from target provider for %s", targetConfig.ProviderConfig.String())
		}

		p := &plan.Plan{
			Desired:        desired,
			Current:        current,
			ManagedRecords: zoneConfig.RecordFilter.IncludeTypes,
			Policies: []plan.Policy{
				&plan.SyncPolicy{},
			},
		}

		p = Calculate(p)
		for _, i := range p.Changes.Create {
			log.Printf("+%s\n", i.String())
		}
		for _, i := range p.Changes.UpdateNew {
			log.Printf("~%s\n", i.String())
		}
		for _, i := range p.Changes.UpdateOld {
			log.Printf("~%s\n", i.String())
		}
		for _, i := range p.Changes.Delete {
			log.Printf("-%s\n", i.String())
		}
		log.Printf("Sync %s (%s): %d creates, %d updates, %d deletes", zoneConfig.Name, targetConfig.ProviderConfig.String(),
			len(p.Changes.Create), len(p.Changes.UpdateNew)+len(p.Changes.UpdateOld), len(p.Changes.Delete))

		if s.config.Sync.DryRun {
			log.Printf("Dry run enabled, skipping apply changes for target %s", targetConfig.ProviderConfig.String())
			continue
		} else if err := target.ApplyChanges(ctx, p.Changes); err != nil {
			return nil, errors.Wrapf(err, "failed to apply changes to target %s for zone %s", targetConfig.ProviderConfig.String(), zoneConfig.Name)
		}

		changes[targetConfig] = p.Changes
	}
	log.Printf("Completed sync for zone: %s", zoneConfig.Name)

	return changes, nil
}

func (s *Synchronizer) listRecords(ctx context.Context, p provider.Provider, zone config.ZoneConfig) ([]*endpoint.Endpoint, error) {

	records, err := p.Records(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch records from provider %s for zone %s", p, zone.Name)
	}

	filtered := s.filterRecords(records, zone.RecordFilter)

	log.Printf("Fetched %d records, filtered: %d from source provider for zone %s", len(records), len(filtered), zone.Name)

	transformed := s.transformRecords(filtered, zone)
	return transformed, nil

}

// filterRecords filters records based on the configured filter
func (s *Synchronizer) filterRecords(records []*endpoint.Endpoint, filter config.RecordFilterConfig) []*endpoint.Endpoint {
	var filtered []*endpoint.Endpoint

	for _, record := range records {
		// Check if record type should be included
		if len(filter.IncludeTypes) > 0 {
			included := false
			for _, includeType := range filter.IncludeTypes {
				if record.RecordType == includeType {
					included = true
					break
				}
			}
			if !included {
				continue
			}
		}

		// Check if record type should be excluded
		excluded := false
		for _, excludeType := range filter.ExcludeTypes {
			if record.RecordType == excludeType {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Check if record name should be included
		if len(filter.IncludeNames) > 0 {
			included := false
			for _, includeName := range filter.IncludeNames {
				if matchesPattern(record.DNSName, includeName) {
					included = true
					break
				}
			}
			if !included {
				continue
			}
		}

		// Check if record name should be excluded
		excluded = false
		for _, excludeName := range filter.ExcludeNames {
			if matchesPattern(record.DNSName, excludeName) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		filtered = append(filtered, record)
	}

	return filtered
}

// transformRecords applies transformations to records
func (s *Synchronizer) transformRecords(records []*endpoint.Endpoint, zoneConfig config.ZoneConfig) []*endpoint.Endpoint {
	var transformed []*endpoint.Endpoint

	for _, record := range records {
		// Apply TTL override if configured
		if s.config.Sync.RecordTTL > 0 {
			record.RecordTTL = endpoint.TTL(s.config.Sync.RecordTTL)
		}

		// Apply zone name transformations if needed
		// This could include mapping source zone names to target zone names

		transformed = append(transformed, record)
	}

	return transformed
}

// findZoneConfig finds a zone configuration by name
func (s *Synchronizer) findZoneConfig(zoneName string) config.ZoneConfig {
	for _, zoneConfig := range s.config.Zones {
		if zoneConfig.Name == zoneName {
			return *zoneConfig
		}
	}
	return config.ZoneConfig{}
}

// matchesPattern checks if a name matches a pattern (supports basic wildcards)
func matchesPattern(name, pattern string) bool {
	// Simple pattern matching - could be enhanced with regex or glob patterns
	if pattern == "*" {
		return true
	}

	// For now, just do exact matching
	return name == pattern
}

// ZoneStatus represents the sync status of a zone
type ZoneStatus struct {
	Name         string         `json:"name"`
	LastSync     time.Time      `json:"last_sync"`
	LastError    string         `json:"last_error,omitempty"`
	RecordCount  int            `json:"record_count"`
	TargetStatus []TargetStatus `json:"target_status"`
}

// TargetStatus represents the sync status for a specific target
type TargetStatus struct {
	Provider    string    `json:"provider"`
	ZoneID      string    `json:"zone_id"`
	LastSync    time.Time `json:"last_sync"`
	LastError   string    `json:"last_error,omitempty"`
	RecordCount int       `json:"record_count"`
}

// GetStatus returns the current sync status
func (s *Synchronizer) GetStatus() []ZoneStatus {
	// This would be implemented to track and return sync status
	// For now, return empty slice
	return []ZoneStatus{}
}
