# Fluent Manager Agent - Ansible Role

Ansible role for batch deployment of Fluent Manager Agent, with optional Fluent Bit / Fluentd installation.

## Quick Start

```bash
# 1. Put your pre-built agent binary here
cp /path/to/fluent-manager-agent scripts/ansible/files/

# 2. Copy and edit inventory
cp inventory.example.ini inventory.ini
vim inventory.ini

# 3. Run
ansible-playbook -i inventory.ini playbook.example.yml
```

## Role Structure

```
roles/fluent_manager_agent/
├── defaults/main.yml     # All configurable variables with defaults
├── handlers/main.yml     # Service restart/reload handlers
├── meta/main.yml         # Role metadata (supported platforms)
├── tasks/
│   ├── main.yml              # Entry point + validation
│   ├── install_agent.yml     # Binary install (copy or download)
│   ├── install_fluentbit.yml # Optional Fluent Bit from official repo
│   ├── install_fluentd.yml   # Optional Fluentd (td-agent)
│   ├── configure.yml         # agent.yaml template
│   └── service.yml           # Systemd unit setup
└── templates/
    ├── agent.yaml.j2                    # Agent config template
    └── fluent-manager-agent.service.j2  # Systemd unit template
```

## Required Variables

| Variable | Description |
|----------|-------------|
| `fm_server_url` | Fluent Manager server URL |
| `fm_api_key` | Agent API key (use ansible-vault) |
| `fm_agent_binary` or `fm_agent_download_url` | Agent binary source |

## Key Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `fm_fluent_type` | `auto` | `auto`, `fluentbit`, `fluentd` |
| `fm_install_fluentbit` | `false` | Install Fluent Bit from official repo |
| `fm_install_fluentd` | `false` | Install Fluentd (td-agent) |
| `fm_labels` | `""` | JSON labels for match rules |
| `fm_node_uid` | `""` | Stable node UID (auto-generated if empty) |
| `fm_agent_log_groups` | `["adm"]` / `["systemd-journal"]` | Supplementary OS groups for reading system logs. Auto-selected by OS family (Debian→`adm`, RedHat→`systemd-journal`). Set to `[]` to disable. |
| `fm_extra_config` | `""` | Raw YAML appended to agent.yaml |

See [defaults/main.yml](roles/fluent_manager_agent/defaults/main.yml) for all variables.

## Usage in Existing Playbooks

```yaml
# In your existing playbook, just add the role:
- hosts: app_servers
  become: true
  roles:
    - role: /path/to/fluent_manager_agent
      vars:
        fm_server_url: "https://fm.example.com"
        fm_api_key: "{{ vault_fm_api_key }}"
        fm_agent_binary: files/fluent-manager-agent
        fm_install_fluentbit: true
```

Or add to `requirements.yml` if hosting internally:

```yaml
# requirements.yml
- src: git+https://git.example.com/infra/fluent-manager.git
  scm: git
  version: main
  name: fluent_manager_agent
```

```bash
ansible-galaxy install -r requirements.yml -p roles/
```

## API Key Security

Store `fm_api_key` in Ansible Vault:

```bash
# Create vault file
ansible-vault create group_vars/all/vault.yml
# Add: vault_fm_api_key: "your-key-here"

# Reference in playbook
fm_api_key: "{{ vault_fm_api_key }}"

# Run with vault
ansible-playbook -i inventory.ini playbook.yml --ask-vault-pass
```

## Per-Host Customization

```ini
# inventory.ini
[app_servers]
node-01  ansible_host=10.0.1.11  fm_labels='{"team":"api","env":"prod"}'
node-02  ansible_host=10.0.1.12  fm_labels='{"team":"web","env":"prod"}'
node-03  ansible_host=10.0.1.13  fm_node_uid=stable-uid-for-node-03
```

## Supported Platforms

- Ubuntu 20.04 / 22.04 / 24.04
- Debian 11 / 12
- CentOS / RHEL 7 / 8 / 9
- Amazon Linux 2 / 2023
