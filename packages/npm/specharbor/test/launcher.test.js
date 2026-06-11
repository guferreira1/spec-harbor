'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

const { NATIVE_DIR } = require('../lib/install');

const PACKAGE_ROOT = path.join(__dirname, '..');
const LAUNCHER = path.join(PACKAGE_ROOT, 'bin', 'specharbor.js');
const POSTINSTALL = path.join(PACKAGE_ROOT, 'scripts', 'postinstall.js');
const FAKE_BINARY = path.join(NATIVE_DIR, 'specharbor');

const isPosix = process.platform !== 'win32';

// Installs a fake native binary so launcher tests run offline. The fake
// echoes its arguments one per line and exits with FAKE_EXIT_CODE.
function installFakeBinary() {
  fs.mkdirSync(NATIVE_DIR, { recursive: true });
  fs.writeFileSync(
    FAKE_BINARY,
    '#!/bin/sh\nfor arg in "$@"; do printf \'arg:%s\\n\' "$arg"; done\nexit "${FAKE_EXIT_CODE:-0}"\n'
  );
  fs.chmodSync(FAKE_BINARY, 0o755);
}

function removeFakeBinary() {
  fs.rmSync(NATIVE_DIR, { recursive: true, force: true });
}

test('launcher forwards arguments verbatim with no shell interpretation', { skip: !isPosix }, () => {
  installFakeBinary();
  try {
    const tricky = '; echo injected && rm -rf $(HOME) `whoami` | cat';
    const result = spawnSync(process.execPath, [LAUNCHER, 'validate', 'my change', tricky], {
      encoding: 'utf8',
    });
    assert.equal(result.status, 0);
    assert.ok(result.stdout.includes('arg:validate'));
    assert.ok(result.stdout.includes('arg:my change'));
    assert.ok(result.stdout.includes(`arg:${tricky}`), 'metacharacters must arrive literally');
    assert.ok(!result.stdout.includes('injected\n'), 'shell metacharacters must not execute');
  } finally {
    removeFakeBinary();
  }
});

test('launcher preserves the native binary exit code', { skip: !isPosix }, () => {
  installFakeBinary();
  try {
    const result = spawnSync(process.execPath, [LAUNCHER, 'version'], {
      encoding: 'utf8',
      env: { ...process.env, FAKE_EXIT_CODE: '7' },
    });
    assert.equal(result.status, 7);
  } finally {
    removeFakeBinary();
  }
});

test('postinstall succeeds without network when the binary is present', { skip: !isPosix }, () => {
  installFakeBinary();
  try {
    const result = spawnSync(process.execPath, [POSTINSTALL], { encoding: 'utf8' });
    assert.equal(result.status, 0);
    assert.ok(result.stdout.includes('native binary ready'));
  } finally {
    removeFakeBinary();
  }
});
