package processing

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"farm-sensor-platform/internal/model"
)

type Report struct {
	GeneratedAt time.Time
	Batches     int
	Readings    int
	Accepted    int
	Pending     int
	Vendors     []model.Vendor
	Fields      []string
	Narrative   string
}

func BuildReport(batches []model.SensorBatch, now time.Time) Report {
	report := Report{GeneratedAt: now.UTC(), Batches: len(batches)}
	vendorSet := make(map[model.Vendor]bool)
	fieldSet := make(map[string]bool)
	for _, batch := range batches {
		stats := batch.Stats()
		report.Readings += stats.Total
		report.Accepted += stats.Accepted
		report.Pending += stats.Pending
		for vendor := range stats.ByVendor {
			vendorSet[vendor] = true
		}
		for field := range stats.ByField {
			fieldSet[field] = true
		}
	}
	for vendor := range vendorSet {
		report.Vendors = append(report.Vendors, vendor)
	}
	for field := range fieldSet {
		report.Fields = append(report.Fields, field)
	}
	sort.Slice(report.Vendors, func(i, j int) bool { return report.Vendors[i] < report.Vendors[j] })
	sort.Strings(report.Fields)
	report.Narrative = fmt.Sprintf("%d batches, %d readings, %d accepted, %d pending", report.Batches, report.Readings, report.Accepted, report.Pending)
	return report
}

func (r Report) Complete() bool { return r.Readings > 0 && r.Pending == 0 }

func (r Report) VendorLine() string {
	names := make([]string, 0, len(r.Vendors))
	for _, vendor := range r.Vendors {
		names = append(names, string(vendor))
	}
	return strings.Join(names, ",")
}

func (r Report) FieldLine() string { return strings.Join(r.Fields, ",") }

func (r Report) Status() string {
	if r.Readings == 0 {
		return "empty"
	}
	if r.Pending > 0 {
		return "partial"
	}
	return "complete"
}
