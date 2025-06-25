package sync

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/flanksource/dns-sync/config"

	"github.com/stretchr/testify/assert"
)

// go:embed ../../fixtures/zones.bind
var sampleZone string

func TestNewSynchronizer(t *testing.T) {
	target, _ := os.CreateTemp("", "target.bind")
	source, _ := os.CreateTemp("", "zones.bind")
	fmt.Println(sampleZone)
	_ = os.WriteFile(source.Name(), []byte(sampleZone), 0644)
	cfg := &config.Config{
		Sync: config.SyncConfig{
			Interval:     time.Second * 30,
			EnableNotify: true,
			NotifyPort:   5354,
		},
		Zones: []*config.ZoneConfig{
			{
				Name: "example.com",

				Source: config.SourceConfig{
					ProviderConfig: config.ProviderConfig{
						File: &config.FileProviderConfig{
							Path: source.Name(),
						},
					},
				},
				Targets: []config.TargetConfig{
					config.TargetConfig{
						ProviderConfig: config.ProviderConfig{
							File: &config.FileProviderConfig{
								Path: target.Name(),
							},
						},
					},
				},
			},
		},
	}

	test(t, *cfg, 14, 0, 0)
	test(t, *cfg, 0, 0, 0)

}

func test(t *testing.T, cfg config.Config, created, updated, deleted int) {
	s := NewSynchronizer(cfg)
	changes, err := s.Once(context.Background())
	assert.NoError(t, err)
	change := changes[cfg.Zones[0].Name][cfg.Zones[0].Targets[0]]
	assert.Equal(t, created, len(change.Create))
	assert.Equal(t, updated, len(change.UpdateNew)+len(change.UpdateOld))
	assert.Equal(t, deleted, len(change.Delete))
}
