package client

import (
	"context"
	"fmt"
)

// MockClient implements Client with fake data for demo/testing.
type MockClient struct{}

// DemoData holds pre-loaded mock data for immediate display.
type DemoData struct {
	VMs        []VMInfo
	Hosts      []HostInfo
	Datastores []DatastoreInfo
	ACLs       []ACLInfo
	Quotas     []QuotaInfo
}

// NewMock creates a MockClient.
func NewMock() *MockClient { return &MockClient{} }

// Preload returns all mock data for instant display without bubbletea commands.
func (m *MockClient) Preload() DemoData {
	vms, _ := m.ListVMs(context.Background())
	hosts, _ := m.ListHosts(context.Background())
	ds, _ := m.ListDatastores(context.Background())
	acls, _ := m.ListACLs(context.Background())
	quotas, _ := m.ListQuotas(context.Background())
	return DemoData{VMs: vms, Hosts: hosts, Datastores: ds, ACLs: acls, Quotas: quotas}
}

func (m *MockClient) ListVMs(_ context.Context) ([]VMInfo, error) {
	return []VMInfo{
		{ID: 1, Name: "web-frontend-01", State: "ACTIVE/RUNNING", User: "admin", Group: "oneadmin", CPU: "2", Memory: "4096", IP: "10.0.1.11", Host: "compute-01", DeployID: "one-1", StartTime: "2025-09-01 10:00:00"},
		{ID: 2, Name: "web-frontend-02", State: "ACTIVE/RUNNING", User: "admin", Group: "oneadmin", CPU: "2", Memory: "4096", IP: "10.0.1.12", Host: "compute-02", DeployID: "one-2", StartTime: "2025-09-01 10:05:00"},
		{ID: 3, Name: "db-primary", State: "ACTIVE/RUNNING", User: "dba", Group: "db-team", CPU: "4", Memory: "16384", IP: "10.0.1.21", Host: "compute-01", DeployID: "one-3", StartTime: "2025-08-15 08:30:00"},
		{ID: 4, Name: "db-replica-01", State: "ACTIVE/RUNNING", User: "dba", Group: "db-team", CPU: "2", Memory: "8192", IP: "10.0.1.22", Host: "compute-03", DeployID: "one-4", StartTime: "2025-08-15 09:00:00"},
		{ID: 5, Name: "k8s-master-01", State: "ACTIVE/RUNNING", User: "devops", Group: "platform", CPU: "4", Memory: "8192", IP: "10.0.1.31", Host: "compute-01", DeployID: "one-5", StartTime: "2025-07-20 14:00:00"},
		{ID: 6, Name: "k8s-worker-01", State: "ACTIVE/RUNNING", User: "devops", Group: "platform", CPU: "4", Memory: "16384", IP: "10.0.1.32", Host: "compute-02", DeployID: "one-6", StartTime: "2025-07-20 14:10:00"},
		{ID: 7, Name: "k8s-worker-02", State: "ACTIVE/RUNNING", User: "devops", Group: "platform", CPU: "4", Memory: "16384", IP: "10.0.1.33", Host: "compute-04", DeployID: "one-7", StartTime: "2025-07-20 14:15:00"},
		{ID: 8, Name: "monitoring-stack", State: "ACTIVE/RUNNING", User: "ops", Group: "platform", CPU: "2", Memory: "4096", IP: "10.0.1.41", Host: "compute-03", DeployID: "one-8", StartTime: "2025-08-01 09:00:00"},
		{ID: 9, Name: "dev-sandbox-01", State: "POWEROFF", User: "dev1", Group: "dev-team", CPU: "2", Memory: "4096", IP: "10.0.1.51", Host: "", DeployID: "", StartTime: "2025-09-10 16:00:00"},
		{ID: 10, Name: "test-runner", State: "SUSPENDED", User: "ci", Group: "dev-team", CPU: "2", Memory: "8192", IP: "10.0.1.61", Host: "", DeployID: "", StartTime: "2025-09-05 11:00:00"},
		{ID: 11, Name: "legacy-app", State: "ERROR", User: "legacy", Group: "old", CPU: "1", Memory: "2048", IP: "10.0.1.71", Host: "compute-04", DeployID: "one-11", StartTime: "2025-06-01 08:00:00"},
		{ID: 12, Name: "batch-job-01", State: "ACTIVE/BOOT", User: "hpc", Group: "compute", CPU: "8", Memory: "32768", IP: "10.0.1.81", Host: "compute-02", DeployID: "one-12", StartTime: "2025-09-20 07:00:00"},
	}, nil
}

func (m *MockClient) ListHosts(_ context.Context) ([]HostInfo, error) {
	return []HostInfo{
		{ID: 0, Name: "compute-01", State: "MONITORED", CPU: "16/64", Memory: "48/128 GB", Cluster: "default", VMs: "4"},
		{ID: 1, Name: "compute-02", State: "MONITORED", CPU: "12/64", Memory: "56/128 GB", Cluster: "default", VMs: "3"},
		{ID: 2, Name: "compute-03", State: "MONITORED", CPU: "8/32", Memory: "24/64 GB", Cluster: "gpu", VMs: "2"},
		{ID: 3, Name: "compute-04", State: "DISABLED", CPU: "0/32", Memory: "0/64 GB", Cluster: "gpu", VMs: "2"},
	}, nil
}

func (m *MockClient) ListDatastores(_ context.Context) ([]DatastoreInfo, error) {
	return []DatastoreInfo{
		{ID: 0, Name: "default", Type: "fs", Total: "500 GB", Used: "210 GB", Free: "290 GB"},
		{ID: 1, Name: "ssd-fast", Type: "ceph", Total: "2 TB", Used: "800 GB", Free: "1.2 TB"},
		{ID: 2, Name: "backup", Type: "nfs", Total: "10 TB", Used: "3.5 TB", Free: "6.5 TB"},
	}, nil
}

func (m *MockClient) ListACLs(_ context.Context) ([]ACLInfo, error) {
	return []ACLInfo{
		{ID: 0, User: "#0", Resource: "VM+HOST+NET+IMG+USER+TPL+DS", Rights: "USE+MANAGE+ADMIN+CREATE", Zone: "*"},
		{ID: 1, User: "#1", Resource: "VM", Rights: "USE+MANAGE", Zone: "#0"},
		{ID: 2, User: "#2", Resource: "VM+HOST", Rights: "USE", Zone: "#0"},
		{ID: 3, User: "@5", Resource: "VM", Rights: "USE+MANAGE", Zone: "#0"},
		{ID: 4, User: "#10", Resource: "IMG", Rights: "USE", Zone: "#0"},
		{ID: 5, User: "#10", Resource: "TPL", Rights: "USE+CREATE", Zone: "#0"},
		{ID: 6, User: "*", Resource: "ZONE", Rights: "USE", Zone: "#0"},
		{ID: 7, User: "#15", Resource: "DS/0", Rights: "USE+MANAGE", Zone: "#0"},
	}, nil
}

func (m *MockClient) ListQuotas(_ context.Context) ([]QuotaInfo, error) {
	return []QuotaInfo{
		{UserID: 0, Entity: "oneadmin", VMs: "2/-1", CPU: "8/-1", Memory: "32768/-1 MB", RunningVMs: "2/-1", Images: "10/-1", Size: "100/-1 GB", Leases: "2/-1", VMsLimit: -1, CPULimit: -1, MemoryLimit: -1, RunningVMsLimit: -1, ImagesLimit: -1, SizeLimit: -1, LeasesLimit: -1},
		{UserID: 1, Entity: "admin", VMs: "4/10", CPU: "8/20", Memory: "16384/65536 MB", RunningVMs: "3/10", Images: "5/20", Size: "50/200 GB", Leases: "3/10", VMsLimit: 10, CPULimit: 20, MemoryLimit: 65536, RunningVMsLimit: 10, ImagesLimit: 20, SizeLimit: 200, LeasesLimit: 10},
		{UserID: 2, Entity: "dba", VMs: "2/5", CPU: "6/12", Memory: "24576/32768 MB", RunningVMs: "2/5", Images: "3/10", Size: "200/500 GB", Leases: "2/5", VMsLimit: 5, CPULimit: 12, MemoryLimit: 32768, RunningVMsLimit: 5, ImagesLimit: 10, SizeLimit: 500, LeasesLimit: 5},
		{UserID: 10, Entity: "dev1", VMs: "1/3", CPU: "2/4", Memory: "4096/8192 MB", RunningVMs: "0/3", Images: "2/5", Size: "10/50 GB", Leases: "0/3", VMsLimit: 3, CPULimit: 4, MemoryLimit: 8192, RunningVMsLimit: 3, ImagesLimit: 5, SizeLimit: 50, LeasesLimit: 3},
		{UserID: 15, Entity: "devops", VMs: "3/8", CPU: "12/16", Memory: "40960/65536 MB", RunningVMs: "3/8", Images: "4/10", Size: "80/200 GB", Leases: "3/8", VMsLimit: 8, CPULimit: 16, MemoryLimit: 65536, RunningVMsLimit: 8, ImagesLimit: 10, SizeLimit: 200, LeasesLimit: 8},
		{UserID: 20, Entity: "ci", VMs: "1/2", CPU: "2/4", Memory: "8192/8192 MB", RunningVMs: "1/2", Images: "1/3", Size: "50/100 GB", Leases: "1/2", VMsLimit: 2, CPULimit: 4, MemoryLimit: 8192, RunningVMsLimit: 2, ImagesLimit: 3, SizeLimit: 100, LeasesLimit: 2},
	}, nil
}

func (m *MockClient) VMAction(_ context.Context, _ int, _ string) error { return nil }

func (m *MockClient) VMMigrate(_ context.Context, _ int, _ int, _ bool) error { return nil }

func (m *MockClient) HostAction(_ context.Context, _ int, _ string) error { return nil }

func (m *MockClient) HostDelete(_ context.Context, _ int) error { return nil }

func (m *MockClient) HostRename(_ context.Context, _ int, _ string) error { return nil }

func (m *MockClient) QuotaUpdate(_ context.Context, _ int, _ string) error { return nil }

func (m *MockClient) GetVMInfo(_ context.Context, id int) (*VMInfo, error) {
	vms, _ := m.ListVMs(context.Background())
	for i := range vms {
		if vms[i].ID == id {
			return &vms[i], nil
		}
	}
	return nil, nil
}

func (m *MockClient) GetVMDetailInfo(_ context.Context, id int) (VMDetail, error) {
	vms, _ := m.ListVMs(context.Background())
	for _, v := range vms {
		if v.ID == id {
			return VMDetail{
				ID: v.ID, Name: v.Name, State: v.State, User: v.User, Group: v.Group,
				CPU: v.CPU, Memory: v.Memory, VCPU: "2", IP: v.IP,
				MAC: "02:00:0a:00:01:" + fmt.Sprintf("%02x", id),
				Network: "private-vlan", Bridge: "br0",
				Host: v.Host, Cluster: "default", DeployID: v.DeployID,
				StartTime: v.StartTime, EndTime: "",
			}, nil
		}
	}
	return VMDetail{}, nil
}

func (m *MockClient) GetHostDetailInfo(_ context.Context, id int) (HostDetail, error) {
	hosts, _ := m.ListHosts(context.Background())
	for _, h := range hosts {
		if h.ID == id {
			return HostDetail{
				ID: h.ID, Name: h.Name, State: h.State,
				IMMAD: "vmware", VMMAD: "kvm",
				ClusterID: 0, ClusterName: h.Cluster,
				CPU: h.CPU, Memory: h.Memory, RunningVMs: 4,
				TotalCPU: 64, TotalMem: 131072, ShareCPU: 16, ShareMem: 32768,
			}, nil
		}
	}
	return HostDetail{}, nil
}

func (m *MockClient) GetHostIDByName(_ context.Context, name string) (int, error) {
	hosts, _ := m.ListHosts(context.Background())
	for _, h := range hosts {
		if h.Name == name {
			return h.ID, nil
		}
	}
	return 0, nil
}
