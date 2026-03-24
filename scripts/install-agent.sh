#!/usr/bin/env bash
# ============================================================================
# Fluent Manager Agent Installer
# Enterprise Linux installer for fluent-manager-agent with optional
# Fluent Bit / Fluentd installation.
#
# Usage:
#   Local binary:   ./install-agent.sh --binary ./fluent-manager-agent
#   Remote download: ./install-agent.sh --download-url https://example.com/fluent-manager-agent
#   Non-interactive: ./install-agent.sh --binary ./fluent-manager-agent \
#                      --server-url https://fm.example.com --api-key SECRET --yes
# ============================================================================
set -euo pipefail

# ----- Defaults -----
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/fluent-manager"
CONFIG_FILE="${CONFIG_DIR}/agent.yaml"
SERVICE_NAME="fluent-manager-agent"
AGENT_USER="fluent-manager"
BINARY_PATH=""
DOWNLOAD_URL=""
SERVER_URL=""
API_KEY=""
NODE_UID=""
FLUENT_TYPE=""
YES=false
SKIP_FLUENT=false

# ----- Colors -----
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }
log_step()  { echo -e "\n${CYAN}${BOLD}>>> $*${NC}"; }

# ============================================================================
# CLI argument parsing
# ============================================================================
usage() {
    cat <<EOF
${BOLD}Fluent Manager Agent Installer${NC}

Usage: $0 [OPTIONS]

Binary source (one required):
  --binary PATH          Path to a local pre-built agent binary
  --download-url URL     URL to download the agent binary from

Configuration:
  --server-url URL       Fluent Manager server URL (required)
  --api-key KEY          Agent API key (required)
  --node-uid UID         Optional stable node UID
  --fluent-type TYPE     fluentbit | fluentd | auto (default: prompt)
  --install-dir DIR      Binary install directory (default: /usr/local/bin)
  --config-dir DIR       Config directory (default: /etc/fluent-manager)

Automation:
  --yes                  Non-interactive mode, accept all defaults
  --skip-fluent          Skip Fluent Bit / Fluentd installation prompt

Examples:
  # Interactive install from local binary
  sudo $0 --binary ./fluent-manager-agent

  # Fully automated remote install
  sudo $0 --download-url https://releases.example.com/agent-linux-amd64 \\
           --server-url https://fm.example.com --api-key mykey --yes
EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --binary)       BINARY_PATH="$2"; shift 2 ;;
        --download-url) DOWNLOAD_URL="$2"; shift 2 ;;
        --server-url)   SERVER_URL="$2"; shift 2 ;;
        --api-key)      API_KEY="$2"; shift 2 ;;
        --node-uid)     NODE_UID="$2"; shift 2 ;;
        --fluent-type)  FLUENT_TYPE="$2"; shift 2 ;;
        --install-dir)  INSTALL_DIR="$2"; shift 2 ;;
        --config-dir)   CONFIG_DIR="$2"; CONFIG_FILE="${CONFIG_DIR}/agent.yaml"; shift 2 ;;
        --yes|-y)       YES=true; shift ;;
        --skip-fluent)  SKIP_FLUENT=true; shift ;;
        --help|-h)      usage ;;
        *) log_error "Unknown option: $1"; usage ;;
    esac
done

# ============================================================================
# Pre-flight checks
# ============================================================================
log_step "Pre-flight checks"

if [[ $EUID -ne 0 ]]; then
    log_error "This script must be run as root (use sudo)."
    exit 1
fi

if [[ -z "$BINARY_PATH" && -z "$DOWNLOAD_URL" ]]; then
    log_error "Either --binary or --download-url is required."
    echo "Run '$0 --help' for usage."
    exit 1
fi

# Detect OS
if [[ ! -f /etc/os-release ]]; then
    log_error "Cannot detect Linux distribution (/etc/os-release not found)."
    exit 1
fi
. /etc/os-release
OS_ID="${ID:-unknown}"
OS_VERSION="${VERSION_ID:-unknown}"
log_info "Detected OS: ${PRETTY_NAME:-$OS_ID $OS_VERSION}"

# Detect arch
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  ARCH_LABEL="amd64" ;;
    aarch64) ARCH_LABEL="arm64" ;;
    *)       ARCH_LABEL="$ARCH" ;;
esac
log_info "Architecture: $ARCH ($ARCH_LABEL)"

# Detect package manager
PKG_MGR=""
if command -v apt-get &>/dev/null; then
    PKG_MGR="apt"
elif command -v yum &>/dev/null; then
    PKG_MGR="yum"
elif command -v dnf &>/dev/null; then
    PKG_MGR="dnf"
elif command -v zypper &>/dev/null; then
    PKG_MGR="zypper"
fi
log_info "Package manager: ${PKG_MGR:-none detected}"

# Detect init system
INIT_SYSTEM=""
if command -v systemctl &>/dev/null && systemctl --version &>/dev/null 2>&1; then
    INIT_SYSTEM="systemd"
elif [[ -d /etc/init.d ]]; then
    INIT_SYSTEM="sysvinit"
fi
log_info "Init system: ${INIT_SYSTEM:-unknown}"

# ============================================================================
# Helper: prompt with default
# ============================================================================
prompt_input() {
    local prompt_text="$1"
    local default_val="$2"
    local var_name="$3"

    if $YES; then
        eval "$var_name=\"$default_val\""
        return
    fi

    local display_default=""
    if [[ -n "$default_val" ]]; then
        display_default=" [${default_val}]"
    fi

    read -rp "$(echo -e "${BOLD}${prompt_text}${display_default}: ${NC}")" input
    eval "$var_name=\"${input:-$default_val}\""
}

prompt_confirm() {
    local prompt_text="$1"
    local default_yes="${2:-true}"

    if $YES; then
        return 0
    fi

    local yn_hint="Y/n"
    $default_yes || yn_hint="y/N"

    read -rp "$(echo -e "${BOLD}${prompt_text} [${yn_hint}]: ${NC}")" answer
    case "${answer,,}" in
        y|yes) return 0 ;;
        n|no)  return 1 ;;
        "")    $default_yes && return 0 || return 1 ;;
        *)     $default_yes && return 0 || return 1 ;;
    esac
}

# ============================================================================
# Step 1: Collect required configuration
# ============================================================================
log_step "Agent configuration"

if [[ -z "$SERVER_URL" ]]; then
    prompt_input "Fluent Manager server URL (e.g. https://fm.example.com)" "" SERVER_URL
fi
if [[ -z "$SERVER_URL" ]]; then
    log_error "server_url is required."
    exit 1
fi

if [[ -z "$API_KEY" ]]; then
    prompt_input "Agent API key" "" API_KEY
fi
if [[ -z "$API_KEY" ]]; then
    log_error "api_key is required."
    exit 1
fi

if [[ -z "$NODE_UID" ]]; then
    prompt_input "Node UID (leave empty for auto-generate)" "" NODE_UID
fi

log_info "Server URL: $SERVER_URL"
log_info "Node UID:   ${NODE_UID:-<auto-generate>}"

# ============================================================================
# Step 2: Detect or install Fluent Bit / Fluentd
# ============================================================================
log_step "Fluent runtime detection"

FLUENTBIT_INSTALLED=false
FLUENTD_INSTALLED=false

if command -v fluent-bit &>/dev/null || [[ -x /opt/fluent-bit/bin/fluent-bit ]] || [[ -x /usr/bin/fluent-bit ]]; then
    FLUENTBIT_INSTALLED=true
    FB_VER=$(fluent-bit --version 2>/dev/null | head -1 || echo "unknown")
    log_info "Fluent Bit detected: $FB_VER"
fi

if command -v fluentd &>/dev/null || command -v td-agent &>/dev/null; then
    FLUENTD_INSTALLED=true
    FD_VER=$(fluentd --version 2>/dev/null || td-agent --version 2>/dev/null || echo "unknown")
    log_info "Fluentd detected: $FD_VER"
fi

install_fluentbit() {
    log_step "Installing Fluent Bit"
    case "$PKG_MGR" in
        apt)
            log_info "Adding Fluent Bit APT repository..."
            curl -fsSL https://packages.fluentbit.io/fluentbit.key | gpg --dearmor -o /usr/share/keyrings/fluentbit-keyring.gpg
            echo "deb [signed-by=/usr/share/keyrings/fluentbit-keyring.gpg] https://packages.fluentbit.io/${OS_ID}/${OS_VERSION} ${VERSION_CODENAME:-stable} main" \
                > /etc/apt/sources.list.d/fluent-bit.list
            apt-get update -qq
            apt-get install -y fluent-bit
            ;;
        yum|dnf)
            log_info "Adding Fluent Bit YUM/DNF repository..."
            cat > /etc/yum.repos.d/fluent-bit.repo <<REPO
[fluent-bit]
name=Fluent Bit
baseurl=https://packages.fluentbit.io/centos/\$releasever/\$basearch/
gpgcheck=1
gpgkey=https://packages.fluentbit.io/fluentbit.key
enabled=1
REPO
            $PKG_MGR install -y fluent-bit
            ;;
        zypper)
            log_info "Adding Fluent Bit Zypper repository..."
            rpm --import https://packages.fluentbit.io/fluentbit.key
            zypper addrepo https://packages.fluentbit.io/sles/ fluent-bit
            zypper --non-interactive install fluent-bit
            ;;
        *)
            log_warn "Unsupported package manager. Please install Fluent Bit manually:"
            log_warn "  https://docs.fluentbit.io/manual/installation/linux"
            return 1
            ;;
    esac

    if [[ "$INIT_SYSTEM" == "systemd" ]]; then
        systemctl enable fluent-bit
        systemctl start fluent-bit
        log_info "Fluent Bit service enabled and started."
    fi
    FLUENTBIT_INSTALLED=true
    log_info "Fluent Bit installed successfully."
}

install_fluentd() {
    log_step "Installing Fluentd (td-agent)"
    case "$PKG_MGR" in
        apt)
            log_info "Installing Fluentd via install script..."
            curl -fsSL https://toolbelt.treasuredata.com/sh/install-${OS_ID}-${VERSION_CODENAME:-stable}-td-agent5.sh | sh
            ;;
        yum|dnf)
            log_info "Installing Fluentd via install script..."
            curl -fsSL https://toolbelt.treasuredata.com/sh/install-redhat-td-agent4.sh | sh
            ;;
        *)
            log_warn "Unsupported package manager. Please install Fluentd manually:"
            log_warn "  https://docs.fluentd.org/installation"
            return 1
            ;;
    esac

    if [[ "$INIT_SYSTEM" == "systemd" ]]; then
        systemctl enable td-agent || systemctl enable fluentd || true
        systemctl start td-agent || systemctl start fluentd || true
        log_info "Fluentd service enabled and started."
    fi
    FLUENTD_INSTALLED=true
    log_info "Fluentd installed successfully."
}

if ! $FLUENTBIT_INSTALLED && ! $FLUENTD_INSTALLED && ! $SKIP_FLUENT; then
    log_warn "No Fluent Bit or Fluentd runtime detected on this system."
    echo ""
    echo -e "  The agent requires a log collector to manage. Please choose:"
    echo -e "    ${BOLD}1)${NC} Install Fluent Bit  ${GREEN}(recommended, lightweight C-based)${NC}"
    echo -e "    ${BOLD}2)${NC} Install Fluentd     ${YELLOW}(Ruby-based, plugin-rich)${NC}"
    echo -e "    ${BOLD}3)${NC} Skip — I will install it separately"
    echo ""

    if $YES; then
        FLUENT_CHOICE="1"
        log_info "Non-interactive mode: defaulting to Fluent Bit."
    else
        read -rp "$(echo -e "${BOLD}Your choice [1/2/3]: ${NC}")" FLUENT_CHOICE
    fi

    case "${FLUENT_CHOICE}" in
        1) install_fluentbit ;;
        2) install_fluentd ;;
        3) log_info "Skipping log collector installation." ;;
        *) log_warn "Invalid choice, skipping." ;;
    esac
elif $FLUENTBIT_INSTALLED && $FLUENTD_INSTALLED; then
    log_info "Both Fluent Bit and Fluentd are installed."
elif $FLUENTBIT_INSTALLED; then
    log_info "Fluent Bit is available."
elif $FLUENTD_INSTALLED; then
    log_info "Fluentd is available."
fi

# Determine fluent_type
if [[ -z "$FLUENT_TYPE" ]]; then
    if $FLUENTBIT_INSTALLED && ! $FLUENTD_INSTALLED; then
        FLUENT_TYPE="fluentbit"
    elif $FLUENTD_INSTALLED && ! $FLUENTBIT_INSTALLED; then
        FLUENT_TYPE="fluentd"
    elif $FLUENTBIT_INSTALLED && $FLUENTD_INSTALLED; then
        if $YES; then
            FLUENT_TYPE="fluentbit"
        else
            echo ""
            echo -e "  Both runtimes detected. Which should the agent manage?"
            echo -e "    ${BOLD}1)${NC} Fluent Bit"
            echo -e "    ${BOLD}2)${NC} Fluentd"
            read -rp "$(echo -e "${BOLD}Your choice [1/2]: ${NC}")" TYPE_CHOICE
            case "$TYPE_CHOICE" in
                2) FLUENT_TYPE="fluentd" ;;
                *) FLUENT_TYPE="fluentbit" ;;
            esac
        fi
    else
        FLUENT_TYPE="auto"
    fi
fi
log_info "Fluent type: $FLUENT_TYPE"

# ============================================================================
# Step 3: Install agent binary
# ============================================================================
log_step "Installing agent binary"

STAGING_BIN=$(mktemp)
trap "rm -f '$STAGING_BIN'" EXIT

if [[ -n "$BINARY_PATH" ]]; then
    if [[ ! -f "$BINARY_PATH" ]]; then
        log_error "Binary not found: $BINARY_PATH"
        exit 1
    fi
    cp "$BINARY_PATH" "$STAGING_BIN"
    log_info "Using local binary: $BINARY_PATH"
elif [[ -n "$DOWNLOAD_URL" ]]; then
    log_info "Downloading agent from: $DOWNLOAD_URL"
    if command -v curl &>/dev/null; then
        curl -fSL --progress-bar -o "$STAGING_BIN" "$DOWNLOAD_URL"
    elif command -v wget &>/dev/null; then
        wget -q --show-progress -O "$STAGING_BIN" "$DOWNLOAD_URL"
    else
        log_error "Neither curl nor wget found. Cannot download binary."
        exit 1
    fi
fi

# Validate it's an ELF binary
if ! file "$STAGING_BIN" | grep -q "ELF"; then
    log_error "Downloaded file is not a valid Linux ELF binary."
    exit 1
fi

chmod 0755 "$STAGING_BIN"
mv "$STAGING_BIN" "${INSTALL_DIR}/${SERVICE_NAME}"
trap - EXIT
log_info "Installed: ${INSTALL_DIR}/${SERVICE_NAME}"

# Verify
INSTALLED_VER=$("${INSTALL_DIR}/${SERVICE_NAME}" --version 2>&1 || true)
log_info "Agent version: ${INSTALLED_VER:-unknown}"

# ============================================================================
# Step 4: Create system user
# ============================================================================
log_step "System user setup"

if id "$AGENT_USER" &>/dev/null; then
    log_info "User '$AGENT_USER' already exists."
else
    useradd --system --no-create-home --shell /usr/sbin/nologin "$AGENT_USER"
    log_info "Created system user: $AGENT_USER"
fi

# ============================================================================
# Step 5: Write configuration
# ============================================================================
log_step "Writing agent configuration"

mkdir -p "$CONFIG_DIR"

if [[ -f "$CONFIG_FILE" ]]; then
    BACKUP="${CONFIG_FILE}.bak.$(date +%Y%m%d%H%M%S)"
    cp "$CONFIG_FILE" "$BACKUP"
    log_warn "Existing config backed up to: $BACKUP"
fi

cat > "$CONFIG_FILE" <<YAML
# Fluent Manager Agent Configuration
# Generated by install-agent.sh on $(date -u +"%Y-%m-%dT%H:%M:%SZ")

server_url: "${SERVER_URL}"
api_key: "${API_KEY}"
YAML

if [[ -n "$NODE_UID" ]]; then
    echo "node_uid: \"${NODE_UID}\"" >> "$CONFIG_FILE"
fi

cat >> "$CONFIG_FILE" <<YAML

# Fluent runtime (auto-detected paths can be overridden here)
fluent_type: "${FLUENT_TYPE}"
# fluent_config_path: ""
# fluent_config_dir: ""
# fluent_binary: ""
# fluent_service_unit: ""
# fluent_log_path: ""

# Intervals (seconds)
# heartbeat_interval: 30
# metrics_interval: 60
# log_upload_interval: 120

# Local health endpoint
# health_port: 9880

# Labels (JSON key-value for match rules)
# labels: '{"env":"production","team":"infra"}'
YAML

chmod 0640 "$CONFIG_FILE"
chown root:"$AGENT_USER" "$CONFIG_FILE"
log_info "Config written: $CONFIG_FILE"

# State directory
STATE_DIR="/var/lib/fluent-manager-agent"
mkdir -p "$STATE_DIR"
chown "$AGENT_USER":"$AGENT_USER" "$STATE_DIR"

# ============================================================================
# Step 6: Install systemd service (or init script)
# ============================================================================
log_step "Service installation"

if [[ "$INIT_SYSTEM" == "systemd" ]]; then
    UNIT_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

    cat > "$UNIT_FILE" <<UNIT
[Unit]
Description=Fluent Manager Agent
Documentation=https://github.com/fluent-manager/fluent-manager
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${AGENT_USER}
Group=${AGENT_USER}
ExecStart=${INSTALL_DIR}/${SERVICE_NAME} --config ${CONFIG_FILE}
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${SERVICE_NAME}

# Hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${STATE_DIR} /var/log
PrivateTmp=true
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
UNIT

    systemctl daemon-reload
    systemctl enable "${SERVICE_NAME}"
    log_info "Systemd unit installed: $UNIT_FILE"

elif [[ "$INIT_SYSTEM" == "sysvinit" ]]; then
    INIT_SCRIPT="/etc/init.d/${SERVICE_NAME}"

    cat > "$INIT_SCRIPT" <<'INITSCRIPT'
#!/bin/sh
### BEGIN INIT INFO
# Provides:          fluent-manager-agent
# Required-Start:    $network $remote_fs
# Required-Stop:     $network $remote_fs
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Description:       Fluent Manager Agent
### END INIT INFO

DAEMON=INSTALL_DIR_PLACEHOLDER/fluent-manager-agent
DAEMON_ARGS="--config CONFIG_FILE_PLACEHOLDER"
PIDFILE=/var/run/fluent-manager-agent.pid
DAEMON_USER=AGENT_USER_PLACEHOLDER

case "$1" in
    start)
        echo "Starting fluent-manager-agent..."
        start-stop-daemon --start --quiet --background --make-pidfile \
            --pidfile "$PIDFILE" --chuid "$DAEMON_USER" --exec "$DAEMON" -- $DAEMON_ARGS
        ;;
    stop)
        echo "Stopping fluent-manager-agent..."
        start-stop-daemon --stop --quiet --pidfile "$PIDFILE" --retry=TERM/30/KILL/5
        rm -f "$PIDFILE"
        ;;
    restart)
        $0 stop
        $0 start
        ;;
    status)
        if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
            echo "fluent-manager-agent is running (PID $(cat "$PIDFILE"))"
        else
            echo "fluent-manager-agent is not running"
            exit 1
        fi
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac
INITSCRIPT

    sed -i "s|INSTALL_DIR_PLACEHOLDER|${INSTALL_DIR}|g" "$INIT_SCRIPT"
    sed -i "s|CONFIG_FILE_PLACEHOLDER|${CONFIG_FILE}|g" "$INIT_SCRIPT"
    sed -i "s|AGENT_USER_PLACEHOLDER|${AGENT_USER}|g" "$INIT_SCRIPT"
    chmod 0755 "$INIT_SCRIPT"
    update-rc.d "${SERVICE_NAME}" defaults 2>/dev/null || chkconfig "${SERVICE_NAME}" on 2>/dev/null || true
    log_info "SysVinit script installed: $INIT_SCRIPT"
else
    log_warn "No supported init system detected. You will need to start the agent manually:"
    log_warn "  ${INSTALL_DIR}/${SERVICE_NAME} --config ${CONFIG_FILE}"
fi

# ============================================================================
# Step 7: Start the agent
# ============================================================================
log_step "Starting agent"

if prompt_confirm "Start fluent-manager-agent now?" true; then
    if [[ "$INIT_SYSTEM" == "systemd" ]]; then
        systemctl start "${SERVICE_NAME}"
        sleep 2
        if systemctl is-active --quiet "${SERVICE_NAME}"; then
            log_info "Agent is running."
            systemctl status "${SERVICE_NAME}" --no-pager --lines=5
        else
            log_error "Agent failed to start. Check logs:"
            log_error "  journalctl -u ${SERVICE_NAME} -n 50 --no-pager"
            exit 1
        fi
    elif [[ "$INIT_SYSTEM" == "sysvinit" ]]; then
        /etc/init.d/${SERVICE_NAME} start
        log_info "Agent start command issued."
    fi
else
    log_info "Skipped. Start manually with:"
    if [[ "$INIT_SYSTEM" == "systemd" ]]; then
        echo "  systemctl start ${SERVICE_NAME}"
    else
        echo "  ${INSTALL_DIR}/${SERVICE_NAME} --config ${CONFIG_FILE}"
    fi
fi

# ============================================================================
# Summary
# ============================================================================
log_step "Installation complete"

echo ""
echo -e "  ${BOLD}Binary:${NC}       ${INSTALL_DIR}/${SERVICE_NAME}"
echo -e "  ${BOLD}Config:${NC}       ${CONFIG_FILE}"
echo -e "  ${BOLD}State dir:${NC}    ${STATE_DIR}"
echo -e "  ${BOLD}Service:${NC}      ${SERVICE_NAME}"
echo -e "  ${BOLD}User:${NC}         ${AGENT_USER}"
echo -e "  ${BOLD}Fluent type:${NC}  ${FLUENT_TYPE}"
echo -e "  ${BOLD}Server:${NC}       ${SERVER_URL}"
echo ""
echo -e "  ${BOLD}Useful commands:${NC}"
if [[ "$INIT_SYSTEM" == "systemd" ]]; then
    echo "    systemctl status  ${SERVICE_NAME}"
    echo "    systemctl restart ${SERVICE_NAME}"
    echo "    journalctl -u ${SERVICE_NAME} -f"
fi
echo "    ${INSTALL_DIR}/${SERVICE_NAME} --version"
echo ""
log_info "Done."
