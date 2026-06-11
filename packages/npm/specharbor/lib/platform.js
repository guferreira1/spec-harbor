'use strict';

// Maps Node's process.platform/process.arch to the finalized GoReleaser
// release asset matrix for guferreira1/spec-harbor.

const DOWNLOAD_BASE_URL =
  'https://github.com/guferreira1/spec-harbor/releases/download/';
const INSTALL_DOCS_URL =
  'https://github.com/guferreira1/spec-harbor/blob/main/docs/install.md';
const CHECKSUMS_NAME = 'checksums.txt';

const SUPPORTED_PLATFORMS = {
  'linux x64': { os: 'Linux', arch: 'x86_64', format: 'tar.gz', binaryName: 'specharbor' },
  'linux arm64': { os: 'Linux', arch: 'arm64', format: 'tar.gz', binaryName: 'specharbor' },
  'darwin x64': { os: 'Darwin', arch: 'x86_64', format: 'tar.gz', binaryName: 'specharbor' },
  'darwin arm64': { os: 'Darwin', arch: 'arm64', format: 'tar.gz', binaryName: 'specharbor' },
  'win32 x64': { os: 'Windows', arch: 'x86_64', format: 'zip', binaryName: 'specharbor.exe' },
  'win32 arm64': { os: 'Windows', arch: 'arm64', format: 'zip', binaryName: 'specharbor.exe' },
};

class UnsupportedPlatformError extends Error {
  constructor(platform, arch) {
    super(
      `unsupported platform: ${platform} ${arch}. ` +
        `Supported: Linux, macOS, and Windows on x64/arm64. ` +
        `See ${INSTALL_DOCS_URL} for manual installation options.`
    );
    this.name = 'UnsupportedPlatformError';
    this.platform = platform;
    this.arch = arch;
  }
}

function resolvePlatform(platform = process.platform, arch = process.arch) {
  const descriptor = SUPPORTED_PLATFORMS[`${platform} ${arch}`];
  if (!descriptor) {
    throw new UnsupportedPlatformError(platform, arch);
  }
  return descriptor;
}

function releaseTag(packageVersion) {
  if (!/^\d+\.\d+\.\d+$/.test(packageVersion)) {
    throw new Error(
      `invalid package version "${packageVersion}"; expected X.Y.Z so it maps to exactly one release tag vX.Y.Z`
    );
  }
  return `v${packageVersion}`;
}

function assetName(descriptor) {
  return `specharbor_${descriptor.os}_${descriptor.arch}.${descriptor.format}`;
}

function assetUrl(tag, name) {
  return `${DOWNLOAD_BASE_URL}${tag}/${name}`;
}

function checksumsUrl(tag) {
  return assetUrl(tag, CHECKSUMS_NAME);
}

module.exports = {
  DOWNLOAD_BASE_URL,
  INSTALL_DOCS_URL,
  CHECKSUMS_NAME,
  SUPPORTED_PLATFORMS,
  UnsupportedPlatformError,
  resolvePlatform,
  releaseTag,
  assetName,
  assetUrl,
  checksumsUrl,
};
