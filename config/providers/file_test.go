package providers

import (
	"context"
	"os"
	"testing"

	"github.com/flanksource/dns-sync/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

const testDomain = "example.com"

func TestFileProvider_Records(t *testing.T) {
	// Create a temporary zone file
	zoneContent := `$ORIGIN example.com.
$TTL 3600

@               IN      SOA     ns1.example.com. admin.example.com. (
                                2023123101 ; serial
                                3600       ; refresh
                                1800       ; retry
                                604800     ; expire
                                86400      ; minimum
                        )

@               IN      NS      ns1.example.com.
@               IN      NS      ns2.example.com.

www             IN      A       192.168.1.10
api             IN      A       192.168.1.20
api             IN      A       192.168.1.21
mail            IN      AAAA    2001:db8::1
ftp             IN      CNAME   www.example.com.
_sip._tcp       IN      SRV     10 5 5060 sip.example.com.
@               IN      MX      10 mail.example.com.
@               IN      TXT     "v=spf1 include:_spf.google.com ~all"
test            300     IN      A       192.168.1.30
`

	tmpfile, err := os.CreateTemp("", "test-zone-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(zoneContent)
	require.NoError(t, err)
	tmpfile.Close()

	// Create file provider
	config := config.FileProviderConfig{
		Path: tmpfile.Name(),
	}

	domainFilter := endpoint.NewDomainFilter([]string{})
	provider := NewFileProvider(config, domainFilter)

	// Get records
	ctx := context.Background()
	records, err := provider.Records(ctx)
	require.NoError(t, err)

	// Verify records
	assert.NotEmpty(t, records)

	// Check specific records
	var foundA, foundAAAA, foundCNAME, foundMX, foundTXT, foundSRV bool

	for _, record := range records {
		switch record.RecordType {
		case "A":
			if record.DNSName == "www.example.com" {
				assert.Equal(t, []string{"192.168.1.10"}, []string(record.Targets))
				foundA = true
			} else if record.DNSName == "test.example.com" {
				assert.Equal(t, []string{"192.168.1.30"}, []string(record.Targets))
				assert.Equal(t, endpoint.TTL(300), record.RecordTTL)
			}
		case "AAAA":
			if record.DNSName == "mail.example.com" {
				assert.Equal(t, []string{"2001:db8::1"}, []string(record.Targets))
				foundAAAA = true
			}
		case "CNAME":
			if record.DNSName == "ftp.example.com" {
				assert.Equal(t, []string{"www.example.com"}, []string(record.Targets))
				foundCNAME = true
			}
		case "MX":
			if record.DNSName == testDomain {
				assert.Equal(t, []string{"10 mail.example.com"}, []string(record.Targets))
				foundMX = true
			}
		case "TXT":
			if record.DNSName == testDomain {
				assert.Equal(t, []string{"v=spf1 include:_spf.google.com ~all"}, []string(record.Targets))
				foundTXT = true
			}
		case "SRV":
			if record.DNSName == "_sip._tcp.example.com" {
				assert.Equal(t, []string{"10 5 5060 sip.example.com"}, []string(record.Targets))
				foundSRV = true
			}
		}
	}

	assert.True(t, foundA, "Should find A record")
	assert.True(t, foundAAAA, "Should find AAAA record")
	assert.True(t, foundCNAME, "Should find CNAME record")
	assert.True(t, foundMX, "Should find MX record")
	assert.True(t, foundTXT, "Should find TXT record")
	assert.True(t, foundSRV, "Should find SRV record")
}

func TestFileProvider_MultipleMXRecords(t *testing.T) {
	// Create a temporary zone file with multiple MX records
	zoneContent := `$ORIGIN example.com.
$TTL 3600

@               IN      SOA     ns1.example.com. admin.example.com. (
                                2023123101 ; serial
                                3600       ; refresh
                                1800       ; retry
                                604800     ; expire
                                86400      ; minimum
                        )

@               IN      NS      ns1.example.com.
@               IN      MX      10 mail1.example.com.
@               IN      MX      20 mail2.example.com.
@               IN      MX      30 mail3.example.com.
`

	tmpfile, err := os.CreateTemp("", "test-mx-zone-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(zoneContent)
	require.NoError(t, err)
	tmpfile.Close()

	// Create file provider
	config := config.FileProviderConfig{
		Path: tmpfile.Name(),
	}

	domainFilter := endpoint.NewDomainFilter([]string{})
	provider := NewFileProvider(config, domainFilter)

	// Get records
	ctx := context.Background()
	records, err := provider.Records(ctx)
	require.NoError(t, err)

	// Find MX records
	var mxRecords []*endpoint.Endpoint
	for _, record := range records {
		if record.RecordType == "MX" && record.DNSName == testDomain {
			mxRecords = append(mxRecords, record)
		}
	}

	// Should have exactly 3 MX records
	assert.Len(t, mxRecords, 3, "Should have 3 MX records")

	// Verify the MX record priorities and targets
	expectedMX := map[string]bool{
		"10 mail1.example.com": false,
		"20 mail2.example.com": false,
		"30 mail3.example.com": false,
	}

	for _, mx := range mxRecords {
		assert.Len(t, mx.Targets, 1, "MX record should have exactly one target")
		if len(mx.Targets) > 0 {
			target := mx.Targets[0]
			if _, exists := expectedMX[target]; exists {
				expectedMX[target] = true
			} else {
				t.Errorf("Unexpected MX target: %s", target)
			}
		}
	}

	// Verify all expected MX records were found
	for target, found := range expectedMX {
		assert.True(t, found, "Expected MX record not found: %s", target)
	}
}

func TestFileProvider_ApplyChanges_MultipleMXRecords(t *testing.T) {
	// Create a temporary zone file with multiple MX records
	zoneContent := `$ORIGIN example.com.
$TTL 3600

@               IN      SOA     ns1.example.com. admin.example.com. (
                                2023123101 ; serial
                                3600       ; refresh
                                1800       ; retry
                                604800     ; expire
                                86400      ; minimum
                        )

@               IN      NS      ns1.example.com.
@               IN      MX      10 mail1.example.com.
@               IN      MX      20 mail2.example.com.
`

	tmpfile, err := os.CreateTemp("", "test-mx-changes-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(zoneContent)
	require.NoError(t, err)
	tmpfile.Close()

	// Create file provider
	config := config.FileProviderConfig{
		Path: tmpfile.Name(),
	}

	domainFilter := endpoint.NewDomainFilter([]string{})
	provider := NewFileProvider(config, domainFilter)

	ctx := context.Background()

	// Test adding a new MX record
	newMXRecord := &endpoint.Endpoint{
		DNSName:    "example.com",
		RecordType: "MX",
		Targets:    []string{"30 mail3.example.com"},
		RecordTTL:  endpoint.TTL(3600),
	}

	changes := &plan.Changes{
		Create: []*endpoint.Endpoint{newMXRecord},
	}

	err = provider.ApplyChanges(ctx, changes)
	require.NoError(t, err)

	// Verify the new MX record was added
	records, err := provider.Records(ctx)
	require.NoError(t, err)

	var mxRecords []*endpoint.Endpoint
	for _, record := range records {
		if record.RecordType == "MX" && record.DNSName == "example.com" {
			mxRecords = append(mxRecords, record)
		}
	}

	assert.Len(t, mxRecords, 3, "Should have 3 MX records after adding one")

	// Test deleting a specific MX record (should only delete the one with priority 20)
	deleteMXRecord := &endpoint.Endpoint{
		DNSName:    "example.com",
		RecordType: "MX",
		Targets:    []string{"20 mail2.example.com"},
		RecordTTL:  endpoint.TTL(3600),
	}

	changes = &plan.Changes{
		Delete: []*endpoint.Endpoint{deleteMXRecord},
	}

	err = provider.ApplyChanges(ctx, changes)
	require.NoError(t, err)

	// Verify only the specific MX record was deleted
	records, err = provider.Records(ctx)
	require.NoError(t, err)

	mxRecords = []*endpoint.Endpoint{}
	for _, record := range records {
		if record.RecordType == "MX" && record.DNSName == "example.com" {
			mxRecords = append(mxRecords, record)
		}
	}

	assert.Len(t, mxRecords, 2, "Should have 2 MX records after deleting one")

	// Verify the remaining records are correct
	expectedMX := map[string]bool{
		"10 mail1.example.com": false,
		"30 mail3.example.com": false,
	}

	for _, mx := range mxRecords {
		if len(mx.Targets) > 0 {
			target := mx.Targets[0]
			if _, exists := expectedMX[target]; exists {
				expectedMX[target] = true
			} else {
				t.Errorf("Unexpected MX target after deletion: %s", target)
			}
		}
	}

	// Verify all expected MX records are still present
	for target, found := range expectedMX {
		assert.True(t, found, "Expected MX record missing after deletion: %s", target)
	}
}
