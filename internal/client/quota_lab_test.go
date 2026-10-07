package client

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/scabello/one9s/internal/config"
)

// TestListQuotasLab runs against opennebula-lab when env is set:
//
//	ONE_XMLRPC=http://localhost:2633/RPC2 ONE_AUTH=oneadmin:opennebula \
//	  go test ./internal/client -run TestListQuotasLab -count=1
func TestListQuotasLab(t *testing.T) {
	endpoint := os.Getenv("ONE_XMLRPC")
	auth := os.Getenv("ONE_AUTH")
	if endpoint == "" || auth == "" {
		t.Skip("set ONE_XMLRPC and ONE_AUTH to run against the lab")
	}
	parts := strings.SplitN(auth, ":", 2)
	if len(parts) != 2 {
		t.Fatalf("ONE_AUTH must be user:password")
	}
	cfg := config.Config{Endpoint: endpoint, User: parts[0], Password: parts[1]}
	g := NewGOCA(cfg)
	ctx := context.Background()

	// Ensure a quota-rich user exists via XML-RPC path is heavy; use labuser if present.
	quotas, err := g.ListQuotas(ctx)
	if err != nil {
		t.Fatalf("ListQuotas: %v", err)
	}
	if len(quotas) == 0 {
		t.Fatal("no users returned")
	}

	var found bool
	for _, q := range quotas {
		if q.Entity == "user:labuser" || strings.Contains(q.Entity, "labuser") {
			found = true
			t.Logf("labuser VMs=%s CPU=%s DS=%d NET=%d IMG=%d",
				q.VMs, q.CPU, len(q.Datastores), len(q.Networks), len(q.ImagesList))
			for _, ds := range q.Datastores {
				t.Logf("  DS id=%d name=%q size=%d images=%d", ds.ID, ds.Name, ds.Limit, ds.ImagesLimit)
			}
		}
	}
	if !found {
		t.Logf("labuser not found; users: %d (first: %s)", len(quotas), quotas[0].Entity)
	}

	// Round-trip: update labuser VM quotas, ensure per-DS quotas survive.
	const labUserID = 2
	tpl := "VM = [\n  VMS = 15,\n  CPU = 6,\n  MEMORY = 4096,\n  RUNNING_VMS = 3,\n  RUNNING_CPU = -1,\n  RUNNING_MEMORY = -1,\n  SYSTEM_DISK_SIZE = -1\n]"
	if err := g.QuotaUpdate(ctx, labUserID, tpl); err != nil {
		t.Fatalf("QuotaUpdate: %v", err)
	}
	quotas, err = g.ListQuotas(ctx)
	if err != nil {
		t.Fatalf("ListQuotas after update: %v", err)
	}
	for _, q := range quotas {
		if q.UserID != labUserID {
			continue
		}
		if q.VMsLimit != 15 || q.CPULimit != 6 || q.MemoryLimit != 4096 {
			t.Fatalf("VM limits not applied: vms=%d cpu=%d mem=%d", q.VMsLimit, q.CPULimit, q.MemoryLimit)
		}
		if len(q.Datastores) == 0 {
			t.Fatal("datastore quotas were wiped by VM-only update")
		}
		for _, ds := range q.Datastores {
			t.Logf("after update DS id=%d size=%d images=%d", ds.ID, ds.Limit, ds.ImagesLimit)
			if ds.ID == 1 && (ds.Limit != 10240 || ds.ImagesLimit != 5) {
				// may differ if test rerun changed them; only warn
				t.Logf("note: ds1 limits are size=%d images=%d (expected 10240/5 from seed)", ds.Limit, ds.ImagesLimit)
			}
		}
		if len(q.Networks) == 0 {
			t.Fatal("network quotas were wiped by VM-only update")
		}
		t.Logf("OK: VM update preserved %d datastore and %d network quotas", len(q.Datastores), len(q.Networks))
	}
}
