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

func (g *GOCAClient) ListACLs(ctx context.Context) ([]ACLInfo, error) {
	pool, err := g.controller.ACLs().InfoContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("acl.info: %w", err)
	}
	out := make([]ACLInfo, 0, len(pool.ACLs))
	for i := range pool.ACLs {
		a := &pool.ACLs[i]
		out = append(out, ACLInfo{
			ID:       a.ID,
			User:     DecodeACLUser(a.User),
			Resource: DecodeACLResource(a.Resource),
			Rights:   DecodeACLRights(a.Rights),
			Zone:     DecodeACLZone(a.Zone),
		})
	}
	return out, nil
}

func (g *GOCAClient) ListQuotas(ctx context.Context) ([]QuotaInfo, error) {
	userPool, err := g.controller.Users().InfoContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("userpool.info: %w", err)
	}

	// Fetch individual user info (with quotas) concurrently
	type result struct {
		idx  int
		name string
		qi   QuotaInfo
	}

	results := make(chan result, len(userPool.Users))
	sem := make(chan struct{}, 10) // limit concurrency

	for i := range userPool.Users {
		u := &userPool.Users[i]
		go func(idx int, uid int, uname string) {
			sem <- struct{}{}
			defer func() { <-sem }()

			info, err := g.controller.User(uid).InfoContext(ctx, false)
			if err != nil {
				results <- result{idx: idx, name: uname}
				return
			}

			qi := QuotaInfo{
				UserID: uid,
				Entity: fmt.Sprintf("user:%s", uname),
			}
			if len(info.VM) > 0 {
				vmq := info.VM[0]
				qi.VMs = fmt.Sprintf("%d/%d", vmq.VMsUsed, vmq.VMs)
				qi.CPU = fmt.Sprintf("%.0f/%d", vmq.CPUUsed, int(vmq.CPU))
				qi.Memory = fmt.Sprintf("%d/%d MB", vmq.MemoryUsed, vmq.Memory)
				qi.RunningVMs = fmt.Sprintf("%d/%d", vmq.RunningVMsUsed, vmq.RunningVMs)
				qi.VMsLimit = vmq.VMs
				qi.CPULimit = int(vmq.CPU)
				qi.MemoryLimit = vmq.Memory
				qi.RunningVMsLimit = vmq.RunningVMs
			}
			if len(info.Datastore) > 0 {
				dsq := info.Datastore[0]
				qi.Images = fmt.Sprintf("%d/%d", dsq.ImagesUsed, dsq.Images)
				qi.Size = fmt.Sprintf("%d/%d MB", dsq.SizeUsed, dsq.Size)
				qi.ImagesLimit = dsq.Images
				qi.SizeLimit = dsq.Size
			}
			if len(info.Network) > 0 {
				nq := info.Network[0]
				qi.Leases = fmt.Sprintf("%d/%d", nq.LeasesUsed, nq.Leases)
				qi.LeasesLimit = nq.Leases
			}
			results <- result{idx: idx, name: uname, qi: qi}
		}(i, u.ID, u.Name)
	}

	out := make([]QuotaInfo, len(userPool.Users))
	for i := 0; i < len(userPool.Users); i++ {
		r := <-results
		if r.qi.Entity == "" {
			r.qi.Entity = fmt.Sprintf("user:%s", r.name)
		}
		out[r.idx] = r.qi
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

func (g *GOCAClient) GetHostDetailInfo(ctx context.Context, id int) (HostDetail, error) {
	h, err := g.controller.Host(id).InfoContext(ctx, false)
	if err != nil {
		return HostDetail{}, fmt.Errorf("host.info: %w", err)
	}

	cpuPct := 0.0
	if h.Share.TotalCPU > 0 {
		cpuPct = float64(h.Share.CPUUsage) / float64(h.Share.TotalCPU) * 100
	}
	memPct := 0.0
	if h.Share.TotalMem > 0 {
		memPct = float64(h.Share.MemUsage) / float64(h.Share.TotalMem) * 100
	}

	return HostDetail{
		ID:          h.ID,
		Name:        h.Name,
		State:       MapHostState(h.StateRaw),
		IMMAD:       h.IMMAD,
		VMMAD:       h.VMMAD,
		ClusterID:   h.ClusterID,
		ClusterName: h.Cluster,
		CPU:         fmt.Sprintf("%.0f%%", cpuPct),
		Memory:      fmt.Sprintf("%.0f%%", memPct),
		RunningVMs:  h.Share.RunningVMs,
		TotalCPU:    h.Share.TotalCPU,
		TotalMem:    h.Share.TotalMem,
		ShareCPU:    h.Share.CPUUsage,
		ShareMem:    h.Share.MemUsage,
	}, nil
}

func (g *GOCAClient) GetVMInfo(ctx context.Context, id int) (*VMInfo, error) {
	vm, err := g.controller.VM(id).InfoContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("vm.info: %w", err)
	}

	nics := vm.Template.GetNICs()
	ip := ""
	if len(nics) > 0 {
		ip, _ = nics[0].Get(shared.IP)
	}

	hostname := ""
	if len(vm.HistoryRecords) > 0 {
		hostname = vm.HistoryRecords[0].Hostname
	}

	return &VMInfo{
		ID:       vm.ID,
		Name:     vm.Name,
		State:    MapVMState(vm.StateRaw, vm.LCMStateRaw),
		User:     vm.UName,
		Group:    vm.GName,
		CPU:      getTemplateStr(&vm.Template, "CPU"),
		Memory:   getTemplateStr(&vm.Template, "MEMORY"),
		IP:       ip,
		Host:     hostname,
		DeployID: vm.DeployID,
	}, nil
}

func (g *GOCAClient) GetVMDetailInfo(ctx context.Context, id int) (VMDetail, error) {
	vm, err := g.controller.VM(id).InfoContext(ctx, false)
	if err != nil {
		return VMDetail{}, fmt.Errorf("vm.info: %w", err)
	}

	nics := vm.Template.GetNICs()
	ip, mac, network, bridge := "", "", "", ""
	if len(nics) > 0 {
		ip, _ = nics[0].Get(shared.IP)
		mac, _ = nics[0].Get(shared.MAC)
		network, _ = nics[0].Get(shared.Network)
		bridge, _ = nics[0].Get(shared.Bridge)
	}

	hostname, cluster := "", ""
	if len(vm.HistoryRecords) > 0 {
		hostname = vm.HistoryRecords[0].Hostname
		cluster = fmt.Sprintf("%d", vm.HistoryRecords[0].CID)
	}

	vcpu := getTemplateStr(&vm.Template, "VCPU")
	if vcpu == "-" {
		vcpu = getTemplateStr(&vm.Template, "CPU")
	}

	return VMDetail{
		ID:       vm.ID,
		Name:     vm.Name,
		State:    MapVMState(vm.StateRaw, vm.LCMStateRaw),
		User:     vm.UName,
		Group:    vm.GName,
		CPU:      getTemplateStr(&vm.Template, "CPU"),
		Memory:   getTemplateStr(&vm.Template, "MEMORY"),
		VCPU:     vcpu,
		IP:       ip,
		MAC:      mac,
		Network:  network,
		Bridge:   bridge,
		Host:     hostname,
		Cluster:  cluster,
		DeployID: vm.DeployID,
	}, nil
}

func (g *GOCAClient) HostAction(ctx context.Context, id int, action string) error {
	switch action {
	case "enable":
		return g.controller.Host(id).StatusContext(ctx, 0)
	case "disable":
		return g.controller.Host(id).StatusContext(ctx, 1)
	case "offline":
		return g.controller.Host(id).StatusContext(ctx, 2)
	}
	return fmt.Errorf("unknown host action: %s", action)
}

func (g *GOCAClient) HostDelete(ctx context.Context, id int) error {
	return g.controller.Host(id).DeleteContext(ctx)
}

func (g *GOCAClient) HostRename(ctx context.Context, id int, name string) error {
	return g.controller.Host(id).RenameContext(ctx, name)
}

func (g *GOCAClient) QuotaUpdate(ctx context.Context, userID int, tpl string) error {
	return g.controller.User(userID).QuotaContext(ctx, tpl)
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
