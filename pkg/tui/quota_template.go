package tui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/scabello/one9s/internal/client"
)

// parseQuotaField returns (value, present). Empty field means "keep current".
func parseQuotaField(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return v, true
}

// quotaVal resolves a form field against a current limit.
// empty → current; otherwise the parsed value (-1 default, -2 unlimited).
func quotaVal(field string, current int) int {
	if v, ok := parseQuotaField(field); ok {
		return v
	}
	return current
}

// optionalQuotaVal is like quotaVal but maps unset/0 current to -1 (default)
// for optional VM fields that OpenNebula stores as -1 when not configured.
func optionalQuotaVal(field string, current int) int {
	if v, ok := parseQuotaField(field); ok {
		return v
	}
	if current == 0 {
		return -1
	}
	return current
}

// buildQuotaTemplate builds a full OpenNebula user quota template.
//
//   - VM section is always rewritten with resolved limits.
//   - DATASTORE / NETWORK / IMAGE sections include every current entry (so
//     per-resource limits are preserved) plus optional "new_*" form fields.
//   - OpenNebula merges section-wise: resources absent from the template keep
//     their previous limits; fields inside a rewritten VM section reset to -1
//     when omitted, so we always emit the full VM block.
func buildQuotaTemplate(values map[string]string, current client.QuotaInfo) string {
	var parts []string

	// --- VM ---
	vmParts := []string{
		fmt.Sprintf("VMS = %d", quotaVal(values["vms"], current.VMsLimit)),
		fmt.Sprintf("CPU = %d", quotaVal(values["cpu"], current.CPULimit)),
		fmt.Sprintf("MEMORY = %d", quotaVal(values["memory"], current.MemoryLimit)),
		fmt.Sprintf("RUNNING_VMS = %d", quotaVal(values["running"], current.RunningVMsLimit)),
		fmt.Sprintf("RUNNING_CPU = %d", optionalQuotaVal(values["running_cpu"], current.RunningCPULimit)),
		fmt.Sprintf("RUNNING_MEMORY = %d", optionalQuotaVal(values["running_memory"], current.RunningMemoryLimit)),
		fmt.Sprintf("SYSTEM_DISK_SIZE = %d", optionalQuotaVal(values["system_disk"], current.SystemDiskLimit)),
	}
	parts = append(parts, fmt.Sprintf("VM = [\n  %s\n]", strings.Join(vmParts, ",\n  ")))

	// --- Datastores ---
	// OpenNebula user quota syntax uses top-level DATASTORE = [...] blocks
	// (not DATASTORE_QUOTA wrappers) in the quota template.
	dsIDs := map[int]struct{}{}
	var dsParts []string
	dsSorted := append([]client.ResourceQuota(nil), current.Datastores...)
	sort.Slice(dsSorted, func(i, j int) bool { return dsSorted[i].ID < dsSorted[j].ID })
	for _, ds := range dsSorted {
		dsIDs[ds.ID] = struct{}{}
		size := quotaVal(values[fmt.Sprintf("ds_%d_size", ds.ID)], ds.Limit)
		images := quotaVal(values[fmt.Sprintf("ds_%d_images", ds.ID)], ds.ImagesLimit)
		dsParts = append(dsParts, fmt.Sprintf(
			"DATASTORE = [\n  ID = %d,\n  SIZE = %d,\n  IMAGES = %d\n]",
			ds.ID, size, images,
		))
	}
	if id, ok := parseQuotaField(values["new_ds_id"]); ok && id > 0 {
		if _, exists := dsIDs[id]; !exists {
			size := quotaVal(values["new_ds_size"], -2)
			images := quotaVal(values["new_ds_images"], -2)
			dsParts = append(dsParts, fmt.Sprintf(
				"DATASTORE = [\n  ID = %d,\n  SIZE = %d,\n  IMAGES = %d\n]",
				id, size, images,
			))
		}
	}
	parts = append(parts, dsParts...)

	// --- Networks ---
	var netParts []string
	netSorted := append([]client.ResourceQuota(nil), current.Networks...)
	sort.Slice(netSorted, func(i, j int) bool { return netSorted[i].ID < netSorted[j].ID })
	for _, n := range netSorted {
		leases := quotaVal(values[fmt.Sprintf("net_%d_leases", n.ID)], n.Limit)
		netParts = append(netParts, fmt.Sprintf("NETWORK = [\n  ID = %d,\n  LEASES = %d\n]", n.ID, leases))
	}
	if id, ok := parseQuotaField(values["new_net_id"]); ok && id > 0 {
		leases := quotaVal(values["new_net_leases"], -2)
		netParts = append(netParts, fmt.Sprintf("NETWORK = [\n  ID = %d,\n  LEASES = %d\n]", id, leases))
	}
	parts = append(parts, netParts...)

	// --- Images (RVMS) ---
	var imgParts []string
	imgSorted := append([]client.ResourceQuota(nil), current.ImagesList...)
	sort.Slice(imgSorted, func(i, j int) bool { return imgSorted[i].ID < imgSorted[j].ID })
	for _, im := range imgSorted {
		rvms := quotaVal(values[fmt.Sprintf("img_%d_rvms", im.ID)], im.Limit)
		imgParts = append(imgParts, fmt.Sprintf("IMAGE = [\n  ID = %d,\n  RVMS = %d\n]", im.ID, rvms))
	}
	parts = append(parts, imgParts...)

	return strings.Join(parts, "\n")
}

// quotaDetailText renders a human-readable quota breakdown for the info modal.
func quotaDetailText(q client.QuotaInfo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "User: %s (id=%d)\n\n", q.Entity, q.UserID)
	fmt.Fprintf(&b, "VM quotas\n")
	fmt.Fprintf(&b, "  VMs          %s\n", q.VMs)
	fmt.Fprintf(&b, "  CPU          %s\n", q.CPU)
	fmt.Fprintf(&b, "  Memory       %s\n", q.Memory)
	fmt.Fprintf(&b, "  Running VMs  %s\n", q.RunningVMs)
	fmt.Fprintf(&b, "  Running CPU  %d\n", q.RunningCPULimit)
	fmt.Fprintf(&b, "  Running Mem  %d MB\n", q.RunningMemoryLimit)
	fmt.Fprintf(&b, "  Sys disk     %d MB\n", q.SystemDiskLimit)

	fmt.Fprintf(&b, "\nDatastore quotas (%d)\n", len(q.Datastores))
	if len(q.Datastores) == 0 {
		fmt.Fprintf(&b, "  (none)\n")
	}
	ds := append([]client.ResourceQuota(nil), q.Datastores...)
	sort.Slice(ds, func(i, j int) bool { return ds[i].ID < ds[j].ID })
	for _, d := range ds {
		name := d.Name
		if name == "" {
			name = fmt.Sprintf("ds-%d", d.ID)
		}
		fmt.Fprintf(&b, "  [%d] %-16s size %s  images %s\n", d.ID, name, d.LimitText, d.ImagesText)
	}

	fmt.Fprintf(&b, "\nNetwork quotas (%d)\n", len(q.Networks))
	if len(q.Networks) == 0 {
		fmt.Fprintf(&b, "  (none)\n")
	}
	ns := append([]client.ResourceQuota(nil), q.Networks...)
	sort.Slice(ns, func(i, j int) bool { return ns[i].ID < ns[j].ID })
	for _, n := range ns {
		name := n.Name
		if name == "" {
			name = fmt.Sprintf("net-%d", n.ID)
		}
		fmt.Fprintf(&b, "  [%d] %-16s leases %s\n", n.ID, name, n.LimitText)
	}

	fmt.Fprintf(&b, "\nImage quotas (%d)\n", len(q.ImagesList))
	if len(q.ImagesList) == 0 {
		fmt.Fprintf(&b, "  (none)\n")
	}
	ims := append([]client.ResourceQuota(nil), q.ImagesList...)
	sort.Slice(ims, func(i, j int) bool { return ims[i].ID < ims[j].ID })
	for _, im := range ims {
		name := im.Name
		if name == "" {
			name = fmt.Sprintf("img-%d", im.ID)
		}
		fmt.Fprintf(&b, "  [%d] %-16s rvms %s\n", im.ID, name, im.LimitText)
	}

	fmt.Fprintf(&b, "\n[-1]=default  [-2]=unlimited\n")
	fmt.Fprintf(&b, "Empty form field keeps the current limit.\n")
	return b.String()
}

// dsSummary compresses datastore quotas for the list table.
func dsSummary(q client.QuotaInfo) string {
	if len(q.Datastores) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(q.Datastores))
	ds := append([]client.ResourceQuota(nil), q.Datastores...)
	sort.Slice(ds, func(i, j int) bool { return ds[i].ID < ds[j].ID })
	for i, d := range ds {
		if i >= 2 {
			parts = append(parts, fmt.Sprintf("+%d", len(ds)-2))
			break
		}
		parts = append(parts, fmt.Sprintf("%d:%s", d.ID, d.LimitText))
	}
	return strings.Join(parts, " ")
}

// quotaEditForm builds the dynamic edit form for one user's quotas.
func quotaEditForm(q client.QuotaInfo) []formField {
	var fields []formField
	add := func(label, key, val string) {
		ti := textinput.New()
		ti.CharLimit = 16
		ti.SetValue(val)
		fields = append(fields, formField{label: label, key: key, input: ti})
	}

	add("VMs", "vms", itoa(q.VMsLimit))
	add("CPU", "cpu", itoa(q.CPULimit))
	add("Memory (MB)", "memory", itoa(q.MemoryLimit))
	add("Running VMs", "running", itoa(q.RunningVMsLimit))
	add("Running CPU", "running_cpu", itoa(q.RunningCPULimit))
	add("Running Mem (MB)", "running_memory", itoa(q.RunningMemoryLimit))
	add("Sys disk (MB)", "system_disk", itoa(q.SystemDiskLimit))

	for _, ds := range q.Datastores {
		name := ds.Name
		if name == "" {
			name = fmt.Sprintf("ds-%d", ds.ID)
		}
		add(fmt.Sprintf("DS %d %s size MB", ds.ID, name), fmt.Sprintf("ds_%d_size", ds.ID), itoa(ds.Limit))
		add(fmt.Sprintf("DS %d %s images", ds.ID, name), fmt.Sprintf("ds_%d_images", ds.ID), itoa(ds.ImagesLimit))
	}
	for _, n := range q.Networks {
		name := n.Name
		if name == "" {
			name = fmt.Sprintf("net-%d", n.ID)
		}
		add(fmt.Sprintf("Net %d %s leases", n.ID, name), fmt.Sprintf("net_%d_leases", n.ID), itoa(n.Limit))
	}
	for _, im := range q.ImagesList {
		name := im.Name
		if name == "" {
			name = fmt.Sprintf("img-%d", im.ID)
		}
		add(fmt.Sprintf("Img %d %s rvms", im.ID, name), fmt.Sprintf("img_%d_rvms", im.ID), itoa(im.Limit))
	}

	add("NEW DS id", "new_ds_id", "")
	add("NEW DS size MB", "new_ds_size", "")
	add("NEW DS images", "new_ds_images", "")
	add("NEW Net id", "new_net_id", "")
	add("NEW Net leases", "new_net_leases", "")

	if len(fields) > 0 {
		fields[0].input.Focus()
	}
	return fields
}

func itoa(v int) string { return fmt.Sprintf("%d", v) }
