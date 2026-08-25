package adapter

import (
	"context"
	"testing"
	"time"
)

func TestFixtureAdapterFiltersIDs(t *testing.T) {
	a := NewAgriSense([]RawReading{{ExternalID: "one", Field: "f", Timestamp: time.Unix(1, 0), Values: map[string]float64{"moisture": 20}}, {ExternalID: "two", Field: "f", Timestamp: time.Unix(1, 0), Values: map[string]float64{"moisture": 30}}})
	rows, err := a.Fetch(context.Background(), []string{"two"})
	if err != nil || len(rows) != 1 || rows[0].ExternalID != "two" {
		t.Fatalf("unexpected fixture result: %v %+v", err, rows)
	}
}

func TestCheckRawReading(t *testing.T) {
	if err := CheckRawReading(RawReading{}); err == nil {
		t.Fatal("expected invalid raw reading")
	}
}

func TestCatalogAndMerge(t *testing.T) {
	first := NewAgriSense([]RawReading{{ExternalID: "a", Field: "f", Timestamp: FixtureTimestamp(), Values: map[string]float64{"m": 1}}})
	second := NewFieldLink([]RawReading{{ExternalID: "b", Field: "f", Timestamp: FixtureTimestamp(), Values: map[string]float64{"m": 2}}})
	catalog := NewCatalog(first)
	if err := catalog.Register(second); err != nil {
		t.Fatal(err)
	}
	if _, ok := catalog.Get(second.Vendor()); !ok || len(catalog.Vendors()) != 2 {
		t.Fatal("catalog lookup failed")
	}
	groups, errorsFound := catalog.FetchAll(context.Background(), nil)
	if len(errorsFound) != 0 || len(MergeRawReadings(groups)) != 2 {
		t.Fatalf("fetch groups: %v %+v", errorsFound, groups)
	}
}
