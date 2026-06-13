#!/bin/sh
# Release version consistency gate.
#
# Validates that a release tag is an exact stable SemVer tag (vX.Y.Z) and that
# its X.Y.Z matches the npm wrapper package version in
# packages/npm/specharbor/package.json. The npm wrapper maps package version
# X.Y.Z to release tag vX.Y.Z, so this check guarantees the published npm
# package downloads assets from the matching GitHub Release.
#
# Usage:
#   scripts/validate-release-version.sh [tag]
#
# The tag is taken from the first argument, or from GITHUB_REF_NAME when no
# argument is given (the GitHub Actions tag ref name for tag pushes).
#
# Exit status:
#   0  tag is a stable vX.Y.Z and equals the package version
#   1  invalid tag format or version mismatch
#
# This script never prints secrets, never writes files, and never accesses the
# network.

set -eu

REPO_ROOT="$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)"
PACKAGE_JSON="${REPO_ROOT}/packages/npm/specharbor/package.json"

fail() {
    printf 'validate-release-version: %s\n' "$1" >&2
    exit 1
}

tag="${1:-${GITHUB_REF_NAME:-}}"
[ -n "${tag}" ] || fail 'no tag provided; pass a tag argument or set GITHUB_REF_NAME'

# Stable SemVer release tags only: vMAJOR.MINOR.PATCH with no prerelease or
# build metadata. Examples: v0.2.0 (valid); 0.2.0, v0.2, v0.2.0-beta.1 (invalid).
if ! printf '%s' "${tag}" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
    fail "tag \"${tag}\" is not an exact stable SemVer release tag (expected vX.Y.Z)"
fi

tag_version="${tag#v}"

[ -f "${PACKAGE_JSON}" ] || fail "package manifest not found: ${PACKAGE_JSON}"

# Extract the top-level "version" field without depending on Node or jq.
package_version="$(
    sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([0-9][0-9.]*\)".*/\1/p' "${PACKAGE_JSON}" |
        head -n 1
)"
[ -n "${package_version}" ] || fail "could not read version from ${PACKAGE_JSON}"

if ! printf '%s' "${package_version}" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    fail "package version \"${package_version}\" is not a plain X.Y.Z version"
fi

if [ "${tag_version}" != "${package_version}" ]; then
    fail "version mismatch: tag ${tag} maps to ${tag_version} but packages/npm/specharbor/package.json is ${package_version}"
fi

printf 'release version validated: tag %s matches package version %s\n' "${tag}" "${package_version}"
