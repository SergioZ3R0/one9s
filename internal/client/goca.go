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
			hostname = v.HistoryRecords[len(v.HistoryRecords)-1].Hostname
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

func (g *GOCAClient) ListNetworks(ctx context.Context) ([]NetworkInfo, error) {
	pool, err := g.controller.VirtualNetworks().InfoContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("vnpool.info: %w", err)
	}
	out := make([]NetworkInfo, 0, len(pool.VirtualNetworks))
	for i := range pool.VirtualNetworks {
		n := &pool.VirtualNetworks[i]
		total := 0
		for _, ar := range n.ARs {
			total += ar.Size
		}
		out = append(out, NetworkInfo{
			ID:        n.ID,
			Name:      n.Name,
			Bridge:    n.Bridge,
			VNMad:     n.VNMad,
			Used:      n.UsedLeases,
			Total:     total,
			UsedText:  fmt.Sprintf("%d", n.UsedLeases),
			TotalText: formatQuotaInt(total),
			Owner:     n.UName,
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

	// Resolve resource names once for all users.
	dsNames := map[int]string{}
	if pool, err := g.controller.Datastores().InfoContext(ctx); err == nil {
		for _, ds := range pool.Datastores {
			dsNames[ds.ID] = ds.Name
		}
	}
	netNames := map[int]string{}
	if pool, err := g.controller.VirtualNetworks().InfoContext(ctx); err == nil {
		for _, n := range pool.VirtualNetworks {
			netNames[n.ID] = n.Name
		}
	}
	imgNames := map[int]string{}
	if pool, err := g.controller.Images().InfoContext(ctx); err == nil {
		for _, im := range pool.Images {
			imgNames[im.ID] = im.Name
		}
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
				qi.RunningCPULimit = int(vmq.RunningCPU)
				qi.RunningMemoryLimit = vmq.RunningMemory
				qi.SystemDiskLimit = int(vmq.SystemDiskSize)
			}

			// All datastore quotas (per ID), not just [0].
			for _, dsq := range info.Datastore {
				rq := ResourceQuota{
					ID:          dsq.ID,
					Name:        dsNames[dsq.ID],
					Used:        dsq.SizeUsed,
					Limit:       dsq.Size,
					ImagesLimit: dsq.Images,
					ImagesUsed:  dsq.ImagesUsed,
					UsedText:    fmt.Sprintf("%d/%d MB", dsq.SizeUsed, dsq.Size),
					LimitText:   formatQuotaInt(dsq.Size),
					ImagesText:  fmt.Sprintf("%d/%d", dsq.ImagesUsed, dsq.Images),
				}
				qi.Datastores = append(qi.Datastores, rq)
			}
			// Summary columns: first DS as Images/Size for table compatibility.
			if len(info.Datastore) > 0 {
				dsq := info.Datastore[0]
				qi.Images = fmt.Sprintf("%d/%d", dsq.ImagesUsed, dsq.Images)
				qi.Size = fmt.Sprintf("%d/%d MB", dsq.SizeUsed, dsq.Size)
			}

			for _, nq := range info.Network {
				qi.Networks = append(qi.Networks, ResourceQuota{
					ID:        nq.ID,
					Name:      netNames[nq.ID],
					Used:      nq.LeasesUsed,
					Limit:     nq.Leases,
					UsedText:  fmt.Sprintf("%d/%d", nq.LeasesUsed, nq.Leases),
					LimitText: formatQuotaInt(nq.Leases),
				})
			}
			if len(info.Network) > 0 {
				nq := info.Network[0]
				qi.Leases = fmt.Sprintf("%d/%d", nq.LeasesUsed, nq.Leases)
			}

			for _, imq := range info.Image {
				qi.ImagesList = append(qi.ImagesList, ResourceQuota{
					ID:        imq.ID,
					Name:      imgNames[imq.ID],
					Used:      imq.RVMsUsed,
					Limit:     imq.RVMs,
					UsedText:  fmt.Sprintf("%d/%d", imq.RVMsUsed, imq.RVMs),
					LimitText: formatQuotaInt(imq.RVMs),
				})
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

func formatQuotaInt(v int) string {
	switch v {
	case -1:
		return "default"
	case -2:
		return "unlimited"
	default:
		return fmt.Sprintf("%d", v)
	}
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
		hostname = vm.HistoryRecords[len(vm.HistoryRecords)-1].Hostname
		cluster = fmt.Sprintf("%d", vm.HistoryRecords[len(vm.HistoryRecords)-1].CID)
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
