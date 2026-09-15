package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
)

// VMInfo holds the flattened fields the TUI needs from a VM.
type VMInfo struct {
	ID        int
	Name      string
	State     string
	User      string
	Group     string
	CPU       string
	Memory    string
	IP        string
	Host      string
	DeployID  string
	StartTime string
}

type HostInfo struct {
	ID      int
	Name    string
	State   string
	CPU     string
	Memory  string
	Cluster string
	VMs     string
}

type DatastoreInfo struct {
	ID    int
	Name  string
	Type  string
	Total string
	Used  string
	Free  string
}

type ACLInfo struct {
	ID       int
	User     string
	Resource string
	Rights   string
	Zone     string
}

type QuotaInfo struct {
	Entity     string
	VMs        string
	CPU        string
	Memory     string
	RunningVMs string
	Images     string
	Size       string
	Leases     string
}

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

// VM state names
var vmMainStates = map[int]string{
	0: "INIT", 1: "PENDING", 2: "HOLD", 3: "ACTIVE",
	4: "STOPPED", 5: "SUSPENDED", 6: "POWEROFF", 7: "UNDEPLOYED",
	8: "HOLD", // some deployments use 8 for hold
}

// LCM state names (when STATE=3/ACTIVE)
var lcmStateNames = map[int]string{
	0: "LCM_INIT", 1: "PROLOG", 2: "BOOT", 3: "RUNNING",
	4: "MIGRATE", 5: "SAVE_STOP", 6: "SAVE_SUSPEND", 7: "SAVE_MIGRATE",
	8: "PROLOG_MIGRATE", 9: "PROLOG_RESUME", 10: "EPILOG_STOP", 11: "EPILOG",
	12: "SHUTDOWN", 15: "CLEANUP_RESUBMIT", 16: "UNKNOWN", 17: "HOTPLUG",
	18: "SHUTDOWN_POWEROFF", 19: "BOOT_UNKNOWN", 20: "BOOT_POWEROFF",
	21: "BOOT_SUSPENDED", 22: "BOOT_STOPPED", 23: "CLEANUP_DELETE",
	24: "HOTPLUG_SNAPSHOT", 25: "HOTPLUG_NIC", 26: "HOTPLUG_SAVEAS",
	36: "BOOT_FAILURE", 37: "BOOT_MIGRATE_FAILURE",
}

var hostStateNames = map[int]string{
	0: "INIT", 1: "MONITORING", 2: "MONITORED", 3: "ERROR",
	4: "DISABLED", 8: "OFFLINE",
}

func MapVMState(state, lcm int) string {
	if state == 3 {
		// ACTIVE - show LCM sub-state
		if name, ok := lcmStateNames[lcm]; ok {
			return "ACTIVE/" + name
		}
		return fmt.Sprintf("ACTIVE/UNKNOWN(%d)", lcm)
	}
	if name, ok := vmMainStates[state]; ok {
		return name
	}
	return fmt.Sprintf("STATE(%d)", state)
}

func MapHostState(state int) string {
	if name, ok := hostStateNames[state]; ok {
		return name
	}
	return fmt.Sprintf("STATE(%d)", state)
}

// --- ACL hex decoding ---

// DecodeACLUser decodes hex user field: #UID, @GID, *, %CLUSTER
func DecodeACLUser(hex string) string {
	val, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return hex
	}
	// Check flags
	if val&0x400000000 != 0 { // ALL
		return "*"
	}
	if val&0x100000000 != 0 { // UID
		return fmt.Sprintf("#%d", val&0xFFFFFFFF)
	}
	if val&0x200000000 != 0 { // GID
		return fmt.Sprintf("@%d", val&0xFFFFFFFF)
	}
	if val&0x800000000 != 0 { // CLUSTER
		return fmt.Sprintf("%%%d", val&0xFFFFFFFF)
	}
	return hex
}

// DecodeACLResource decodes hex resource field: TYPE+ID
var aclResourceNames = map[int64]string{
	0x1000000000: "VM", 0x2000000000: "HOST", 0x4000000000: "NET",
	0x8000000000: "IMG", 0x10000000000: "USER", 0x20000000000: "TPL",
	0x40000000000: "GRP", 0x100000000000: "DS", 0x200000000000: "CLUSTER",
	0x400000000000: "DOC", 0x800000000000: "ZONE",
	0x1000000000000: "SECGROUP", 0x2000000000000: "VDC",
	0x4000000000000: "VROUTER", 0x8000000000000: "MARKET",
	0x10000000000000: "MARKETAPP", 0x20000000000000: "VMGROUP",
	0x40000000000000: "VNTEMPLATE",
}

func DecodeACLResource(hex string) string {
	val, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return hex
	}

	// Extract ID bits (lower 32 bits or specific ranges)
	id := val & 0xFFFFFFFF
	var types []string
	for mask, name := range aclResourceNames {
		if val&mask != 0 {
			types = append(types, name)
		}
	}
	if len(types) == 0 {
		return hex
	}
	typeStr := strings.Join(types, "+")
	if id != 0 {
		return fmt.Sprintf("%s/%d", typeStr, id)
	}
	return typeStr
}

// DecodeACLRights decodes hex rights: USE, MANAGE, ADMIN, CREATE
func DecodeACLRights(hex string) string {
	val, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return hex
	}
	var rights []string
	if val&1 != 0 {
		rights = append(rights, "USE")
	}
	if val&2 != 0 {
		rights = append(rights, "MANAGE")
	}
	if val&4 != 0 {
		rights = append(rights, "ADMIN")
	}
	if val&8 != 0 {
		rights = append(rights, "CREATE")
	}
	if len(rights) == 0 {
		return hex
	}
	return strings.Join(rights, "+")
}

// DecodeACLZone decodes hex zone: * or zone ID
func DecodeACLZone(hex string) string {
	val, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return hex
	}
	if val&0x400000000 != 0 {
		return "*"
	}
	id := val & 0xFFFFFFFF
	return fmt.Sprintf("#%d", id)
}
