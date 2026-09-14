package client

import (
	"context"
	"fmt"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"

	"github.com/scabello/one9s/internal/config"
)

// GOCAClient implements Client using the official GOCA library.
type GOCAClient struct {
	controller *goca.Controller
}

// NewGOCA creates a GOCAClient from config.
func NewGOCA(cfg config.Config) *GOCAClient {
	gocaClient := goca.NewDefaultClient(
		goca.NewConfig(cfg.User, cfg.Password, cfg.Endpoint),
	)
	return &GOCAClient{
		controller: goca.NewController(gocaClient),
	}
}

func (g *GOCAClient) ListVMs(ctx context.Context) ([]VMInfo, error) {
	pool, err := g.controller.VMs().InfoContext(ctx, -2, -1, -1, -1)
	if err != nil {
		return nil, fmt.Errorf("vmpool.info: %w", err)
	}
	out := make([]VMInfo, 0, len(pool.VMs))
	for i := range pool.VMs {
		v := &pool.VMs[i]
		nics := v.Template.GetNICs()
		ip := ""
		if len(nics) > 0 {
			ip, _ = nics[0].Get(shared.IP)
		}
		hostname := ""
		if len(v.HistoryRecords) > 0 {
			hostname = v.HistoryRecords[0].Hostname
		}
		out = append(out, VMInfo{
			ID:       v.ID,
			Name:     v.Name,
			State:    MapVMState(v.StateRaw, v.LCMStateRaw),
			User:     v.UName,
			Group:    v.GName,
			CPU:      getTemplateStr(&v.Template, "CPU"),
			Memory:   getTemplateStr(&v.Template, "MEMORY"),
			IP:       ip,
			Host:     hostname,
			DeployID: v.DeployID,
		})
	}
	return out, nil
}

func (g *GOCAClient) ListHosts(ctx context.Context) ([]HostInfo, error) {
	pool, err := g.controller.Hosts().InfoContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("hostpool.info: %w", err)
	}
	out := make([]HostInfo, 0, len(pool.Hosts))
	for i := range pool.Hosts {
		h := &pool.Hosts[i]
		cpuPct := 0.0
		if h.Share.TotalCPU > 0 {
			cpuPct = float64(h.Share.CPUUsage) / float64(h.Share.TotalCPU) * 100
		}
		memPct := 0.0
		if h.Share.TotalMem > 0 {
			memPct = float64(h.Share.MemUsage) / float64(h.Share.TotalMem) * 100
		}
		out = append(out, HostInfo{
			ID:      h.ID,
			Name:    h.Name,
			State:   MapHostState(h.StateRaw),
			CPU:     fmt.Sprintf("%.0f%%", cpuPct),
			Memory:  fmt.Sprintf("%.0f%%", memPct),
			Cluster: fmt.Sprintf("%d", h.ClusterID),
			VMs:     fmt.Sprintf("%d", h.Share.RunningVMs),
		})
	}
	return out, nil
}

func (g *GOCAClient) ListDatastores(ctx context.Context) ([]DatastoreInfo, error) {
	pool, err := g.controller.Datastores().InfoContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("datastorepool.info: %w", err)
	}
	out := make([]DatastoreInfo, 0, len(pool.Datastores))
	for i := range pool.Datastores {
		d := &pool.Datastores[i]
		out = append(out, DatastoreInfo{
			ID:    d.ID,
			Name:  d.Name,
			Type:  d.Type,
			Total: fmt.Sprintf("%d MB", d.TotalMB),
			Used:  fmt.Sprintf("%d MB", d.UsedMB),
			Free:  fmt.Sprintf("%d MB", d.FreeMB),
		})
	}
	return out, nil
}

func (g *GOCAClient) VMAction(ctx context.Context, id int, action string) error {
	return g.controller.VM(id).ActionContext(ctx, action)
}

func (g *GOCAClient) VMMigrate(ctx context.Context, id, hostID int, live bool) error {
	return g.controller.VM(id).MigrateContext(ctx, hostID, live, false, -1, 0)
}

func (g *GOCAClient) GetHostIDByName(ctx context.Context, name string) (int, error) {
	return g.controller.Hosts().ByNameContext(ctx, name)
}

func (g *GOCAClient) GetVMIP(v *vm.VM) string {
	nics := v.Template.GetNICs()
	if len(nics) == 0 {
		return ""
	}
	ip, _ := nics[0].Get(shared.IP)
	return ip
}

func getTemplateStr(t *vm.Template, key string) string {
	v, err := t.GetStr(key)
	if err != nil {
		return "-"
	}
	return v
}
