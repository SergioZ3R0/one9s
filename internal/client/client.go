package client

import (
	"context"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
)

// VMInfo holds the flattened fields the TUI needs from a VM.
type VMInfo struct {
	ID        int
	Name      string
	State     string
	LCMState  string
	User      string
	Group     string
	CPU       string
	Memory    string
	IP        string
	Host      string
	DeployID  string
	StartTime string
}

// HostInfo holds flattened host fields for the TUI.
type HostInfo struct {
	ID      int
	Name    string
	State   string
	CPU     string
	Memory  string
	Cluster string
	VMs     string
}

// DatastoreInfo holds flattened datastore fields for the TUI.
type DatastoreInfo struct {
	ID    int
	Name  string
	Type  string
	Total string
	Used  string
	Free  string
}

// Client is the interface the TUI layer programs against.
type Client interface {
	ListVMs(ctx context.Context) ([]VMInfo, error)
	ListHosts(ctx context.Context) ([]HostInfo, error)
	ListDatastores(ctx context.Context) ([]DatastoreInfo, error)
	VMAction(ctx context.Context, id int, action string) error
	VMMigrate(ctx context.Context, id, hostID int, live bool) error
	GetHostIDByName(ctx context.Context, name string) (int, error)
	GetVMIP(v *vm.VM) string
}

// Pre-allocated state lookup arrays (index = state int).
var vmStates = [18]string{
	0: "INIT", 1: "READY", 2: "BOOT_POWEROFF", 3: "BOOT_SUSPENDED",
	4: "BOOT_STOPPED", 5: "SHUTDOWN", 6: "UNDEPLOYED", 7: "BOOT_UNDEPLOYED",
	8: "HOTPLUG", 9: "CLONING", 10: "CLONING_FAILURE", 11: "UNKNOWN",
	12: "HOLD", 13: "STOPPED", 14: "SUSPENDED", 15: "POWEROFF",
	16: "UNKNOWN", 17: "ACTIVE",
}

var lcmStates = [34]string{
	0: "INIT", 1: "PROLOG", 2: "BOOT", 3: "RUNNING",
	4: "MIGRATE", 5: "SNAPSHOT", 6: "STOPPED", 7: "SUSPENDED",
	8: "DONE", 9: "SHUTDOWN_POWEROFF", 10: "UNDPLOY_POWEROFF",
	11: "UNDPLOY_SUSPENDED", 12: "UNDPLOY_STOPPED", 13: "BOOT_UNDEPLOYED",
	14: "BOOT_UNKNOWN", 15: "BOOT_POWEROFF", 16: "BOOT_SUSPENDED",
	17: "BOOT_STOPPED", 18: "FAILURE", 19: "PROLOG_MIGRATE_POWEROFF",
	20: "PROLOG_MIGRATE_SUSPENDED", 21: "PROLOG_MIGRATE_STOPPED",
	22: "PROLOG_RESUME", 23: "DISK_SNAPSHOT_POWEROFF", 24: "DISK_SNAPSHOT_SUSPENDED",
	25: "DISK_SNAPSHOT_STOPPED", 26: "PROLOG_UNDEPLOY", 27: "EPILOG_STOPPED",
	28: "EPILOG_UNDEPLOY", 29: "BOOT_MIGRATE", 30: "BOOT_FAILURE",
	31: "BOOT_MIGRATE_FAILURE", 32: "BOOT_NIC_FAILURE", 33: "BOOT_SCSI_FAILURE",
}

var hostStates = [6]string{
	0: "INIT", 1: "MONITORING", 2: "MONITORED", 3: "ERROR",
	4: "DISABLED", 5: "OFFLINE",
}

func MapVMState(state, lcm int) string {
	if state >= 0 && state < len(vmStates) {
		s := vmStates[state]
		if state == 17 {
			if lcm >= 0 && lcm < len(lcmStates) {
				return "ACTIVE/" + lcmStates[lcm]
			}
			return "ACTIVE/UNKNOWN"
		}
		return s
	}
	return "UNKNOWN"
}

func MapHostState(state int) string {
	if state >= 0 && state < len(hostStates) {
		return hostStates[state]
	}
	return "UNKNOWN"
}
