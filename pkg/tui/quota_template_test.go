package tui

import (
	"strings"
	"testing"

	"github.com/scabello/one9s/internal/client"
)

func TestBuildQuotaTemplatePreservesDatastoreQuotas(t *testing.T) {
	current := client.QuotaInfo{
		UserID:          2,
		Entity:          "user:labuser",
		VMsLimit:        10,
		CPULimit:        4,
		MemoryLimit:     8192,
		RunningVMsLimit: 5,
		Datastores: []client.ResourceQuota{
			{ID: 1, Name: "default", Limit: 10240, ImagesLimit: 5},
			{ID: 3, Name: "ceph", Limit: 1048576, ImagesLimit: 50},
		},
		Networks: []client.ResourceQuota{
			{ID: 1, Name: "lab-public", Limit: 10},
		},
	}

	// User only edits VMs/CPU in the form; DS fields empty → keep current.
	values := map[string]string{
		"vms":     "20",
		"cpu":     "8",
		"memory":  "",
		"running": "",
	}

	out := buildQuotaTemplate(values, current)

	if !strings.Contains(out, "VMS = 20") {
		t.Fatalf("expected VMS=20 in template:\n%s", out)
	}
	if !strings.Contains(out, "CPU = 8") {
		t.Fatalf("expected CPU=8 in template:\n%s", out)
	}
	if !strings.Contains(out, "MEMORY = 8192") {
		t.Fatalf("expected MEMORY preserved 8192:\n%s", out)
	}
	if !strings.Contains(out, "RUNNING_VMS = 5") {
		t.Fatalf("expected RUNNING_VMS preserved 5:\n%s", out)
	}
	if !strings.Contains(out, "DATASTORE = [") {
		t.Fatalf("expected DATASTORE blocks in template:\n%s", out)
	}
	if strings.Contains(out, "DATASTORE_QUOTA") {
		t.Fatalf("quota template must not use DATASTORE_QUOTA wrapper:\n%s", out)
	}
	if !strings.Contains(out, "ID = 1") || !strings.Contains(out, "SIZE = 10240") || !strings.Contains(out, "IMAGES = 5") {
		t.Fatalf("expected datastore id=1 preserved:\n%s", out)
	}
	if !strings.Contains(out, "ID = 3") || !strings.Contains(out, "SIZE = 1048576") || !strings.Contains(out, "IMAGES = 50") {
		t.Fatalf("expected datastore id=3 preserved:\n%s", out)
	}
	if !strings.Contains(out, "LEASES = 10") {
		t.Fatalf("expected network leases preserved:\n%s", out)
	}
}

func TestBuildQuotaTemplateUpdatesDatastoreLimits(t *testing.T) {
	current := client.QuotaInfo{
		VMsLimit:    10,
		CPULimit:    4,
		MemoryLimit: 2048,
		Datastores: []client.ResourceQuota{
			{ID: 1, Limit: 10240, ImagesLimit: 5},
		},
	}
	values := map[string]string{
		"ds_1_size":   "20480",
		"ds_1_images": "20",
	}
	out := buildQuotaTemplate(values, current)
	if !strings.Contains(out, "SIZE = 20480") {
		t.Fatalf("expected updated size:\n%s", out)
	}
	if !strings.Contains(out, "IMAGES = 20") {
		t.Fatalf("expected updated images:\n%s", out)
	}
}

func TestBuildQuotaTemplateAddsNewDatastore(t *testing.T) {
	current := client.QuotaInfo{
		VMsLimit: 5,
		Datastores: []client.ResourceQuota{
			{ID: 1, Limit: 100, ImagesLimit: 2},
		},
	}
	values := map[string]string{
		"new_ds_id":     "4",
		"new_ds_size":   "5000",
		"new_ds_images": "10",
	}
	out := buildQuotaTemplate(values, current)
	if !strings.Contains(out, "ID = 4") || !strings.Contains(out, "SIZE = 5000") || !strings.Contains(out, "IMAGES = 10") {
		t.Fatalf("expected new datastore quota:\n%s", out)
	}
	if !strings.Contains(out, "ID = 1") {
		t.Fatalf("expected existing datastore kept:\n%s", out)
	}
}

func TestDsSummary(t *testing.T) {
	q := client.QuotaInfo{
		Datastores: []client.ResourceQuota{
			{ID: 1, LimitText: "10G"},
			{ID: 3, LimitText: "1T"},
			{ID: 5, LimitText: "2G"},
		},
	}
	got := dsSummary(q)
	if got != "1:10G 3:1T +1" {
		t.Fatalf("dsSummary = %q", got)
	}
}
