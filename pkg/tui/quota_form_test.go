package tui

import (
	"strings"
	"testing"

	"github.com/scabello/one9s/internal/client"
)

func TestQuotaEditFormSections(t *testing.T) {
	q := client.QuotaInfo{
		UserID:     2,
		Entity:     "user:labuser",
		VMsLimit:   15,
		CPULimit:   6,
		Datastores: []client.ResourceQuota{{ID: 1, Name: "default", Limit: 20480, ImagesLimit: 9}},
		Networks:   []client.ResourceQuota{{ID: 1, Name: "lab-nat", Limit: 12}},
	}
	fields := quotaEditForm(q)

	var headers, labels []string
	for _, f := range fields {
		if f.isHeader {
			headers = append(headers, f.label)
			continue
		}
		labels = append(labels, f.label)
	}

	wantHeaders := []string{
		"Virtual machine quotas",
		"Datastore quotas (existing)",
		"Network quotas (existing)",
		"Create datastore quota",
		"Create network quota",
	}
	for _, h := range wantHeaders {
		found := false
		for _, got := range headers {
			if got == h {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing section header %q in %v", h, headers)
		}
	}

	joined := strings.Join(labels, " | ")
	for _, want := range []string{
		"Datastore ID",
		"Size MB",
		"Images",
		"Network ID",
		"Leases",
		"DS 1 default — size MB",
		"Net 1 lab-nat — leases",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing field %q in %v", want, labels)
		}
	}
	// Ensure ADD/NEW naming is gone
	for _, l := range labels {
		if strings.HasPrefix(l, "ADD") || strings.HasPrefix(l, "NEW") {
			t.Fatalf("old field label still present: %q", l)
		}
	}
}

func TestFormFieldIndexSkipsHeaders(t *testing.T) {
	fields := []formField{
		{label: "H1", isHeader: true},
		{label: "A", key: "a"},
		{label: "H2", isHeader: true},
		{label: "B", key: "b"},
	}
	if got := formFieldIndex(fields, -1, 1); got != 1 {
		t.Fatalf("first editable = %d, want 1", got)
	}
	if got := formFieldIndex(fields, 1, 1); got != 3 {
		t.Fatalf("next editable = %d, want 3", got)
	}
	if got := formFieldIndex(fields, 3, 1); got != 1 {
		t.Fatalf("wrap editable = %d, want 1", got)
	}
}
