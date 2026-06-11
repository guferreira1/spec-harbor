#!/bin/sh
# SpecHarbor installer.
#
# Downloads an official SpecHarbor GitHub Release archive, verifies its
# SHA-256 checksum against the release checksums.txt, and installs the
# specharbor binary into a user-local directory.
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh
#
# Options (flags take precedence over environment variables):
#   --version <vX.Y.Z|X.Y.Z>   Install a specific release (default: latest).
#   --install-dir <dir>        Install target directory (default: $HOME/.local/bin).
#   --dry-run                  Print resolved values without downloading or writing.
#   -h, --help                 Show usage.
#
# Environment variables:
#   SPECHARBOR_VERSION         Same as --version.
#   SPECHARBOR_INSTALL_DIR     Same as --install-dir.
#   SPECHARBOR_DRY_RUN=1       Same as --dry-run.
#
# Safety:
#   - HTTPS only; downloads are restricted to
#     https://github.com/guferreira1/spec-harbor/releases/ URLs.
#   - SHA-256 verification is mandatory; missing checksum tooling fails the install.
#   - Never invokes sudo and never writes outside its temporary directory
#     and the install target.
#   - Never executes downloaded content.

set -eu

SPECHARBOR_REPO_URL="https://github.com/guferreira1/spec-harbor"
SPECHARBOR_RELEASES_URL="${SPECHARBOR_REPO_URL}/releases"
SPECHARBOR_BINARY_NAME="specharbor"
SPECHARBOR_CHECKSUMS_NAME="checksums.txt"
SPECHARBOR_INSTALL_DOCS="${SPECHARBOR_REPO_URL}/blob/main/docs/install.md"

SPECHARBOR_TMP_DIR=""
SPECHARBOR_PARTIAL_TARGET=""

sh_log() {
    printf '%s\n' "$1"
}

sh_warn() {
    printf 'warning: %s\n' "$1" >&2
}

sh_fail() {
    printf 'error: %s\n' "$1" >&2
    exit 1
}

cleanup() {
    if [ -n "${SPECHARBOR_TMP_DIR}" ] && [ -d "${SPECHARBOR_TMP_DIR}" ]; then
        rm -rf "${SPECHARBOR_TMP_DIR}"
    fi
    if [ -n "${SPECHARBOR_PARTIAL_TARGET}" ] && [ -f "${SPECHARBOR_PARTIAL_TARGET}" ]; then
        rm -f "${SPECHARBOR_PARTIAL_TARGET}"
    fi
}

usage() {
    sh_log "SpecHarbor installer"
    sh_log ""
    sh_log "Usage: install.sh [--version <vX.Y.Z>] [--install-dir <dir>] [--dry-run]"
    sh_log ""
    sh_log "Environment variables: SPECHARBOR_VERSION, SPECHARBOR_INSTALL_DIR, SPECHARBOR_DRY_RUN=1"
    sh_log "Documentation: ${SPECHARBOR_INSTALL_DOCS}"
}

detect_os() {
    case "$1" in
        Linux)
            printf 'Linux\n'
            ;;
        Darwin)
            printf 'Darwin\n'
            ;;
        *)
            printf 'error: unsupported operating system "%s". Supported: Linux, Darwin (macOS). For Windows and other platforms see %s\n' "$1" "${SPECHARBOR_INSTALL_DOCS}" >&2
            return 1
            ;;
    esac
}

detect_arch() {
    case "$1" in
        x86_64 | amd64)
            printf 'x86_64\n'
            ;;
        aarch64 | arm64)
            printf 'arm64\n'
            ;;
        *)
            printf 'error: unsupported architecture "%s". Supported: x86_64 (amd64), arm64 (aarch64). See %s\n' "$1" "${SPECHARBOR_INSTALL_DOCS}" >&2
            return 1
            ;;
    esac
}

is_valid_version() {
    printf '%s' "$1" | grep -Eq '^v?[0-9]+\.[0-9]+\.[0-9]+$'
}

normalize_tag() {
    case "$1" in
        v*)
            printf '%s\n' "$1"
            ;;
        *)
            printf 'v%s\n' "$1"
            ;;
    esac
}

asset_name() {
    # $1 = OS (Linux|Darwin), $2 = arch (x86_64|arm64)
    printf '%s_%s_%s.tar.gz\n' "${SPECHARBOR_BINARY_NAME}" "$1" "$2"
}

asset_url() {
    # $1 = tag, $2 = asset name
    printf '%s/download/%s/%s\n' "${SPECHARBOR_RELEASES_URL}" "$1" "$2"
}

checksums_url() {
    # $1 = tag
    printf '%s/download/%s/%s\n' "${SPECHARBOR_RELEASES_URL}" "$1" "${SPECHARBOR_CHECKSUMS_NAME}"
}

http_client() {
    if command -v curl >/dev/null 2>&1; then
        printf 'curl\n'
    elif command -v wget >/dev/null 2>&1; then
        printf 'wget\n'
    else
        printf 'error: this installer requires curl or wget.\n' >&2
        return 1
    fi
}

assert_allowed_url() {
    case "$1" in
        "${SPECHARBOR_RELEASES_URL}/"*) ;;
        *)
            sh_fail "refusing to download from non-allowlisted URL: $1"
            ;;
    esac
}

http_download() {
    # $1 = URL, $2 = destination file
    assert_allowed_url "$1"
    client="$(http_client)"
    if [ "${client}" = "curl" ]; then
        curl --proto '=https' --proto-redir '=https' -fsSL -o "$2" "$1" ||
            sh_fail "download failed: $1"
    else
        wget --https-only -q -O "$2" "$1" ||
            sh_fail "download failed: $1"
    fi
}

resolve_latest_tag() {
    latest_endpoint="${SPECHARBOR_RELEASES_URL}/latest"
    client="$(http_client)"
    if [ "${client}" = "curl" ]; then
        final_url="$(curl --proto '=https' --proto-redir '=https' -fsSL -o /dev/null -w '%{url_effective}' "${latest_endpoint}")" || {
            printf 'error: could not resolve the latest release from %s\n' "${latest_endpoint}" >&2
            return 1
        }
    else
        final_url="$(wget --https-only -q -O /dev/null --server-response "${latest_endpoint}" 2>&1 | sed -n 's/^ *[Ll]ocation: *//p' | tail -n 1)"
    fi
    tag="${final_url##*/}"
    if [ -z "${tag}" ] || [ "${tag}" = "latest" ]; then
        printf 'error: could not resolve the latest release tag. No release may be published yet. Pin a version with SPECHARBOR_VERSION or --version. See %s\n' "${SPECHARBOR_INSTALL_DOCS}" >&2
        return 1
    fi
    if ! is_valid_version "${tag}"; then
        printf 'error: resolved latest release tag "%s" is not a valid version.\n' "${tag}" >&2
        return 1
    fi
    printf '%s\n' "${tag}"
}

compute_sha256() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | awk '{print $1}'
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | awk '{print $1}'
    else
        printf 'error: checksum verification requires sha256sum or shasum. Install one of them and retry; verification is never skipped.\n' >&2
        return 1
    fi
}

verify_checksum() {
    # $1 = directory containing the asset and checksums.txt, $2 = asset name
    checksums_file="$1/${SPECHARBOR_CHECKSUMS_NAME}"
    [ -f "${checksums_file}" ] || sh_fail "checksum file not found: ${checksums_file}"
    expected="$(awk -v name="$2" '$2 == name || $2 == "*" name { print $1; exit }' "${checksums_file}")"
    [ -n "${expected}" ] || sh_fail "no checksum entry for $2 in ${SPECHARBOR_CHECKSUMS_NAME}; refusing to install"
    actual="$(compute_sha256 "$1/$2")"
    expected="$(printf '%s' "${expected}" | tr 'A-F' 'a-f')"
    actual="$(printf '%s' "${actual}" | tr 'A-F' 'a-f')"
    if [ "${expected}" != "${actual}" ]; then
        sh_fail "checksum verification failed for $2 (expected ${expected}, got ${actual}); aborting install"
    fi
}

path_contains() {
    case ":${PATH}:" in
        *":$1:"*)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

install_binary() {
    # $1 = verified extracted binary, $2 = install directory
    if ! mkdir -p "$2" 2>/dev/null; then
        sh_fail "could not create install directory $2. This installer never uses sudo; choose a user-writable directory with SPECHARBOR_INSTALL_DIR."
    fi
    if [ ! -w "$2" ]; then
        sh_fail "install directory $2 is not writable. This installer never uses sudo; choose a user-writable directory such as \$HOME/.local/bin with SPECHARBOR_INSTALL_DIR."
    fi
    SPECHARBOR_PARTIAL_TARGET="$2/.${SPECHARBOR_BINARY_NAME}.partial.$$"
    cp "$1" "${SPECHARBOR_PARTIAL_TARGET}"
    chmod 0755 "${SPECHARBOR_PARTIAL_TARGET}"
    mv -f "${SPECHARBOR_PARTIAL_TARGET}" "$2/${SPECHARBOR_BINARY_NAME}"
    SPECHARBOR_PARTIAL_TARGET=""
}

main() {
    version_input="${SPECHARBOR_VERSION:-}"
    install_dir="${SPECHARBOR_INSTALL_DIR:-}"
    dry_run="${SPECHARBOR_DRY_RUN:-0}"

    while [ $# -gt 0 ]; do
        case "$1" in
            --version)
                shift
                [ $# -gt 0 ] || sh_fail "--version requires a value"
                version_input="$1"
                ;;
            --install-dir)
                shift
                [ $# -gt 0 ] || sh_fail "--install-dir requires a value"
                install_dir="$1"
                ;;
            --dry-run)
                dry_run=1
                ;;
            -h | --help)
                usage
                return 0
                ;;
            *)
                sh_fail "unknown argument: $1 (use --help)"
                ;;
        esac
        shift
    done

    if [ -z "${install_dir}" ]; then
        [ -n "${HOME:-}" ] || sh_fail "HOME is not set; provide an install directory with SPECHARBOR_INSTALL_DIR or --install-dir"
        install_dir="${HOME}/.local/bin"
    fi

    os="$(detect_os "$(uname -s)")"
    arch="$(detect_arch "$(uname -m)")"

    if [ -n "${version_input}" ]; then
        if ! is_valid_version "${version_input}"; then
            sh_fail "invalid version \"${version_input}\"; expected the form X.Y.Z or vX.Y.Z"
        fi
        tag="$(normalize_tag "${version_input}")"
    else
        tag="$(resolve_latest_tag)"
    fi

    asset="$(asset_name "${os}" "${arch}")"
    archive_url="$(asset_url "${tag}" "${asset}")"
    sums_url="$(checksums_url "${tag}")"

    if [ "${dry_run}" = "1" ]; then
        sh_log "specharbor install (dry run)"
        sh_log "os: ${os}"
        sh_log "arch: ${arch}"
        sh_log "version: ${tag}"
        sh_log "asset url: ${archive_url}"
        sh_log "checksums url: ${sums_url}"
        sh_log "install dir: ${install_dir}"
        return 0
    fi

    trap cleanup EXIT INT TERM
    SPECHARBOR_TMP_DIR="$(mktemp -d)"

    sh_log "Downloading ${archive_url}"
    http_download "${archive_url}" "${SPECHARBOR_TMP_DIR}/${asset}"
    sh_log "Downloading ${sums_url}"
    http_download "${sums_url}" "${SPECHARBOR_TMP_DIR}/${SPECHARBOR_CHECKSUMS_NAME}"

    sh_log "Verifying SHA-256 checksum for ${asset}"
    verify_checksum "${SPECHARBOR_TMP_DIR}" "${asset}"

    mkdir -p "${SPECHARBOR_TMP_DIR}/extract"
    tar -xzf "${SPECHARBOR_TMP_DIR}/${asset}" -C "${SPECHARBOR_TMP_DIR}/extract" "${SPECHARBOR_BINARY_NAME}" ||
        sh_fail "could not extract ${SPECHARBOR_BINARY_NAME} from ${asset}"
    [ -f "${SPECHARBOR_TMP_DIR}/extract/${SPECHARBOR_BINARY_NAME}" ] ||
        sh_fail "archive ${asset} did not contain the ${SPECHARBOR_BINARY_NAME} binary"

    install_binary "${SPECHARBOR_TMP_DIR}/extract/${SPECHARBOR_BINARY_NAME}" "${install_dir}"

    sh_log "Installed ${SPECHARBOR_BINARY_NAME} ${tag} to ${install_dir}/${SPECHARBOR_BINARY_NAME}"
    sh_log "Run \"${SPECHARBOR_BINARY_NAME} version\" to verify the installation."

    if ! path_contains "${install_dir}"; then
        sh_warn "${install_dir} is not on your PATH."
        sh_warn "Add it to your shell profile, for example:"
        sh_warn "  export PATH=\"${install_dir}:\$PATH\""
    fi
}

# Test harnesses source this file with SPECHARBOR_INSTALL_SH_NO_MAIN=1 to call
# individual functions without running an install.
if [ "${SPECHARBOR_INSTALL_SH_NO_MAIN:-0}" != "1" ]; then
    main "$@"
fi
