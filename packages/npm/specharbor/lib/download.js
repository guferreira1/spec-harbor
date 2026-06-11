'use strict';

const https = require('https');
const crypto = require('crypto');

const { DOWNLOAD_BASE_URL } = require('./platform');

// Redirect targets must stay on HTTPS and on GitHub-controlled hosts.
// GitHub serves release asset downloads through *.githubusercontent.com.
const ALLOWED_REDIRECT_HOSTS = /^(github\.com|([a-z0-9-]+\.)*githubusercontent\.com)$/;
const MAX_REDIRECTS = 5;

function assertAllowedInitialUrl(url) {
  if (!url.startsWith(DOWNLOAD_BASE_URL)) {
    throw new Error(
      `refusing to download from non-allowlisted URL: ${url} (allowed prefix: ${DOWNLOAD_BASE_URL})`
    );
  }
}

function assertAllowedRedirectUrl(url) {
  const parsed = new URL(url);
  if (parsed.protocol !== 'https:') {
    throw new Error(`refusing non-HTTPS redirect: ${url}`);
  }
  if (!ALLOWED_REDIRECT_HOSTS.test(parsed.hostname)) {
    throw new Error(`refusing redirect to non-allowlisted host: ${parsed.hostname}`);
  }
}

// Downloads a URL into memory. The initial URL must be an official release
// download URL; redirects must stay on HTTPS GitHub hosts. `get` is
// injectable so tests never touch the network.
function fetchBuffer(url, { get = https.get, redirects = 0, isRedirect = false } = {}) {
  try {
    if (isRedirect) {
      assertAllowedRedirectUrl(url);
    } else {
      assertAllowedInitialUrl(url);
    }
  } catch (error) {
    // Always reject instead of throwing: redirect checks run inside response
    // callbacks, where a synchronous throw could not be caught by callers.
    return Promise.reject(error);
  }
  return new Promise((resolve, reject) => {
    const request = get(url, (response) => {
      const status = response.statusCode;
      if ([301, 302, 303, 307, 308].includes(status)) {
        response.resume();
        if (redirects >= MAX_REDIRECTS) {
          reject(new Error(`too many redirects while downloading ${url}`));
          return;
        }
        const location = response.headers.location;
        if (!location) {
          reject(new Error(`redirect without Location header from ${url}`));
          return;
        }
        const nextUrl = new URL(location, url).toString();
        fetchBuffer(nextUrl, { get, redirects: redirects + 1, isRedirect: true }).then(
          resolve,
          reject
        );
        return;
      }
      if (status !== 200) {
        response.resume();
        reject(new Error(`download failed with HTTP ${status}: ${url}`));
        return;
      }
      const chunks = [];
      response.on('data', (chunk) => chunks.push(chunk));
      response.on('end', () => resolve(Buffer.concat(chunks)));
      response.on('error', reject);
    });
    request.on('error', reject);
  });
}

function sha256Hex(buffer) {
  return crypto.createHash('sha256').update(buffer).digest('hex');
}

// Parses sha256sum-style lines: "<hex>  <name>" or "<hex> *<name>".
function parseChecksums(text) {
  const entries = new Map();
  for (const line of text.split('\n')) {
    const match = /^([0-9a-fA-F]{64})[ \t]+\*?(.+)$/.exec(line.trim());
    if (match) {
      entries.set(match[2], match[1].toLowerCase());
    }
  }
  return entries;
}

function verifyChecksum(buffer, checksumsText, name) {
  const entries = parseChecksums(checksumsText);
  const expected = entries.get(name);
  if (!expected) {
    throw new Error(`no checksum entry for ${name} in checksums.txt; refusing to install`);
  }
  const actual = sha256Hex(buffer);
  if (actual !== expected) {
    throw new Error(
      `checksum verification failed for ${name} (expected ${expected}, got ${actual}); aborting`
    );
  }
}

module.exports = {
  MAX_REDIRECTS,
  assertAllowedInitialUrl,
  assertAllowedRedirectUrl,
  fetchBuffer,
  sha256Hex,
  parseChecksums,
  verifyChecksum,
};
