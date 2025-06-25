package providers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/dns-sync/config"
	"github.com/miekg/dns"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

// fileProvider implements the external-dns Provider interface for reading DNS records from BIND zone files
type fileProvider struct {
	config       config.FileProviderConfig
	domainFilter endpoint.DomainFilter
}

// NewFileProvider creates a new file-based DNS provider
func NewFileProvider(config config.FileProviderConfig, domainFilter endpoint.DomainFilter) provider.Provider {
	return &fileProvider{
		config:       config,
		domainFilter: domainFilter,
	}
}

// Records retrieves all DNS records from the zone file
func (f *fileProvider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	file, err := os.Open(f.config.Path)
	if err != nil {

		return nil, fmt.Errorf("failed to open zone file %s: %w", f.config.Path, err)
	}
	defer file.Close()

	var endpoints []*endpoint.Endpoint

	// Create a new zone parser
	zp := dns.NewZoneParser(file, "", "")

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		endpoint, err := f.convertRRToEndpoint(rr)
		if err != nil {
			return nil, fmt.Errorf("error converting RR to endpoint: %w", err)
		}

		if endpoint != nil {
			endpoints = append(endpoints, endpoint)
		}
	}

	if err := zp.Err(); err != nil {
		return nil, fmt.Errorf("error parsing zone file: %w", err)
	}

	return endpoints, nil
}

// ApplyChanges applies DNS record changes by updating the zone file
func (f *fileProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	if len(changes.Create) == 0 && len(changes.UpdateNew) == 0 && len(changes.Delete) == 0 {
		return nil // No changes to apply
	}

	// Read the current zone file
	currentRecords, err := f.parseZoneFile()
	if err != nil {
		return fmt.Errorf("failed to parse current zone file: %w", err)
	}

	// Apply deletions
	for _, endpoint := range changes.Delete {
		f.deleteRecord(&currentRecords, endpoint)
	}

	// Apply updates (delete old, add new)
	for _, endpoint := range changes.UpdateOld {
		f.deleteRecord(&currentRecords, endpoint)
	}
	for _, endpoint := range changes.UpdateNew {
		f.addRecord(&currentRecords, endpoint)
	}

	// Apply creations
	for _, endpoint := range changes.Create {
		f.addRecord(&currentRecords, endpoint)
	}

	// Write the updated zone file
	return f.writeZoneFile(currentRecords)
}

// AdjustEndpoints canonicalizes endpoints (no adjustments needed for file provider)
func (f *fileProvider) AdjustEndpoints(endpoints []*endpoint.Endpoint) ([]*endpoint.Endpoint, error) {
	return endpoints, nil
}

// GetDomainFilter returns the domain filter for this provider
func (f *fileProvider) GetDomainFilter() endpoint.DomainFilterInterface {
	return f.domainFilter
}

// convertRRToEndpoint converts a DNS resource record to an external-dns endpoint
func (f *fileProvider) convertRRToEndpoint(rr dns.RR) (*endpoint.Endpoint, error) {
	header := rr.Header()

	// Skip SOA records as they're not typically managed by external-dns
	if header.Rrtype == dns.TypeSOA {
		return nil, nil
	}

	// Get the DNS name and remove trailing dot
	dnsName := strings.TrimSuffix(header.Name, ".")

	// Get record type as string
	recordType := dns.TypeToString[header.Rrtype]

	// Extract target based on record type
	var targets []string

	switch rr := rr.(type) {

	case *dns.A:
		targets = []string{rr.A.String()}
	case *dns.AAAA:
		targets = []string{rr.AAAA.String()}
	case *dns.CNAME:
		targets = []string{strings.TrimSuffix(rr.Target, ".")}
	case *dns.MX:
		targets = []string{fmt.Sprintf("%d %s", rr.Preference, strings.TrimSuffix(rr.Mx, "."))}
	case *dns.TXT:
		// Join all TXT strings
		targets = []string{strings.Join(rr.Txt, "")}
	case *dns.SRV:
		targets = []string{fmt.Sprintf("%d %d %d %s", rr.Priority, rr.Weight, rr.Port, strings.TrimSuffix(rr.Target, "."))}
	case *dns.NS:
		targets = []string{strings.TrimSuffix(rr.Ns, ".")}
	case *dns.PTR:
		targets = []string{strings.TrimSuffix(rr.Ptr, ".")}
	default:
		// For other record types, use the string representation
		targets = []string{strings.TrimSpace(strings.TrimPrefix(rr.String(), header.String()))}
	}

	endpoint := &endpoint.Endpoint{
		DNSName:    dnsName,
		RecordType: recordType,
		Targets:    targets,
		RecordTTL:  endpoint.TTL(header.Ttl),
	}

	return endpoint, nil
}

// parseZoneFile reads and parses the entire zone file into a slice of DNS resource records
func (f *fileProvider) parseZoneFile() ([]dns.RR, error) {
	file, err := os.Open(f.config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open zone file %s: %w", f.config.Path, err)
	}
	defer file.Close()

	var records []dns.RR
	zp := dns.NewZoneParser(file, "", "")

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		records = append(records, rr)
	}

	if err := zp.Err(); err != nil {
		return nil, fmt.Errorf("error parsing zone file: %w", err)
	}

	return records, nil
}

// deleteRecord removes matching records from the records slice
func (f *fileProvider) deleteRecord(records *[]dns.RR, endpoint *endpoint.Endpoint) {
	targetRRs := f.endpointToRRs(endpoint)
	if len(targetRRs) == 0 {
		return
	}

	// Filter out matching records
	var filtered []dns.RR
	for _, rr := range *records {
		shouldKeep := true
		for _, targetRR := range targetRRs {
			if f.recordsMatch(rr, targetRR) {
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			filtered = append(filtered, rr)
		}
	}
	*records = filtered
}

// addRecord adds new records to the records slice (one per target)
func (f *fileProvider) addRecord(records *[]dns.RR, endpoint *endpoint.Endpoint) {
	rrs := f.endpointToRRs(endpoint)
	*records = append(*records, rrs...)
}

// endpointToRRs converts an external-dns endpoint to DNS resource records (one per target)
func (f *fileProvider) endpointToRRs(endpoint *endpoint.Endpoint) []dns.RR {
	var rrs []dns.RR

	// Ensure DNS name has trailing dot for DNS library
	dnsName := endpoint.DNSName
	if !strings.HasSuffix(dnsName, ".") {
		dnsName += "."
	}

	ttl := uint32(endpoint.RecordTTL)
	if ttl == 0 {
		ttl = 300 // Default TTL
	}

	// Create header template
	header := dns.RR_Header{
		Name:   dnsName,
		Rrtype: dns.StringToType[endpoint.RecordType],
		Class:  dns.ClassINET,
		Ttl:    ttl,
	}

	// Create specific record type based on endpoint type, one per target
	for _, target := range endpoint.Targets {
		switch endpoint.RecordType {
		case "A":
			rrs = append(rrs, &dns.A{
				Hdr: header,
				A:   net.ParseIP(target),
			})
		case "AAAA":
			rrs = append(rrs, &dns.AAAA{
				Hdr:  header,
				AAAA: net.ParseIP(target),
			})
		case "CNAME":
			targetName := target
			if !strings.HasSuffix(targetName, ".") {
				targetName += "."
			}
			rrs = append(rrs, &dns.CNAME{
				Hdr:    header,
				Target: targetName,
			})
		case "MX":
			parts := strings.Fields(target)
			if len(parts) >= 2 {
				preference := uint16(0)
				if pref, err := strconv.ParseUint(parts[0], 10, 16); err == nil {
					preference = uint16(pref)
				}
				mx := parts[1]
				if !strings.HasSuffix(mx, ".") {
					mx += "."
				}
				rrs = append(rrs, &dns.MX{
					Hdr:        header,
					Preference: preference,
					Mx:         mx,
				})
			}
		case "TXT":
			rrs = append(rrs, &dns.TXT{
				Hdr: header,
				Txt: []string{target},
			})
		case "SRV":
			parts := strings.Fields(target)
			if len(parts) >= 4 {
				priority, _ := strconv.ParseUint(parts[0], 10, 16)
				weight, _ := strconv.ParseUint(parts[1], 10, 16)
				port, _ := strconv.ParseUint(parts[2], 10, 16)
				targetName := parts[3]
				if !strings.HasSuffix(targetName, ".") {
					targetName += "."
				}
				rrs = append(rrs, &dns.SRV{
					Hdr:      header,
					Priority: uint16(priority),
					Weight:   uint16(weight),
					Port:     uint16(port),
					Target:   targetName,
				})
			}
		case "NS":
			ns := target
			if !strings.HasSuffix(ns, ".") {
				ns += "."
			}
			rrs = append(rrs, &dns.NS{
				Hdr: header,
				Ns:  ns,
			})
		case "PTR":
			ptr := target
			if !strings.HasSuffix(ptr, ".") {
				ptr += "."
			}
			rrs = append(rrs, &dns.PTR{
				Hdr: header,
				Ptr: ptr,
			})
		}
	}

	return rrs
}

// recordsMatch compares two DNS resource records for equality
func (f *fileProvider) recordsMatch(rr1, rr2 dns.RR) bool {
	if rr1 == nil || rr2 == nil {
		return false
	}

	h1, h2 := rr1.Header(), rr2.Header()

	// Compare headers (name, type, class)
	if h1.Name != h2.Name || h1.Rrtype != h2.Rrtype || h1.Class != h2.Class {
		return false
	}

	// For specific record types, compare the actual data fields
	switch rr1.(type) {
	case *dns.MX:
		mx1, mx2 := rr1.(*dns.MX), rr2.(*dns.MX)
		return mx1.Preference == mx2.Preference && mx1.Mx == mx2.Mx
	case *dns.SRV:
		srv1, srv2 := rr1.(*dns.SRV), rr2.(*dns.SRV)
		return srv1.Priority == srv2.Priority && srv1.Weight == srv2.Weight &&
			srv1.Port == srv2.Port && srv1.Target == srv2.Target
	case *dns.A:
		a1, a2 := rr1.(*dns.A), rr2.(*dns.A)
		return a1.A.Equal(a2.A)
	case *dns.AAAA:
		aaaa1, aaaa2 := rr1.(*dns.AAAA), rr2.(*dns.AAAA)
		return aaaa1.AAAA.Equal(aaaa2.AAAA)
	case *dns.CNAME:
		cname1, cname2 := rr1.(*dns.CNAME), rr2.(*dns.CNAME)
		return cname1.Target == cname2.Target
	case *dns.TXT:
		txt1, txt2 := rr1.(*dns.TXT), rr2.(*dns.TXT)
		if len(txt1.Txt) != len(txt2.Txt) {
			return false
		}
		for i, txt := range txt1.Txt {
			if txt != txt2.Txt[i] {
				return false
			}
		}
		return true
	case *dns.NS:
		ns1, ns2 := rr1.(*dns.NS), rr2.(*dns.NS)
		return ns1.Ns == ns2.Ns
	case *dns.PTR:
		ptr1, ptr2 := rr1.(*dns.PTR), rr2.(*dns.PTR)
		return ptr1.Ptr == ptr2.Ptr
	default:
		// For other record types, fall back to string comparison
		return strings.TrimSpace(strings.TrimPrefix(rr1.String(), h1.String())) ==
			strings.TrimSpace(strings.TrimPrefix(rr2.String(), h2.String()))
	}
}

// writeZoneFile writes the DNS records back to the zone file
func (f *fileProvider) writeZoneFile(records []dns.RR) error {
	// Create a backup of the original file
	backupPath := f.config.Path + ".backup." + time.Now().Format("20060102-150405")
	if err := f.copyFile(f.config.Path, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Write the new zone file
	file, err := os.Create(f.config.Path)
	if err != nil {
		return fmt.Errorf("failed to create zone file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write zone file header
	fmt.Fprintf(writer, "; Zone file updated by dns-sync at %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(writer, "; Backup saved as: %s\n\n", filepath.Base(backupPath))

	// Group records by type for better organization
	recordsByType := make(map[uint16][]dns.RR)
	for _, rr := range records {
		rrtype := rr.Header().Rrtype
		recordsByType[rrtype] = append(recordsByType[rrtype], rr)
	}

	// Write records in a logical order
	typeOrder := []uint16{
		dns.TypeSOA,
		dns.TypeNS,
		dns.TypeA,
		dns.TypeAAAA,
		dns.TypeCNAME,
		dns.TypeMX,
		dns.TypeTXT,
		dns.TypeSRV,
		dns.TypePTR,
	}

	// Write records in order
	for _, rrtype := range typeOrder {
		if records, exists := recordsByType[rrtype]; exists {
			if rrtype != dns.TypeSOA {
				fmt.Fprintf(writer, "; %s records\n", dns.TypeToString[rrtype])
			}
			for _, rr := range records {
				fmt.Fprintf(writer, "%s\n", rr.String())
			}
			fmt.Fprintf(writer, "\n")
		}
	}

	// Write any remaining record types not in the standard order
	for rrtype, records := range recordsByType {
		found := false
		for _, standardType := range typeOrder {
			if rrtype == standardType {
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(writer, "; %s records\n", dns.TypeToString[rrtype])
			for _, rr := range records {
				fmt.Fprintf(writer, "%s\n", rr.String())
			}
			fmt.Fprintf(writer, "\n")
		}
	}

	return nil
}

// copyFile creates a backup copy of a file
func (f *fileProvider) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
