'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('fs');
const os = require('os');
const path = require('path');

const { ensureBinary, pinnedVersion } = require('../lib/install');
const { sha256Hex } = require('../lib/download');
const { assetUrl, checksumsUrl, releaseTag, UnsupportedPlatformError } = require('../lib/platform');
const { makeTarGz, fakeGet } = require('./helpers');
const { fetchBuffer } = require('../lib/download');

const quiet = () => {};

function makeTempDir() {
  return fs.mkdtempSync(path.join(os.tmpdir(), 'specharbor-npm-test-'));
}

function fixtureRelease() {
  const tag = releaseTag(pinnedVersion());
  const name = 'specharbor_Linux_x86_64.tar.gz';
  const archive = makeTarGz([{ name: 'specharbor', data: 'native-binary-bytes' }]);
  const checksums = `${sha256Hex(archive)}  ${name}\n`;
  return { tag, name, archive, checksums };
}

test('existing binary is reused without any download', async () => {
  const nativeDir = makeTempDir();
  try {
    fs.writeFileSync(path.join(nativeDir, 'specharbor'), 'already-installed');
    const neverFetch = () => {
      throw new Error('fetch must not be called when the binary already exists');
    };
    const binary = await ensureBinary({
      platform: 'linux',
      arch: 'x64',
      nativeDir,
      fetchBuffer: neverFetch,
      log: quiet,
    });
    assert.equal(binary, path.join(nativeDir, 'specharbor'));
  } finally {
    fs.rmSync(nativeDir, { recursive: true, force: true });
  }
});

test('first-run fallback downloads, verifies, and installs the binary', async () => {
  const nativeDir = path.join(makeTempDir(), 'native');
  try {
    const { tag, name, archive, checksums } = fixtureRelease();
    const get = fakeGet({
      [assetUrl(tag, name)]: { body: archive },
      [checksumsUrl(tag)]: { body: checksums },
    });
    const binary = await ensureBinary({
      platform: 'linux',
      arch: 'x64',
      nativeDir,
      fetchBuffer: (url) => fetchBuffer(url, { get }),
      log: quiet,
    });
    assert.equal(fs.readFileSync(binary, 'utf8'), 'native-binary-bytes');
    assert.ok(fs.statSync(binary).mode & 0o100, 'binary must be executable');
  } finally {
    fs.rmSync(path.dirname(nativeDir), { recursive: true, force: true });
  }
});

test('checksum mismatch aborts the install and leaves no binary behind', async () => {
  const nativeDir = path.join(makeTempDir(), 'native');
  try {
    const { tag, name, archive } = fixtureRelease();
    const get = fakeGet({
      [assetUrl(tag, name)]: { body: archive },
      [checksumsUrl(tag)]: { body: `${'0'.repeat(64)}  ${name}\n` },
    });
    await assert.rejects(
      ensureBinary({
        platform: 'linux',
        arch: 'x64',
        nativeDir,
        fetchBuffer: (url) => fetchBuffer(url, { get }),
        log: quiet,
      }),
      /checksum verification failed/
    );
    assert.ok(
      !fs.existsSync(path.join(nativeDir, 'specharbor')),
      'no binary may be installed after a checksum failure'
    );
  } finally {
    fs.rmSync(path.dirname(nativeDir), { recursive: true, force: true });
  }
});

test('missing checksum entry aborts the install', async () => {
  const nativeDir = path.join(makeTempDir(), 'native');
  try {
    const { tag, name, archive } = fixtureRelease();
    const get = fakeGet({
      [assetUrl(tag, name)]: { body: archive },
      [checksumsUrl(tag)]: { body: `${'0'.repeat(64)}  unrelated.tar.gz\n` },
    });
    await assert.rejects(
      ensureBinary({
        platform: 'linux',
        arch: 'x64',
        nativeDir,
        fetchBuffer: (url) => fetchBuffer(url, { get }),
        log: quiet,
      }),
      /no checksum entry/
    );
  } finally {
    fs.rmSync(path.dirname(nativeDir), { recursive: true, force: true });
  }
});

test('unsupported platforms fail before any download', async () => {
  const neverFetch = () => {
    throw new Error('fetch must not be called for unsupported platforms');
  };
  await assert.rejects(
    ensureBinary({ platform: 'freebsd', arch: 'x64', fetchBuffer: neverFetch, log: quiet }),
    (error) => error instanceof UnsupportedPlatformError
  );
});

test('pinned version maps to exactly one release tag', () => {
  assert.match(pinnedVersion(), /^\d+\.\d+\.\d+$/);
  assert.equal(releaseTag(pinnedVersion()), `v${pinnedVersion()}`);
});
