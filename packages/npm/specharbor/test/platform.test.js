'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');

const {
  DOWNLOAD_BASE_URL,
  INSTALL_DOCS_URL,
  UnsupportedPlatformError,
  resolvePlatform,
  releaseTag,
  assetName,
  assetUrl,
  checksumsUrl,
} = require('../lib/platform');

test('supported platforms map to finalized GoReleaser asset names', () => {
  const expectations = [
    ['linux', 'x64', 'specharbor_Linux_x86_64.tar.gz', 'specharbor'],
    ['linux', 'arm64', 'specharbor_Linux_arm64.tar.gz', 'specharbor'],
    ['darwin', 'x64', 'specharbor_Darwin_x86_64.tar.gz', 'specharbor'],
    ['darwin', 'arm64', 'specharbor_Darwin_arm64.tar.gz', 'specharbor'],
    ['win32', 'x64', 'specharbor_Windows_x86_64.zip', 'specharbor.exe'],
    ['win32', 'arm64', 'specharbor_Windows_arm64.zip', 'specharbor.exe'],
  ];
  for (const [platform, arch, expectedAsset, expectedBinary] of expectations) {
    const descriptor = resolvePlatform(platform, arch);
    assert.equal(assetName(descriptor), expectedAsset);
    assert.equal(descriptor.binaryName, expectedBinary);
  }
});

test('unsupported platforms produce a deterministic error naming the platform', () => {
  const unsupported = [
    ['freebsd', 'x64'],
    ['linux', 'ia32'],
    ['darwin', 'ppc64'],
    ['sunos', 'arm64'],
  ];
  for (const [platform, arch] of unsupported) {
    assert.throws(
      () => resolvePlatform(platform, arch),
      (error) => {
        assert.ok(error instanceof UnsupportedPlatformError);
        assert.match(error.message, new RegExp(`${platform} ${arch}`));
        assert.ok(error.message.includes(INSTALL_DOCS_URL));
        return true;
      }
    );
  }
});

test('release tag maps one package version to one release tag', () => {
  assert.equal(releaseTag('0.1.0'), 'v0.1.0');
  assert.equal(releaseTag('12.34.56'), 'v12.34.56');
});

test('release tag rejects versions that do not map to exactly one tag', () => {
  for (const bad of ['0.1', 'v0.1.0', '0.1.0-beta.1', '0.1.0.0', 'latest', '']) {
    assert.throws(() => releaseTag(bad), /invalid package version/);
  }
});

test('asset URLs always use the allowlisted HTTPS release download prefix', () => {
  assert.equal(
    DOWNLOAD_BASE_URL,
    'https://github.com/guferreira1/spec-harbor/releases/download/'
  );
  assert.equal(
    assetUrl('v0.1.0', 'specharbor_Linux_x86_64.tar.gz'),
    'https://github.com/guferreira1/spec-harbor/releases/download/v0.1.0/specharbor_Linux_x86_64.tar.gz'
  );
  assert.equal(
    checksumsUrl('v0.1.0'),
    'https://github.com/guferreira1/spec-harbor/releases/download/v0.1.0/checksums.txt'
  );
});
