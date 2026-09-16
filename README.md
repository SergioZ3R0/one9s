<div align="center">
  <img src="https://raw.githubusercontent.com/SergioZ3R0/one9s/main/docs/one9s.svg" alt="one9s" width="180">
</div>

<h1 align="center">one9s</h1>
<p align="center">
  <b>Terminal UI for OpenNebula cluster management</b><br>
  <i>Monitor VMs, Hosts, Datastores, Manage ACLs and Quotas from your terminal</i><br>
  <a href="https://one9s.scszero.com/">one9s.scszero.com</a> &nbsp;|&nbsp;
  <a href="https://one9s.scszero.com/docs.html">Documentation</a>
</p>

`one9s` is a **TUI** (terminal user interface) for managing and monitoring
**OpenNebula** clusters directly from your terminal, inspired by
[k9s](https://github.com/derailed/k9s) and
[lazygit](https://github.com/jesseduffield/lazygit).

No Sunstone web UI. No Ruby CLI dependencies. Just a single binary that talks
to the OpenNebula XML-RPC API.

## Features

- **VM Management** — monitor all VMs with state, CPU, memory, IP, host. Filter by state (active, stopped, poweroff, error) or fuzzy text search.
- **VM Lifecycle** — reboot, stop, resume, suspend, terminate directly from the terminal with confirmation modals.
- **Host Management** — enable, disable, offline, delete, rename hosts with confirmation.
- **Datastore Monitoring** — view storage pools with type, total capacity, used and free space.
- **ACL Rules** — view and decode access control rules (user, resource, rights, zone) from hex to human-readable.
- **User Quotas** — view and edit per-user quotas for VMs, CPU, Memory, Running VMs, Images, Size and Leases.
- **Vault Encryption** — credentials encrypted with AES-256-GCM, safe from third-party access.
- **Cross-platform** — single binary for Linux, macOS, Windows (amd64/arm64).

## Quick Start

### Download

```bash
curl -LO https://github.com/SergioZ3R0/one9s/releases/latest/download/one9s-linux-amd64.zip
unzip one9s-linux-amd64.zip
chmod +x one9s
```

### Configure

```bash
mkdir -p ~/.one9s
cat > ~/.one9s/config << 'EOF'
ONE_XMLRPC=http://opennebula:2633/RPC2
ONE_AUTH=oneadmin:password
EOF
chmod 600 ~/.one9s/config
```

Or as a one-liner:

```bash
ONE_AUTH="oneadmin:password" ONE_XMLRPC="http://opennebula:2633/RPC2" ./one9s
```

### Encrypted vault (recommended)

Store credentials encrypted with AES-256-GCM:

```bash
# Create encrypted vault
one9s vault init

# Run (prompts for vault password)
./one9s

# Or skip vault with direct env vars (no password prompt)
ONE_AUTH="oneadmin:password" ONE_XMLRPC="http://opennebula:2633/RPC2" ./one9s
```

Vault commands:

| Command | Description |
|---------|-------------|
| `one9s vault init` | Create new encrypted config |
| `one9s vault encrypt` | Encrypt existing plain config |
| `one9s vault decrypt` | Decrypt and display contents |

Vault password can be set via `ONE_VAULT_PASS` env var to skip the prompt.

### Run

```bash
./one9s
```

## Key Bindings

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Next / previous tab |
| `↑/↓` / `j/k` | Navigate rows |
| `PgUp/PgDn` / `b/f` | Page up / page down |
| `g/G` | Go to start / end |
| `/` | Fuzzy search |
| `F5` | Refresh current view |
| `?` | Help |

### VM Actions (VMs tab)

| Key | Action | Valid States |
|-----|--------|--------------|
| `r` | Reboot | ACTIVE |
| `s` | Stop | ACTIVE |
| `u` | Resume/Start | POWEROFF, SUSPENDED, STOPPED, UNDEPLOYED |
| `x` | Suspend | ACTIVE |
| `d` | Terminate (hard) | Any (with confirmation) |

### VM State Filters (VMs tab)

| Key | Filter |
|-----|--------|
| `a` | All VMs |
| `i` | Active only |
| `o` | Stopped only |
| `p` | Poweroff only |
| `e` | Error only |

### Host Actions (Hosts tab)

| Key | Action |
|-----|--------|
| `e` | Enable host |
| `d` | Disable host |
| `o` | Offline host (confirmation required) |
| `x` | Delete host (type `yes` to confirm) |
| `n` | Rename host (type new name) |

### Quota Actions (Quotas tab)

| Key | Action |
|-----|--------|
| `e` | Edit user quota (form modal with VMs, CPU, Memory, etc.) |

## Permissions

one9s requires appropriate OpenNebula permissions. If an action fails due to insufficient permissions, you'll see a clear message like "permission denied: requires VM:MANAGE".

| Action | Required Permission |
|--------|-------------------|
| View VMs, Hosts, Datastores, ACLs | VM:USE, HOST:USE, DS:USE, ACL:USE |
| Reboot, Stop, Suspend, Resume VM | VM:MANAGE |
| Terminate VM | VM:ADMIN |
| Enable/Disable/Offline Host | HOST:ADMIN |
| Delete/Rename Host | HOST:ADMIN |
| Edit Quotas | Quota:ADMIN |

## Stack

- **Language:** Go 1.24+
- **TUI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **API:** [GOCA](https://github.com/OpenNebula/one/src/oca/go/src/goca) (Official OpenNebula Go Cloud API)

## License

[Apache License 2.0](LICENSE)
