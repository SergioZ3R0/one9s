package client

import (
	"context"
	"fmt"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
)

// VMInfo holds the flattened fields the TUI needs from a VM.
type VMInfo struct {
	ID        int
	Name      string
	State     string // Human-readable state (e.g. "ACTIVE/RUNNING", "POWEROFF")
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

// ACLInfo holds flattened ACL fields for the TUI.
type ACLInfo struct {
	ID       int
	User     string
	Resource string
	Rights   string
	Zone     string
}

// QuotaInfo holds flattened quota fields for the TUI.
type QuotaInfo struct {
	Entity     string // "user:oneadmin" or "group:oneadmin"
	VMs        string // used/limit
	CPU        string
	Memory     string
	RunningVMs string
	Images     string
	Size       string
	Leases     string
}

// Client is the interface the TUI layer programs against.
type Client interface {
	ListVMs(ctx context.Context) ([]VMInfo, error)
	ListHosts(ctx context.Context) ([]HostInfo, error)
	ListDatastores(ctx context.Context) ([]DatastoreInfo, error)
	ListACLs(ctx context.Context) ([]ACLInfo, error)
	ListQuotas(ctx context.Context) ([]QuotaInfo, error)
	VMAction(ctx context.Context, id int, action string) error
	VMMigrate(ctx context.Context, id, hostID int, live bool) error
	GetHostIDByName(ctx context.Context, name string) (int, error)
	GetVMIP(v *vm.VM) string
}

// OpenNebula VM STATE values (the main state).
// Reference: https://docs.opennebula.io/7.4/product/operation_references/configuration_references/vm_states
var vmMainStates = [8]string{
	0: "INIT",       // Internal init, not visible to users
	1: "PENDING",    // Waiting for scheduler to deploy
	2: "HOLD",       // Held by owner, won't be scheduled
	3: "ACTIVE",     // Active - use LCM_STATE for detail
	4: "STOPPED",    // Stopped, state saved to datastore
	5: "SUSPENDED",  // Suspended, files left on host
	6: "POWEROFF",   // Powered off, files left on host
	7: "UNDEPLOYED", // Shut down, disks in system datastore
}

// OpenNebula LCM_STATE values (only when STATE=3/ACTIVE).
var lcmStates = [58]string{
	0:  "LCM_INIT",
	1:  "PROLOG",
	2:  "BOOT",
	3:  "RUNNING",
	4:  "MIGRATE",
	5:  "SAVE_STOP",
	6:  "SAVE_SUSPEND",
	7:  "SAVE_MIGRATE",
	8:  "PROLOG_MIGRATE",
	9:  "PROLOG_RESUME",
	10: "EPILOG_STOP",
	11: "EPILOG",
	12: "SHUTDOWN",
	// 13-14 unused
	15: "CLEANUP_RESUBMIT",
	16: "UNKNOWN",
	17: "HOTPLUG",
	18: "SHUTDOWN_POWEROFF",
	19: "BOOT_UNKNOWN",
	20: "BOOT_POWEROFF",
	21: "BOOT_SUSPENDED",
	22: "BOOT_STOPPED",
	23: "CLEANUP_DELETE",
	24: "HOTPLUG_SNAPSHOT",
	25: "HOTPLUG_NIC",
	26: "HOTPLUG_SAVEAS",
	27: "HOTPLUG_SAVEAS_POWEROFF",
	28: "HOTPLUG_SAVEAS_SUSPENDED",
	29: "SHUTDOWN_UNDEPLOY",
	30: "EPILOG_UNDEPLOY",
	31: "PROLOG_UNDEPLOY",
	32: "BOOT_UNDEPLOY",
	33: "HOTPLUG_PROLOG_POWEROFF",
	34: "HOTPLUG_EPILOG_POWEROFF",
	35: "BOOT_MIGRATE",
	36: "BOOT_FAILURE",
	37: "BOOT_MIGRATE_FAILURE",
	38: "PROLOG_MIGRATE_FAILURE",
	39: "PROLOG_FAILURE",
	40: "EPILOG_FAILURE",
	41: "EPILOG_STOP_FAILURE",
	42: "EPILOG_UNDEPLOY_FAILURE",
	43: "PROLOG_MIGRATE_POWEROFF",
	44: "PROLOG_MIGRATE_POWEROFF_FAILURE",
	45: "PROLOG_MIGRATE_SUSPEND",
	46: "PROLOG_MIGRATE_SUSPEND_FAILURE",
	47: "BOOT_UNDEPLOY_FAILURE",
	48: "BOOT_STOPPED_FAILURE",
	49: "PROLOG_RESUME_FAILURE",
	50: "PROLOG_UNDEPLOY_FAILURE",
	51: "DISK_SNAPSHOT_POWEROFF",
	52: "DISK_SNAPSHOT_REVERT_POWEROFF",
	53: "DISK_SNAPSHOT_REVERT_SUSPENDED",
	54: "DISK_SNAPSHOT_DELETE_POWEROFF",
	55: "DISK_SNAPSHOT_DELETE_SUSPENDED",
	56: "PROLOG_MIGRATE_SUSPEND",
	57: "PROLOG_MIGRATE_SUSPEND_FAILURE",
}

var hostStates = [9]string{
	0: "INIT",
	1: "MONITORING_MONITORED",
	2: "MONITORED",
	3: "ERROR",
	4: "DISABLED",
	5: "MONITORING_ERROR",
	6: "MONITORING_INIT",
	7: "MONITORING_DISABLED",
	8: "OFFLINE",
}

// MapVMState translates numeric STATE + LCM_STATE to human-readable string.
// STATE=3 (ACTIVE) uses LCM_STATE for detail, otherwise uses main state name.
func MapVMState(state, lcm int) string {
	if state < 0 || state >= len(vmMainStates) {
		return fmt.Sprintf("UNKNOWN(%d)", state)
	}
	main := vmMainStates[state]

	if state != 3 {
		// Not ACTIVE - LCM_STATE is LCM_INIT (irrelevant), just show main state
		return main
	}

	// STATE=ACTIVE - show LCM sub-state
	if lcm >= 0 && lcm < len(lcmStates) && lcmStates[lcm] != "" {
		return main + "/" + lcmStates[lcm]
	}
	return main + "/UNKNOWN"
}

// MapHostState maps host state int to string.
func MapHostState(state int) string {
	if state >= 0 && state < len(hostStates) {
		return hostStates[state]
	}
	return "UNKNOWN"
}
