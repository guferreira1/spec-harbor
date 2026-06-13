#!/bin/sh
# Tests for scripts/validate-release-version.sh.
#
# Runs the validator against accept and reject cases using the real package
# version from packages/npm/specharbor/package.json, so the suite stays correct
# across version bumps. No network access and no writes outside a temp file.
#
# Run from anywhere:
#   sh scripts/test-validate-release-version.sh

set -u

REPO_ROOT="$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)"
VALIDATOR="${REPO_ROOT}/scripts/validate-release-version.sh"
PACKAGE_JSON="${REPO_ROOT}/packages/npm/specharbor/package.json"

TESTS_RUN=0
TESTS_FAILED=0

package_version="$(
    sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([0-9][0-9.]*\)".*/\1/p' "${PACKAGE_JSON}" |
        head -n 1
)"

# Build a mismatched version by bumping the patch number, so the value is a
# valid X.Y.Z that is guaranteed not to equal the real package version.
mismatch_version="${package_version%.*}.$(( ${package_version##*.} + 1 ))"

t_pass() {
    TESTS_RUN=$((TESTS_RUN + 1))
    printf 'ok - %s\n' "$1"
}

t_fail() {
    TESTS_RUN=$((TESTS_RUN + 1))
    TESTS_FAILED=$((TESTS_FAILED + 1))
    printf 'not ok - %s\n' "$1"
}

assert_accepts() {
    # $1 = tag, $2 = description
    if sh "${VALIDATOR}" "$1" >/dev/null 2>&1; then
        t_pass "$2"
    else
        t_fail "$2 (expected exit 0 for tag \"$1\")"
    fi
}

assert_rejects() {
    # $1 = tag, $2 = description
    if sh "${VALIDATOR}" "$1" >/dev/null 2>&1; then
        t_fail "$2 (expected non-zero exit for tag \"$1\")"
    else
        t_pass "$2"
    fi
}

# Accept: exact stable SemVer tag matching the package version.
assert_accepts "v${package_version}" "accepts vX.Y.Z matching package.json"

# Reject: bad tag formats.
assert_rejects "${package_version}" "rejects X.Y.Z without leading v"
assert_rejects "v0.2" "rejects vX.Y (missing patch)"
assert_rejects "v0.2.0-beta.1" "rejects prerelease tag"
assert_rejects "vX.Y.Z" "rejects non-numeric tag"

# Reject: well-formed tag whose version does not match package.json.
assert_rejects "v${mismatch_version}" "rejects tag/package version mismatch"

# Reject: missing tag input.
if ( unset GITHUB_REF_NAME; sh "${VALIDATOR}" >/dev/null 2>&1 ); then
    t_fail "rejects missing tag input (expected non-zero exit)"
else
    t_pass "rejects missing tag input"
fi

printf '\n%d checks, %d failures\n' "${TESTS_RUN}" "${TESTS_FAILED}"
[ "${TESTS_FAILED}" -eq 0 ]
