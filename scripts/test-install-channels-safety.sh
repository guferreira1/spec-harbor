#!/bin/sh
# Static safety and regression checks for the install channels change.
#
# Confirms repository-level boundaries: no Homebrew formula files in this
# repository, no Linux/Windows package-manager manifests, no generated release
# artifacts, and documentation covering the required install topics.
#
# Run from the repository root:
#   sh scripts/test-install-channels-safety.sh

set -u

REPO_ROOT="$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)"
cd "${REPO_ROOT}" || exit 1

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

assert_absent() {
    # $1 = path that must not exist
    if [ -e "$1" ]; then
        t_fail "forbidden path must not exist: $1"
    else
        t_pass "forbidden path absent: $1"
    fi
}

assert_present() {
    # $1 = path that must exist
    if [ -e "$1" ]; then
        t_pass "required path present: $1"
    else
        t_fail "required path missing: $1"
    fi
}

assert_doc_contains() {
    # $1 = file, $2 = required content
    if grep -qF "$2" "$1" 2>/dev/null; then
        t_pass "$1 documents: $2"
    else
        t_fail "$1 must document: $2"
    fi
}

# --- No Linux/Windows package-manager or Homebrew files in this repository ------

for forbidden in nfpm.yaml .nfpm.yaml packaging debian rpm scoop winget Formula HomebrewFormula homebrew dist; do
    assert_absent "${forbidden}"
done

if find . -path ./node_modules -prune -o -name '*.rb' -print | grep -q .; then
    t_fail "no Homebrew formula (.rb) files in this repository"
else
    t_pass "no Homebrew formula (.rb) files in this repository"
fi

if find . -name '*.deb' -o -name '*.rpm' | grep -q .; then
    t_fail "no .deb or .rpm artifacts in this repository"
else
    t_pass "no .deb or .rpm artifacts in this repository"
fi

# --- No generated release artifacts ----------------------------------------------

if find . -path ./.git -prune -o \( -name 'checksums.txt' -o -name 'specharbor_*.tar.gz' -o -name 'specharbor_*.zip' \) -print | grep -q .; then
    t_fail "no generated release archives or checksum files committed"
else
    t_pass "no generated release archives or checksum files committed"
fi

# --- Install channel files exist --------------------------------------------------

assert_present install.sh
assert_present docs/install.md
assert_present packages/npm/specharbor/package.json
assert_present packages/npm/specharbor/bin/specharbor.js
assert_present packages/npm/specharbor/scripts/postinstall.js

if head -n 1 install.sh | grep -q '^#!/bin/sh$'; then
    t_pass "install.sh declares a POSIX sh shebang"
else
    t_fail "install.sh declares a POSIX sh shebang"
fi

# --- No lockfiles for the npm wrapper ---------------------------------------------

for lockfile in packages/npm/specharbor/package-lock.json packages/npm/specharbor/npm-shrinkwrap.json packages/npm/specharbor/yarn.lock; do
    assert_absent "${lockfile}"
done

# --- No publishing or source-control automation -----------------------------------

if grep -qE 'npm publish|gh release|git tag|git push' install.sh packages/npm/specharbor/package.json packages/npm/specharbor/scripts/postinstall.js; then
    t_fail "install channels contain no publishing or source-control automation"
else
    t_pass "install channels contain no publishing or source-control automation"
fi

# --- Homebrew tap expectations are documented (formula lives in the external tap) -

assert_doc_contains docs/install.md "guferreira1/homebrew-tap"
assert_doc_contains docs/install.md "brew install guferreira1/tap/specharbor"
assert_doc_contains docs/install.md "sha256"
assert_doc_contains docs/install.md "specharbor version"
assert_doc_contains docs/install.md "test do"

# --- Documentation covers the required install topics -----------------------------

assert_doc_contains docs/install.md "checksums.txt"
assert_doc_contains docs/install.md "sha256sum"
assert_doc_contains docs/install.md "shasum -a 256"
assert_doc_contains docs/install.md "SPECHARBOR_INSTALL_DIR"
assert_doc_contains docs/install.md "SPECHARBOR_VERSION"
assert_doc_contains docs/install.md "PATH"
assert_doc_contains docs/install.md "go install"
assert_doc_contains docs/install.md "Scoop"
assert_doc_contains docs/install.md "Winget"
assert_doc_contains docs/install.md ".deb"
assert_doc_contains docs/install.md ".rpm"
assert_doc_contains README.md "docs/install.md"
assert_doc_contains docs/release.md "install.md"

# --- Summary -----------------------------------------------------------------------

printf '\n%d checks, %d failures\n' "${TESTS_RUN}" "${TESTS_FAILED}"
[ "${TESTS_FAILED}" -eq 0 ]
