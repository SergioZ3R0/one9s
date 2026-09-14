package tui

import (
	"time"

	"github.com/scabello/one9s/internal/client"
)

// --- View / Navigation messages ---

type viewName string

const (
	viewVMs        viewName = "vms"
	viewHosts      viewName = "hosts"
	viewDatastores viewName = "datastores"
)

type switchViewMsg struct{ view viewName }

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

// --- Polling ---

type tickMsg time.Time

// --- Action results ---

type actionResultMsg struct {
	resource string
	id       int
	action   string
	err      error
}

// --- Error ---

type errorMsg struct{ err error }
