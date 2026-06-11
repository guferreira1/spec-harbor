'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const crypto = require('crypto');

const {
  assertAllowedInitialUrl,
  assertAllowedRedirectUrl,
  fetchBuffer,
  sha256Hex,
  parseChecksums,
  verifyChecksum,
} = require('../lib/download');
const { fakeGet } = require('./helpers');

const ASSET_URL =
  'https://github.com/guferreira1/spec-harbor/releases/download/v0.1.0/specharbor_Linux_x86_64.tar.gz';

test('initial URLs outside the release download allowlist are rejected', () => {
  assertAllowedInitialUrl(ASSET_URL);
  const rejected = [
    'http://github.com/guferreira1/spec-harbor/releases/download/v0.1.0/x.tar.gz',
    'https://evil.example.com/guferreira1/spec-harbor/releases/download/v0.1.0/x.tar.gz',
    'https://github.com.evil.example.com/guferreira1/spec-harbor/releases/download/v0.1.0/x',
    'https://github.com/someone-else/spec-harbor/releases/download/v0.1.0/x.tar.gz',
    'https://github.com/guferreira1/spec-harbor/releases/latest',
    'ftp://github.com/guferreira1/spec-harbor/releases/download/v0.1.0/x.tar.gz',
  ];
  for (const url of rejected) {
    assert.throws(() => assertAllowedInitialUrl(url), /non-allowlisted/);
  }
});

test('redirects must stay on HTTPS GitHub hosts', () => {
  assertAllowedRedirectUrl('https://objects.githubusercontent.com/some/asset');
  assertAllowedRedirectUrl('https://release-assets.githubusercontent.com/some/asset');
  assertAllowedRedirectUrl('https://github.com/guferreira1/spec-harbor/releases/download/v0.1.0/x');
  assert.throws(
    () => assertAllowedRedirectUrl('http://objects.githubusercontent.com/some/asset'),
    /non-HTTPS/
  );
  assert.throws(
    () => assertAllowedRedirectUrl('https://evil.example.com/some/asset'),
    /non-allowlisted host/
  );
  assert.throws(
    () => assertAllowedRedirectUrl('https://fakegithubusercontent.com/some/asset'),
    /non-allowlisted host/
  );
});

test('fetchBuffer downloads an allowlisted URL through the injected transport', async () => {
  const get = fakeGet({ [ASSET_URL]: { body: 'archive-bytes' } });
  const buffer = await fetchBuffer(ASSET_URL, { get });
  assert.equal(buffer.toString('utf8'), 'archive-bytes');
});

test('fetchBuffer follows HTTPS redirects to GitHub asset hosts', async () => {
  const redirected = 'https://objects.githubusercontent.com/real/asset';
  const get = fakeGet({
    [ASSET_URL]: { status: 302, headers: { location: redirected } },
    [redirected]: { body: 'redirected-bytes' },
  });
  const buffer = await fetchBuffer(ASSET_URL, { get });
  assert.equal(buffer.toString('utf8'), 'redirected-bytes');
});

test('fetchBuffer rejects redirects to plain HTTP', async () => {
  const get = fakeGet({
    [ASSET_URL]: {
      status: 302,
      headers: { location: 'http://objects.githubusercontent.com/real/asset' },
    },
  });
  await assert.rejects(fetchBuffer(ASSET_URL, { get }), /non-HTTPS/);
});

test('fetchBuffer rejects redirects to non-allowlisted hosts', async () => {
  const get = fakeGet({
    [ASSET_URL]: { status: 302, headers: { location: 'https://evil.example.com/asset' } },
  });
  await assert.rejects(fetchBuffer(ASSET_URL, { get }), /non-allowlisted host/);
});

test('fetchBuffer rejects non-200 responses', async () => {
  const get = fakeGet({ [ASSET_URL]: { status: 404 } });
  await assert.rejects(fetchBuffer(ASSET_URL, { get }), /HTTP 404/);
});

test('fetchBuffer rejects endless redirect loops', async () => {
  const loop = 'https://objects.githubusercontent.com/loop';
  const get = fakeGet({
    [ASSET_URL]: { status: 302, headers: { location: loop } },
    [loop]: { status: 302, headers: { location: loop } },
  });
  await assert.rejects(fetchBuffer(ASSET_URL, { get }), /too many redirects/);
});

test('parseChecksums reads sha256sum text and binary formats', () => {
  const hash = 'a'.repeat(64);
  const entries = parseChecksums(
    `${hash}  specharbor_Linux_x86_64.tar.gz\n${'b'.repeat(64)} *specharbor_Windows_x86_64.zip\nnot a checksum line\n`
  );
  assert.equal(entries.get('specharbor_Linux_x86_64.tar.gz'), hash);
  assert.equal(entries.get('specharbor_Windows_x86_64.zip'), 'b'.repeat(64));
  assert.equal(entries.size, 2);
});

test('checksum verification passes for matching content', () => {
  const content = Buffer.from('release archive content');
  const checksums = `${sha256Hex(content)}  asset.tar.gz\n`;
  verifyChecksum(content, checksums, 'asset.tar.gz');
});

test('checksum verification fails for tampered content', () => {
  const content = Buffer.from('release archive content');
  const checksums = `${sha256Hex(Buffer.from('different content'))}  asset.tar.gz\n`;
  assert.throws(
    () => verifyChecksum(content, checksums, 'asset.tar.gz'),
    /checksum verification failed/
  );
});

test('checksum verification fails when the asset has no checksum entry', () => {
  const content = Buffer.from('release archive content');
  const checksums = `${sha256Hex(content)}  another-asset.tar.gz\n`;
  assert.throws(() => verifyChecksum(content, checksums, 'asset.tar.gz'), /no checksum entry/);
});

test('sha256Hex matches Node crypto output', () => {
  const content = Buffer.from('abc');
  assert.equal(sha256Hex(content), crypto.createHash('sha256').update(content).digest('hex'));
});
