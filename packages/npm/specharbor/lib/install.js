'use strict';

const fs = require('fs');
const path = require('path');

const platformModule = require('./platform');
const downloadModule = require('./download');
const archiveModule = require('./archive');

const PACKAGE_ROOT = path.join(__dirname, '..');
const NATIVE_DIR = path.join(PACKAGE_ROOT, 'native');

function pinnedVersion() {
  // package.json is the single source of the pinned release version: npm
  // package version X.Y.Z maps to exactly the GitHub Release tag vX.Y.Z.
  const manifest = JSON.parse(
    fs.readFileSync(path.join(PACKAGE_ROOT, 'package.json'), 'utf8')
  );
  return manifest.version;
}

function binaryPath(descriptor, nativeDir = NATIVE_DIR) {
  return path.join(nativeDir, descriptor.binaryName);
}

// Resolves the native binary path, downloading and verifying it when missing.
// Used by postinstall and as the first-run fallback for --ignore-scripts
// installs. All writes stay inside the package's own native/ directory.
async function ensureBinary({
  platform = process.platform,
  arch = process.arch,
  fetchBuffer = downloadModule.fetchBuffer,
  nativeDir = NATIVE_DIR,
  log = (message) => process.stderr.write(`${message}\n`),
} = {}) {
  const descriptor = platformModule.resolvePlatform(platform, arch);
  const target = binaryPath(descriptor, nativeDir);
  if (fs.existsSync(target)) {
    return target;
  }

  const tag = platformModule.releaseTag(pinnedVersion());
  const name = platformModule.assetName(descriptor);
  const archiveUrl = platformModule.assetUrl(tag, name);
  const checksumsUrl = platformModule.checksumsUrl(tag);

  log(`specharbor: downloading ${archiveUrl}`);
  const archive = await fetchBuffer(archiveUrl);
  log(`specharbor: downloading ${checksumsUrl}`);
  const checksums = await fetchBuffer(checksumsUrl);

  log(`specharbor: verifying SHA-256 checksum for ${name}`);
  downloadModule.verifyChecksum(archive, checksums.toString('utf8'), name);

  const binary = archiveModule.extractEntry(archive, descriptor.format, descriptor.binaryName);

  fs.mkdirSync(nativeDir, { recursive: true });
  const partial = `${target}.partial-${process.pid}`;
  try {
    fs.writeFileSync(partial, binary);
    if (platform !== 'win32') {
      fs.chmodSync(partial, 0o755);
    }
    fs.renameSync(partial, target);
  } catch (error) {
    fs.rmSync(partial, { force: true });
    throw error;
  }
  log(`specharbor: installed native binary at ${target}`);
  return target;
}

module.exports = {
  NATIVE_DIR,
  pinnedVersion,
  binaryPath,
  ensureBinary,
};
