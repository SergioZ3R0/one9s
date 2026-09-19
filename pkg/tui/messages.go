package tui

import "github.com/scabello/one9s/internal/client"

// --- Data fetched messages ---

type vmsFetchedMsg struct {
	vms []client.VMInfo
	err error
}

type hostsFetchedMsg struct {
	hosts []client.HostInfo
	err   error
}

type datastoresFetchedMsg struct {
	datastores []client.DatastoreInfo
	err        error
}

type aclsFetchedMsg struct {
	acls []client.ACLInfo
	err  error
}

type quotasFetchedMsg struct {
	quotas []client.QuotaInfo
	err    error
}

type vmDetailFetchedMsg struct {
	vm client.VMDetail
}

// --- Host actions ---

type hostActionResultMsg struct {
	hostID int
	action string
	err    error
}

// --- Action results ---

type actionResultMsg struct {
	resource string
	id       int
	action   string
	err      error
}

// --- Error ---

type errorMsg struct{ err error }
