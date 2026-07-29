#!/usr/bin/env bash
# shellcheck disable=SC2317  # trap EXIT/ERR not directly invoked
set -Eeuo pipefail

# ---------------------------------------------------------------------------
# create-cloud-init.sh — Create a cloud-init FAT disk image for cloud-hypervisor
#
# Options:
#   -o, --output FILE       Output image path (default: /tmp/ubuntu-cloudinit.img)
#   -u, --user-data FILE    User-data file (default: ./cloud-init/user-data)
#   -m, --meta-data FILE    Meta-data file (default: ./cloud-init/meta-data)
#   -n, --network-config FILE Network-config file (default: ./cloud-init/network-config)
#   -h, --help              Show this help message
# ---------------------------------------------------------------------------

# Resolve script directory symlink-safe
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"

# ---- Defaults ----
OUTPUT_FILE="/tmp/ubuntu-cloudinit.img"
USER_DATA="${SCRIPT_DIR}/cloud-init/user-data"
META_DATA="${SCRIPT_DIR}/cloud-init/meta-data"
NETWORK_CONFIG="${SCRIPT_DIR}/cloud-init/network-config"

# ---- Cleanup handler ----
_CLEANUP_TMP=""

_cleanup() {
    local exit_code=$?
    if [[ -n "$_CLEANUP_TMP" && -f "$_CLEANUP_TMP" ]]; then
        rm -f -- "$_CLEANUP_TMP"
    fi
    trap - EXIT ERR
    exit "$exit_code"
}

_error_handler() {
    local line=$1
    local cmd=$2
    printf 'ERROR: Command failed at line %d: %s\n' "$line" "$cmd" >&2
}

trap _cleanup EXIT
trap '_error_handler $LINENO "$BASH_COMMAND"' ERR

# ---- Usage ----
usage() {
    local exit_code="${1:-0}"
    cat <<EOF
Usage: ${0##*/} [OPTIONS]

Create a cloud-init FAT disk image for use with cloud-hypervisor.

Options:
  -o, --output FILE         Output image path (default: /tmp/ubuntu-cloudinit.img)
  -u, --user-data FILE      User-data file (default: ./cloud-init/user-data)
  -m, --meta-data FILE      Meta-data file (default: ./cloud-init/meta-data)
  -n, --network-config FILE Network-config file (default: ./cloud-init/network-config)
  -h, --help                Show this help message
EOF
    exit "$exit_code"
}

# ---- Dependency check ----
_check_deps() {
    local missing=()
    local cmd
    for cmd in mkdosfs mcopy; do
        if ! command -v "$cmd" &>/dev/null; then
            missing+=("$cmd")
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        {
            printf 'ERROR: Missing required tools: %s\n' "${missing[*]}"
            printf 'Install them with:\n'
            for cmd in "${missing[@]}"; do
                case "$cmd" in
                    mkdosfs) printf '  sudo apt install dosfstools   # Debian/Ubuntu\n' ;;
                    mcopy)   printf '  sudo apt install mtools       # Debian/Ubuntu\n' ;;
                esac
            done
        } >&2
        exit 1
    fi
}

# ---- Parse arguments ----
while [[ $# -gt 0 ]]; do
    case "$1" in
        -o|--output)
            OUTPUT_FILE="${2:?--output requires a value}"
            shift 2
            ;;
        -o=*|--output=*)
            OUTPUT_FILE="${1#*=}"
            shift
            ;;
        -u|--user-data)
            USER_DATA="${2:?--user-data requires a value}"
            shift 2
            ;;
        -u=*|--user-data=*)
            USER_DATA="${1#*=}"
            shift
            ;;
        -m|--meta-data)
            META_DATA="${2:?--meta-data requires a value}"
            shift 2
            ;;
        -m=*|--meta-data=*)
            META_DATA="${1#*=}"
            shift
            ;;
        -n|--network-config)
            NETWORK_CONFIG="${2:?--network-config requires a value}"
            shift 2
            ;;
        -n=*|--network-config=*)
            NETWORK_CONFIG="${1#*=}"
            shift
            ;;
        -h|--help)
            usage
            ;;
        --)
            shift
            break
            ;;
        -*)
            printf 'ERROR: Unknown option: %s\n' "$1" >&2
            usage 1
            ;;
        *)
            printf 'ERROR: Unexpected argument: %s\n' "$1" >&2
            usage 1
            ;;
    esac
done

# ---- Validate input files ----
for fvar in USER_DATA META_DATA NETWORK_CONFIG; do
    path="${!fvar}"
    if [[ ! -f "$path" ]]; then
        printf 'ERROR: File not found: %s (%s)\n' "$path" "$fvar" >&2
        exit 1
    fi
done

# ---- Create FAT image ----
printf 'Creating cloud-init disk image ...\n'

# Remove existing output to avoid mkdosfs interactive prompt
rm -f -- "$OUTPUT_FILE"

_check_deps
mkdosfs -n CIDATA -C "$OUTPUT_FILE" 8192

mcopy -oi "$OUTPUT_FILE" -s "$USER_DATA" ::
mcopy -oi "$OUTPUT_FILE" -s "$META_DATA" ::
mcopy -oi "$OUTPUT_FILE" -s "$NETWORK_CONFIG" ::

printf 'Done. Cloud-init disk image created: %s\n' "$OUTPUT_FILE"

# Verify content
printf '\nContents of %s:\n' "$OUTPUT_FILE"
mdir -i "$OUTPUT_FILE" ::
