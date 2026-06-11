#!/bin/sh
# Tests for install.sh.
#
# All downloads are served by a stub curl placed first on PATH; no test
# touches the network, real GitHub Releases, Git state, or package registries.
#
# Run from the repository root:
#   sh scripts/test-install-sh.sh

set -u

REPO_ROOT="$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)"
INSTALL_SH="${REPO_ROOT}/install.sh"

TESTS_RUN=0
TESTS_FAILED=0

t_pass() {
    TESTS_RUN=$((TESTS_RUN + 1))
    printf 'ok - %s\n' "$1"
}

t_fail() {
    TESTS_RUN=$((TESTS_RUN + 1))
    TESTS_FAILED=$((TESTS_FAILED + 1))
    printf 'not ok - %s\n' "$1"
}

assert_eq() {
    # $1 = test name, $2 = expected, $3 = actual
    if [ "$2" = "$3" ]; then
        t_pass "$1"
    else
        t_fail "$1 (expected \"$2\", got \"$3\")"
    fi
}

assert_contains() {
    # $1 = test name, $2 = haystack, $3 = needle
    case "$2" in
        *"$3"*)
            t_pass "$1"
            ;;
        *)
            t_fail "$1 (output did not contain \"$3\")"
            ;;
    esac
}

assert_success() {
    # $1 = test name, $2 = exit code
    if [ "$2" -eq 0 ]; then
        t_pass "$1"
    else
        t_fail "$1 (expected exit 0, got $2)"
    fi
}

assert_failure() {
    # $1 = test name, $2 = exit code
    if [ "$2" -ne 0 ]; then
        t_pass "$1"
    else
        t_fail "$1 (expected non-zero exit, got 0)"
    fi
}

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT INT TERM

# --- Stub curl ----------------------------------------------------------------

STUB_BIN="${WORK_DIR}/stub-bin"
FAKE_ASSET_DIR="${WORK_DIR}/assets"
mkdir -p "${STUB_BIN}" "${FAKE_ASSET_DIR}"

cat >"${STUB_BIN}/curl" <<'EOF'
#!/bin/sh
# Stub curl for install.sh tests. Serves files from SPECHARBOR_FAKE_ASSET_DIR
# and answers the releases/latest redirect with SPECHARBOR_FAKE_LATEST_URL.
out=""
writeout=""
url=""
while [ $# -gt 0 ]; do
    case "$1" in
        -o)
            shift
            out="$1"
            ;;
        -w)
            shift
            writeout="$1"
            ;;
        --proto | --proto-redir)
            shift
            ;;
        -*) ;;
        *)
            url="$1"
            ;;
    esac
    shift
done
case "${url}" in
    */releases/latest)
        if [ -n "${writeout}" ]; then
            printf '%s' "${SPECHARBOR_FAKE_LATEST_URL:-${url}}"
        fi
        exit 0
        ;;
    *)
        name="${url##*/}"
        if [ -f "${SPECHARBOR_FAKE_ASSET_DIR:-/nonexistent}/${name}" ]; then
            cp "${SPECHARBOR_FAKE_ASSET_DIR}/${name}" "${out}"
            exit 0
        fi
        exit 22
        ;;
esac
EOF
chmod 0755 "${STUB_BIN}/curl"

# --- Fixture release ----------------------------------------------------------

FIXTURE_SRC="${WORK_DIR}/fixture-src"
mkdir -p "${FIXTURE_SRC}"
printf '#!/bin/sh\necho "SpecHarbor 0.1.0"\n' >"${FIXTURE_SRC}/specharbor"
chmod 0755 "${FIXTURE_SRC}/specharbor"

host_arch_name() {
    case "$(uname -m)" in
        x86_64 | amd64) printf 'x86_64\n' ;;
        aarch64 | arm64) printf 'arm64\n' ;;
        *) printf 'x86_64\n' ;;
    esac
}

HOST_OS="$(uname -s)"
HOST_ASSET="specharbor_${HOST_OS}_$(host_arch_name).tar.gz"

(cd "${FIXTURE_SRC}" && tar -czf "${FAKE_ASSET_DIR}/${HOST_ASSET}" specharbor)
if command -v sha256sum >/dev/null 2>&1; then
    FIXTURE_SHA="$(sha256sum "${FAKE_ASSET_DIR}/${HOST_ASSET}" | awk '{print $1}')"
else
    FIXTURE_SHA="$(shasum -a 256 "${FAKE_ASSET_DIR}/${HOST_ASSET}" | awk '{print $1}')"
fi
printf '%s  %s\n' "${FIXTURE_SHA}" "${HOST_ASSET}" >"${FAKE_ASSET_DIR}/checksums.txt"

# --- Function-level tests (sourced, no main) -----------------------------------

SPECHARBOR_INSTALL_SH_NO_MAIN=1
. "${INSTALL_SH}"
# install.sh enables `set -e`; the harness must keep running after expected failures.
set +e

out="$(detect_os Linux)" && assert_eq "detect_os maps Linux" "Linux" "${out}"
out="$(detect_os Darwin)" && assert_eq "detect_os maps Darwin" "Darwin" "${out}"
if out="$(detect_os FreeBSD 2>/dev/null)"; then
    t_fail "detect_os rejects FreeBSD"
else
    t_pass "detect_os rejects FreeBSD"
fi
if out="$(detect_os MINGW64_NT 2>/dev/null)"; then
    t_fail "detect_os rejects Windows uname values"
else
    t_pass "detect_os rejects Windows uname values"
fi

out="$(detect_arch x86_64)" && assert_eq "detect_arch maps x86_64 to x86_64" "x86_64" "${out}"
out="$(detect_arch amd64)" && assert_eq "detect_arch maps amd64 to x86_64" "x86_64" "${out}"
out="$(detect_arch aarch64)" && assert_eq "detect_arch maps aarch64 to arm64" "arm64" "${out}"
out="$(detect_arch arm64)" && assert_eq "detect_arch maps arm64 to arm64" "arm64" "${out}"
if out="$(detect_arch i686 2>/dev/null)"; then
    t_fail "detect_arch rejects i686"
else
    t_pass "detect_arch rejects i686"
fi

for good in 0.1.0 v0.1.0 v12.34.56; do
    if is_valid_version "${good}"; then
        t_pass "is_valid_version accepts ${good}"
    else
        t_fail "is_valid_version accepts ${good}"
    fi
done
for bad in "v0.1" "0.1.0.0" "latest" "v0.1.0;rm -rf /" "../../etc" "v0.1.0 " ""; do
    if is_valid_version "${bad}"; then
        t_fail "is_valid_version rejects \"${bad}\""
    else
        t_pass "is_valid_version rejects \"${bad}\""
    fi
done

assert_eq "normalize_tag keeps v prefix" "v0.1.0" "$(normalize_tag v0.1.0)"
assert_eq "normalize_tag adds v prefix" "v0.1.0" "$(normalize_tag 0.1.0)"

assert_eq "asset_name Linux x86_64" "specharbor_Linux_x86_64.tar.gz" "$(asset_name Linux x86_64)"
assert_eq "asset_name Darwin arm64" "specharbor_Darwin_arm64.tar.gz" "$(asset_name Darwin arm64)"
assert_eq "asset_url is the allowlisted release download URL" \
    "https://github.com/guferreira1/spec-harbor/releases/download/v0.1.0/specharbor_Linux_x86_64.tar.gz" \
    "$(asset_url v0.1.0 specharbor_Linux_x86_64.tar.gz)"
assert_eq "checksums_url is the allowlisted release download URL" \
    "https://github.com/guferreira1/spec-harbor/releases/download/v0.1.0/checksums.txt" \
    "$(checksums_url v0.1.0)"

out="$( (http_download "https://evil.example.com/specharbor.tar.gz" "${WORK_DIR}/never") 2>&1)"
assert_failure "http_download rejects non-allowlisted hosts" "$?"
assert_contains "http_download names the rejected URL" "${out}" "non-allowlisted"
out="$( (http_download "http://github.com/guferreira1/spec-harbor/releases/download/v0.1.0/x" "${WORK_DIR}/never") 2>&1)"
assert_failure "http_download rejects plain http" "$?"
[ ! -f "${WORK_DIR}/never" ] && t_pass "rejected downloads write nothing" || t_fail "rejected downloads write nothing"

# --- Static safety checks ------------------------------------------------------

# "sudo" may appear inside comments and sh_fail guidance messages, but must
# never appear on an executable line as a command.
if grep -vE '^[[:space:]]*#' "${INSTALL_SH}" | grep -v 'sh_fail' |
    grep -E '(^|[^[:alnum:]_])sudo($|[^[:alnum:]_])' >/dev/null; then
    t_fail "install.sh never invokes sudo"
else
    t_pass "install.sh never invokes sudo"
fi
if grep -E '(^|[^[:alnum:]_])eval($|[^[:alnum:]_])' "${INSTALL_SH}" >/dev/null; then
    t_fail "install.sh never uses eval"
else
    t_pass "install.sh never uses eval"
fi
if grep -F 'http://' "${INSTALL_SH}" >/dev/null; then
    t_fail "install.sh contains no plain-http URLs"
else
    t_pass "install.sh contains no plain-http URLs"
fi

# --- Dry-run tests --------------------------------------------------------------

out="$(env PATH="${STUB_BIN}:${PATH}" SPECHARBOR_FAKE_ASSET_DIR="${FAKE_ASSET_DIR}" \
    SPECHARBOR_DRY_RUN=1 SPECHARBOR_VERSION=0.1.0 sh "${INSTALL_SH}" 2>&1)"
code=$?
assert_success "dry run with pinned version succeeds" "${code}"
assert_contains "dry run prints os" "${out}" "os: ${HOST_OS}"
assert_contains "dry run prints arch" "${out}" "arch: $(host_arch_name)"
assert_contains "dry run prints version" "${out}" "version: v0.1.0"
assert_contains "dry run prints asset url" "${out}" \
    "asset url: https://github.com/guferreira1/spec-harbor/releases/download/v0.1.0/${HOST_ASSET}"
assert_contains "dry run prints checksums url" "${out}" \
    "checksums url: https://github.com/guferreira1/spec-harbor/releases/download/v0.1.0/checksums.txt"

out="$(env PATH="${STUB_BIN}:${PATH}" SPECHARBOR_FAKE_ASSET_DIR="${FAKE_ASSET_DIR}" \
    SPECHARBOR_FAKE_LATEST_URL="https://github.com/guferreira1/spec-harbor/releases/tag/v0.2.3" \
    sh "${INSTALL_SH}" --dry-run 2>&1)"
code=$?
assert_success "dry run resolves the latest release tag" "${code}"
assert_contains "latest resolution uses the redirect tag" "${out}" "version: v0.2.3"

out="$(env PATH="${STUB_BIN}:${PATH}" SPECHARBOR_FAKE_ASSET_DIR="${FAKE_ASSET_DIR}" \
    SPECHARBOR_FAKE_LATEST_URL="https://github.com/guferreira1/spec-harbor/releases/latest" \
    sh "${INSTALL_SH}" --dry-run 2>&1)"
assert_failure "unresolvable latest release fails clearly" "$?"
assert_contains "unresolvable latest suggests pinning" "${out}" "Pin a version"

out="$(env SPECHARBOR_DRY_RUN=1 SPECHARBOR_VERSION="0.1.0;rm -rf /" sh "${INSTALL_SH}" 2>&1)"
assert_failure "invalid version strings are rejected before URL use" "$?"
assert_contains "invalid version error names the expected form" "${out}" "X.Y.Z"

out="$(env SPECHARBOR_DRY_RUN=1 sh "${INSTALL_SH}" --version 2>&1)"
assert_failure "--version without a value fails" "$?"

out="$(env SPECHARBOR_DRY_RUN=1 SPECHARBOR_VERSION=0.1.0 sh "${INSTALL_SH}" --bogus-flag 2>&1)"
assert_failure "unknown arguments fail" "$?"

out="$(env SPECHARBOR_DRY_RUN=1 SPECHARBOR_VERSION=0.1.0 SPECHARBOR_INSTALL_DIR=/custom/dir sh "${INSTALL_SH}" 2>&1)"
assert_contains "SPECHARBOR_INSTALL_DIR overrides the install dir" "${out}" "install dir: /custom/dir"

out="$(env SPECHARBOR_DRY_RUN=1 SPECHARBOR_VERSION=0.1.0 HOME="${WORK_DIR}/home" sh "${INSTALL_SH}" 2>&1)"
assert_contains "default install dir is HOME/.local/bin" "${out}" "install dir: ${WORK_DIR}/home/.local/bin"

out="$(env SPECHARBOR_DRY_RUN=1 SPECHARBOR_VERSION=0.1.0 sh "${INSTALL_SH}" --install-dir /flag/dir 2>&1)"
assert_contains "--install-dir overrides the install dir" "${out}" "install dir: /flag/dir"

# --- End-to-end install with stubbed downloads ----------------------------------

INSTALL_HOME="${WORK_DIR}/install-home"
INSTALL_TMP="${WORK_DIR}/install-tmp"
mkdir -p "${INSTALL_HOME}" "${INSTALL_TMP}"
out="$(env PATH="${STUB_BIN}:${PATH}" SPECHARBOR_FAKE_ASSET_DIR="${FAKE_ASSET_DIR}" \
    HOME="${INSTALL_HOME}" TMPDIR="${INSTALL_TMP}" SPECHARBOR_VERSION=0.1.0 \
    sh "${INSTALL_SH}" 2>&1)"
code=$?
assert_success "stubbed install succeeds" "${code}"
[ -x "${INSTALL_HOME}/.local/bin/specharbor" ] &&
    t_pass "binary installed to default HOME/.local/bin with exec permission" ||
    t_fail "binary installed to default HOME/.local/bin with exec permission"
assert_contains "success output names the installed path" "${out}" "${INSTALL_HOME}/.local/bin/specharbor"
assert_contains "success output suggests specharbor version" "${out}" "specharbor version"
assert_contains "install dir not on PATH produces guidance" "${out}" "is not on your PATH"
[ -z "$(ls -A "${INSTALL_TMP}")" ] &&
    t_pass "temporary directory is cleaned up after success" ||
    t_fail "temporary directory is cleaned up after success"

INSTALL_DIR_OVERRIDE="${WORK_DIR}/override-bin"
out="$(env PATH="${STUB_BIN}:${PATH}" SPECHARBOR_FAKE_ASSET_DIR="${FAKE_ASSET_DIR}" \
    HOME="${INSTALL_HOME}" TMPDIR="${INSTALL_TMP}" SPECHARBOR_VERSION=0.1.0 \
    SPECHARBOR_INSTALL_DIR="${INSTALL_DIR_OVERRIDE}" sh "${INSTALL_SH}" 2>&1)"
assert_success "stubbed install honors SPECHARBOR_INSTALL_DIR" "$?"
[ -x "${INSTALL_DIR_OVERRIDE}/specharbor" ] &&
    t_pass "binary installed into SPECHARBOR_INSTALL_DIR" ||
    t_fail "binary installed into SPECHARBOR_INSTALL_DIR"

# --- Checksum failure aborts and cleans up ---------------------------------------

BAD_ASSET_DIR="${WORK_DIR}/bad-assets"
mkdir -p "${BAD_ASSET_DIR}"
cp "${FAKE_ASSET_DIR}/${HOST_ASSET}" "${BAD_ASSET_DIR}/${HOST_ASSET}"
printf '%s  %s\n' \
    "0000000000000000000000000000000000000000000000000000000000000000" \
    "${HOST_ASSET}" >"${BAD_ASSET_DIR}/checksums.txt"

FAIL_HOME="${WORK_DIR}/fail-home"
FAIL_TMP="${WORK_DIR}/fail-tmp"
mkdir -p "${FAIL_HOME}" "${FAIL_TMP}"
out="$(env PATH="${STUB_BIN}:${PATH}" SPECHARBOR_FAKE_ASSET_DIR="${BAD_ASSET_DIR}" \
    HOME="${FAIL_HOME}" TMPDIR="${FAIL_TMP}" SPECHARBOR_VERSION=0.1.0 \
    sh "${INSTALL_SH}" 2>&1)"
assert_failure "checksum mismatch aborts the install" "$?"
assert_contains "checksum mismatch error mentions verification" "${out}" "checksum verification failed"
[ ! -e "${FAIL_HOME}/.local/bin/specharbor" ] &&
    t_pass "checksum failure leaves no binary in the install target" ||
    t_fail "checksum failure leaves no binary in the install target"
[ -z "$(ls -A "${FAIL_TMP}")" ] &&
    t_pass "checksum failure removes the temporary directory" ||
    t_fail "checksum failure removes the temporary directory"

# --- Missing checksum entry aborts ----------------------------------------------

NOENTRY_ASSET_DIR="${WORK_DIR}/noentry-assets"
mkdir -p "${NOENTRY_ASSET_DIR}"
cp "${FAKE_ASSET_DIR}/${HOST_ASSET}" "${NOENTRY_ASSET_DIR}/${HOST_ASSET}"
printf '%s  %s\n' "${FIXTURE_SHA}" "some_other_asset.tar.gz" >"${NOENTRY_ASSET_DIR}/checksums.txt"
out="$(env PATH="${STUB_BIN}:${PATH}" SPECHARBOR_FAKE_ASSET_DIR="${NOENTRY_ASSET_DIR}" \
    HOME="${FAIL_HOME}" TMPDIR="${FAIL_TMP}" SPECHARBOR_VERSION=0.1.0 \
    sh "${INSTALL_SH}" 2>&1)"
assert_failure "missing checksum entry aborts the install" "$?"
assert_contains "missing checksum entry is reported" "${out}" "no checksum entry"

# --- Unwritable install target fails without sudo --------------------------------

READONLY_DIR="${WORK_DIR}/readonly-bin"
mkdir -p "${READONLY_DIR}"
chmod 0555 "${READONLY_DIR}"
out="$(env PATH="${STUB_BIN}:${PATH}" SPECHARBOR_FAKE_ASSET_DIR="${FAKE_ASSET_DIR}" \
    HOME="${FAIL_HOME}" TMPDIR="${FAIL_TMP}" SPECHARBOR_VERSION=0.1.0 \
    SPECHARBOR_INSTALL_DIR="${READONLY_DIR}" sh "${INSTALL_SH}" 2>&1)"
assert_failure "unwritable install dir fails" "$?"
assert_contains "unwritable install dir suggests user-local guidance" "${out}" "never uses sudo"
chmod 0755 "${READONLY_DIR}"

# --- Summary ---------------------------------------------------------------------

printf '\n%d tests, %d failures\n' "${TESTS_RUN}" "${TESTS_FAILED}"
[ "${TESTS_FAILED}" -eq 0 ]
