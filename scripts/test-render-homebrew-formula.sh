#!/bin/sh
# Tests for scripts/render-homebrew-formula.sh using a fixture checksums file.
# Runs fully offline and writes only to a temporary directory.
#
# Run from anywhere:
#   sh scripts/test-render-homebrew-formula.sh

set -u

REPO_ROOT="$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)"
RENDERER="${REPO_ROOT}/scripts/render-homebrew-formula.sh"

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

assert_contains() {
    # $1 = haystack, $2 = needle, $3 = description
    case "$1" in
        *"$2"*) t_pass "$3" ;;
        *) t_fail "$3 (missing: $2)" ;;
    esac
}

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

CHECKSUMS="${TMP_DIR}/checksums.txt"
ARM_SHA="1111111111111111111111111111111111111111111111111111111111111111"
INTEL_SHA="2222222222222222222222222222222222222222222222222222222222222222"
cat >"${CHECKSUMS}" <<EOF
${ARM_SHA}  specharbor_Darwin_arm64.tar.gz
${INTEL_SHA}  specharbor_Darwin_x86_64.tar.gz
3333333333333333333333333333333333333333333333333333333333333333  specharbor_Linux_x86_64.tar.gz
EOF

formula="$(sh "${RENDERER}" --version 1.2.3 --checksums "${CHECKSUMS}")"
status=$?

if [ "${status}" -ne 0 ]; then
    t_fail "renderer exits zero for valid input"
else
    t_pass "renderer exits zero for valid input"
    assert_contains "${formula}" 'class Specharbor < Formula' "renders the formula class"
    assert_contains "${formula}" 'version "1.2.3"' "renders the version"
    assert_contains "${formula}" 'releases/download/v1.2.3/specharbor_Darwin_arm64.tar.gz' "uses the arm64 release URL"
    assert_contains "${formula}" "sha256 \"${ARM_SHA}\"" "pins the arm64 sha256"
    assert_contains "${formula}" 'releases/download/v1.2.3/specharbor_Darwin_x86_64.tar.gz' "uses the x86_64 release URL"
    assert_contains "${formula}" "sha256 \"${INTEL_SHA}\"" "pins the x86_64 sha256"
    assert_contains "${formula}" 'bin.install "specharbor"' "installs the binary"
    assert_contains "${formula}" 'shell_output("#{bin}/specharbor version")' "tests specharbor version"
fi

# Reject: missing checksum entry for a required asset.
if printf 'abc  unrelated.tar.gz\n' >"${TMP_DIR}/bad.txt" &&
    sh "${RENDERER}" --version 1.2.3 --checksums "${TMP_DIR}/bad.txt" >/dev/null 2>&1; then
    t_fail "rejects checksums without the macOS assets"
else
    t_pass "rejects checksums without the macOS assets"
fi

# Reject: non-X.Y.Z version.
if sh "${RENDERER}" --version v1.2.3 --checksums "${CHECKSUMS}" >/dev/null 2>&1; then
    t_fail "rejects a non-plain version"
else
    t_pass "rejects a non-plain version"
fi

printf '\n%d checks, %d failures\n' "${TESTS_RUN}" "${TESTS_FAILED}"
[ "${TESTS_FAILED}" -eq 0 ]
