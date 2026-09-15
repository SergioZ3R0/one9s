<div align="center">
  <img src="https://raw.githubusercontent.com/SergioZ3R0/one9s/main/docs/one9s.svg" alt="one9s" width="180">
</div>

<h1 align="center">one9s</h1>
<p align="center">
  <b>Terminal UI for OpenNebula cluster management</b><br>
  <i>Monitor VMs, Hosts, Datastores &bull; Manage ACLs & Quotas &bull; Lifecycle actions from your terminal</i><br>
  <a href="https://one9s.scszero.com/">one9s.scszero.com</a> &nbsp;|&nbsp;
  <a href="https://one9s.scszero.com/docs.html">Documentation</a>
</p>

`one9s` is a **TUI** (terminal user interface) for managing and monitoring
**OpenNebula** clusters directly from your terminal, inspired by
[k9s](https://github.com/derailed/k9s) and
[lazygit](https://github.com/jesseduffield/lazygit).

No Sunstone web UI. No Ruby CLI dependencies. Just a single binary that talks
to the OpenNebula XML-RPC API.

## About

**one9s** is a terminal user interface built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) that turns the
OpenNebula API into a live, interactive dashboard. It lets you:

- Monitor **VMs**, **Hosts**, **Datastores**, **ACLs** and **Quotas** at a glance.
- Filter VMs by state (active, stopped, poweroff, error) with a single keypress.
- Perform **lifecycle actions**, reboot, stop, resume, suspend, terminate,
  directly from the TUI.
- Manage **Hosts**, enable, disable, offline, delete, rename, with
  confirmation modals for destructive operations.
- Decode **ACL rules** from hex to human-readable format (user, resource,
  rights, zone).
- View **User Quotas**, VMs, CPU, Memory, Running VMs, Images, Leases.

It is written in **Go**, styled with
[Lip Gloss](https://github.com/charmbracelet/lipgloss), and speaks directly to
OpenNebula using the official GOCA library, no SSH, no Sunstone, no Ruby.

## Philosophy

**one9s is a pure API client.** It talks to OpenNebula's XML-RPC API and
nothing else, no SSH, no local CLI commands, no filesystem access. Run it from
your laptop against any cluster with zero dependencies on the cluster's tooling.

This makes one9s:

- **Portable**, a single binary, no OpenNebula CLI installation required.
- **Secure**, no shell access needed; only the XML-RPC endpoint must be reachable.
- **Fast**, zero auto-polling; data is fetched on startup and on explicit refresh (F5).
- **Real-time**, instant navigation with direct line rendering (no viewport overhead).

## Status

`one9s` is under active development. Current features:

- [x] XML-RPC client via GOCA with `ONE_AUTH` / `ONE_XMLRPC` environment variables.
- [x] VM pool view with state, CPU, Memory, IP, Host, state-aware color coding.
- [x] VM state filters: All, Active, Stopped, Poweroff, Error.
- [x] VM lifecycle actions: reboot, stop, resume, suspend, terminate (with confirmation).
- [x] Host pool view with state, CPU%, Memory%, VMs running.
- [x] Host actions: enable, disable, offline, delete, rename (with confirmation modals).
- [x] Datastore pool view with type, Total, Used, Free.
- [x] ACL rules view with decoded hex values (user, resource, rights, zone).
- [x] User Quotas view with VMs, CPU, Memory, Running, Images, Leases.
- [x] Fuzzy text search and state-based filtering.
- [x] Help overlay with context-sensitive key bindings.
- [x] Cross-platform binaries (Linux, macOS, Windows × amd64/arm64).
- [ ] VM migration dialog (host picker).
- [ ] Log streaming.
- [ ] Template instantiation.

## Features

**Tabs** (navigate with `tab`, `?` for help):

- **VMs**, all VMs with state, user, CPU, memory, IP, host. Filter by state or text search.
- **Hosts**, cluster hosts with state, CPU%, memory%, running VMs. Enable/disable/rename/delete.
- **Datastores**, storage pools with type, capacity, usage.
- **ACLs**, decoded access control rules (user, resource, rights, zone).
- **Quotas**, user quotas (VMs, CPU, Memory, Running, Images, Leases).
- **Help**, context-sensitive key bindings for the current tab.

**Key bindings**

| Key | Action |
| --- | ------ |
| `q` / `Ctrl+C` | quit |
| `tab` | next tab |
| `?` | help / back to VMs |
| `↑/↓` / `j/k` | navigate rows |
| `PgUp/PgDn` / `b/f` | page up / page down |
| `g/G` | go to start / end |
| `/` | fuzzy search / filter |
| `F5` | refresh current view |

**VM actions** (VMs tab)

| Key | Action |
| --- | ------ |
| `r` | reboot (ACTIVE only) |
| `s` | stop (ACTIVE only) |
| `u` | resume / start (POWEROFF/SUSPENDED/STOPPED/UNDEPLOYED) |
| `x` | suspend (ACTIVE only) |
| `d` | terminate (hard, with confirmation) |

**VM state filters** (VMs tab)

| Key | Filter |
| --- | ------ |
| `a` | all VMs |
| `i` | active only |
| `o` | stopped only |
| `p` | poweroff only |
| `e` | error only |

**Host actions** (Hosts tab)

| Key | Action |
| --- | ------ |
| `e` | enable host |
| `d` | disable host |
| `o` | offline host (confirmation required) |
| `x` | delete host (type `yes` to confirm) |
| `n` | rename host (type new name) |

## Stack

- **Language:** Go 1.24+
- **TUI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **API:** [GOCA](https://github.com/OpenNebula/one/src/oca/go/src/goca) (Official OpenNebula Go Cloud API)

## Architecture

```
.
├── cmd/one9s/                  # Entry point: config + Bubble Tea startup
├── internal/
│   ├── config/                 # Auth loading (ONE_AUTH, ONE_XMLRPC, ~/.one/one_auth)
│   └── client/                 # GOCA wrapper + state mappers + ACL decoders
├── pkg/tui/
│   ├── viewport.go             # Root model: routing, modals, key dispatch
│   ├── vm_list.go              # VM table with state filters
│   ├── host_list.go            # Host table with actions
│   ├── datastore_list.go       # Datastore table
│   ├── acl_list.go             # ACL rules with hex decoding
│   ├── quota_list.go           # User quotas
│   ├── modal.go                # Confirmation modals (y/n + text input)
│   ├── styles.go               # Lip Gloss theme (OpenNebula blue palette)
│   └── messages.go             # Custom tea.Msg types
├── docs/                       # Website (one9s.scszero.com)
├── Makefile                    # build, install, run, test, lint
└── go.mod
```

## Requirements

- Go 1.24+ (to build from source).
- OpenNebula endpoint with XML-RPC API enabled (default port 2633).
- Credentials via `ONE_AUTH` or `~/.one/one_auth`.

## Installation

### Pre-built binaries

Download from [GitHub Releases](https://github.com/SergioZ3R0/one9s/releases):

```bash
curl -LO https://github.com/SergioZ3R0/one9s/releases/latest/download/one9s-linux-amd64.zip
unzip one9s-linux-amd64.zip
chmod +x one9s
```

### From source

```bash
git clone https://github.com/SergioZ3R0/one9s.git
cd one9s
make build
```

## Configuration

`one9s` is configured through environment variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `ONE_XMLRPC` | No | OpenNebula XML-RPC endpoint. Defaults to `http://localhost:2633/RPC2`. |
| `ONE_AUTH` | Yes* | Credentials as `user:password` or path to auth file. |

\* Falls back to `~/.one/one_auth` if `ONE_AUTH` is not set.

```bash
ONE_AUTH="oneadmin:password" \
ONE_XMLRPC="http://opennebula:2633/RPC2" \
./one9s
```

## License

[Apache License 2.0](LICENSE).
