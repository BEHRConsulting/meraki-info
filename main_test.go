package main

import (
	"testing"

	"meraki-info/internal/meraki"
)

// TestOrganizationsToReport checks how -org narrows the organizations a
// consolidated report walks. Before this helper existed, the consolidated
// licenses, down and alerting reports ignored -org and listed every
// organization the API key could see.
func TestOrganizationsToReport(t *testing.T) {
	orgs := []meraki.Organization{{ID: "1", Name: "One"}, {ID: "2", Name: "Two"}}

	if got := organizationsToReport(orgs, ""); len(got) != 2 {
		t.Errorf("no -org: got %d organizations, want all 2", len(got))
	}

	got := organizationsToReport(orgs, "2")
	if len(got) != 1 || got[0].ID != "2" {
		t.Errorf("-org 2: got %+v, want only organization 2", got)
	}

	if got := organizationsToReport(orgs, "3"); len(got) != 0 {
		t.Errorf("-org 3 (not visible): got %+v, want none", got)
	}
}
