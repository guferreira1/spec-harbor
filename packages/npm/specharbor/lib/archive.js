'use strict';

const zlib = require('zlib');

// Minimal dependency-free extraction of one entry from the release archives.
// Archive integrity is guaranteed by the SHA-256 verification that runs
// before extraction, so these parsers only need to locate the binary entry.

function readTarString(buffer, offset, length) {
  const slice = buffer.subarray(offset, offset + length);
  const end = slice.indexOf(0);
  return slice.toString('utf8', 0, end === -1 ? length : end);
}

function extractTarGzEntry(archive, entryName) {
  const tar = zlib.gunzipSync(archive);
  let offset = 0;
  while (offset + 512 <= tar.length) {
    const header = tar.subarray(offset, offset + 512);
    if (header.every((byte) => byte === 0)) {
      break;
    }
    const name = readTarString(header, 0, 100);
    const prefix = readTarString(header, 345, 155);
    const fullName = prefix ? `${prefix}/${name}` : name;
    const size = parseInt(readTarString(header, 124, 12).trim() || '0', 8);
    const typeFlag = header[156];
    offset += 512;
    const isRegularFile = typeFlag === 0x30 || typeFlag === 0;
    if (isRegularFile && (fullName === entryName || fullName === `./${entryName}`)) {
      return Buffer.from(tar.subarray(offset, offset + size));
    }
    offset += Math.ceil(size / 512) * 512;
  }
  throw new Error(`entry ${entryName} not found in tar.gz archive`);
}

function extractZipEntry(archive, entryName) {
  // End of central directory record: signature 0x06054b50, minimum 22 bytes.
  let eocd = -1;
  for (let i = archive.length - 22; i >= 0; i -= 1) {
    if (archive.readUInt32LE(i) === 0x06054b50) {
      eocd = i;
      break;
    }
  }
  if (eocd === -1) {
    throw new Error('invalid zip archive: end of central directory not found');
  }
  const entryCount = archive.readUInt16LE(eocd + 10);
  let offset = archive.readUInt32LE(eocd + 16);
  for (let i = 0; i < entryCount; i += 1) {
    if (archive.readUInt32LE(offset) !== 0x02014b50) {
      throw new Error('invalid zip archive: malformed central directory');
    }
    const method = archive.readUInt16LE(offset + 10);
    const compressedSize = archive.readUInt32LE(offset + 20);
    const nameLength = archive.readUInt16LE(offset + 28);
    const extraLength = archive.readUInt16LE(offset + 30);
    const commentLength = archive.readUInt16LE(offset + 32);
    const localOffset = archive.readUInt32LE(offset + 42);
    const name = archive.toString('utf8', offset + 46, offset + 46 + nameLength);
    if (name === entryName) {
      if (archive.readUInt32LE(localOffset) !== 0x04034b50) {
        throw new Error('invalid zip archive: malformed local file header');
      }
      const localNameLength = archive.readUInt16LE(localOffset + 26);
      const localExtraLength = archive.readUInt16LE(localOffset + 28);
      const dataStart = localOffset + 30 + localNameLength + localExtraLength;
      const data = archive.subarray(dataStart, dataStart + compressedSize);
      if (method === 0) {
        return Buffer.from(data);
      }
      if (method === 8) {
        return zlib.inflateRawSync(data);
      }
      throw new Error(`unsupported zip compression method ${method} for ${entryName}`);
    }
    offset += 46 + nameLength + extraLength + commentLength;
  }
  throw new Error(`entry ${entryName} not found in zip archive`);
}

function extractEntry(archive, format, entryName) {
  if (format === 'tar.gz') {
    return extractTarGzEntry(archive, entryName);
  }
  if (format === 'zip') {
    return extractZipEntry(archive, entryName);
  }
  throw new Error(`unsupported archive format: ${format}`);
}

module.exports = {
  extractTarGzEntry,
  extractZipEntry,
  extractEntry,
};
