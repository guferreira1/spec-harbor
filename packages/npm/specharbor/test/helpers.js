'use strict';

const { EventEmitter } = require('events');
const zlib = require('zlib');

// Builds a ustar tar.gz archive in memory for extraction tests.
function makeTarGz(entries) {
  const blocks = [];
  for (const entry of entries) {
    const data = Buffer.from(entry.data);
    const header = Buffer.alloc(512);
    header.write(entry.name, 0, 100, 'utf8');
    header.write('0000755\0', 100, 8, 'utf8');
    header.write('0000000\0', 108, 8, 'utf8');
    header.write('0000000\0', 116, 8, 'utf8');
    header.write(`${data.length.toString(8).padStart(11, '0')} `, 124, 12, 'utf8');
    header.write('00000000000 ', 136, 12, 'utf8');
    header.write('        ', 148, 8, 'utf8');
    header.write('0', 156, 1, 'utf8');
    header.write('ustar', 257, 5, 'utf8');
    header.write('00', 263, 2, 'utf8');
    let checksum = 0;
    for (const byte of header) {
      checksum += byte;
    }
    header.write(`${checksum.toString(8).padStart(6, '0')}\0 `, 148, 8, 'utf8');
    blocks.push(header);
    blocks.push(data);
    const padding = (512 - (data.length % 512)) % 512;
    if (padding > 0) {
      blocks.push(Buffer.alloc(padding));
    }
  }
  blocks.push(Buffer.alloc(1024));
  return zlib.gzipSync(Buffer.concat(blocks));
}

// Builds a zip archive in memory. Supported methods: 0 (stored), 8 (deflate).
// CRC fields are zeroed: archive integrity is covered by SHA-256 verification
// in the code under test, and the extractor does not read CRCs.
function makeZip(entries, { method = 0 } = {}) {
  const localParts = [];
  const centralParts = [];
  let offset = 0;
  for (const entry of entries) {
    const data = Buffer.from(entry.data);
    const stored = method === 8 ? zlib.deflateRawSync(data) : data;
    const name = Buffer.from(entry.name, 'utf8');

    const local = Buffer.alloc(30);
    local.writeUInt32LE(0x04034b50, 0);
    local.writeUInt16LE(20, 4);
    local.writeUInt16LE(0, 6);
    local.writeUInt16LE(method, 8);
    local.writeUInt32LE(0, 10);
    local.writeUInt32LE(0, 14);
    local.writeUInt32LE(stored.length, 18);
    local.writeUInt32LE(data.length, 22);
    local.writeUInt16LE(name.length, 26);
    local.writeUInt16LE(0, 28);

    const central = Buffer.alloc(46);
    central.writeUInt32LE(0x02014b50, 0);
    central.writeUInt16LE(20, 4);
    central.writeUInt16LE(20, 6);
    central.writeUInt16LE(0, 8);
    central.writeUInt16LE(method, 10);
    central.writeUInt32LE(0, 12);
    central.writeUInt32LE(0, 16);
    central.writeUInt32LE(stored.length, 20);
    central.writeUInt32LE(data.length, 24);
    central.writeUInt16LE(name.length, 28);
    central.writeUInt16LE(0, 30);
    central.writeUInt16LE(0, 32);
    central.writeUInt16LE(0, 34);
    central.writeUInt16LE(0, 36);
    central.writeUInt32LE(0, 38);
    central.writeUInt32LE(offset, 42);

    localParts.push(local, name, stored);
    centralParts.push(central, name);
    offset += 30 + name.length + stored.length;
  }
  const centralDirectory = Buffer.concat(centralParts);
  const eocd = Buffer.alloc(22);
  eocd.writeUInt32LE(0x06054b50, 0);
  eocd.writeUInt16LE(0, 4);
  eocd.writeUInt16LE(0, 6);
  eocd.writeUInt16LE(entries.length, 8);
  eocd.writeUInt16LE(entries.length, 10);
  eocd.writeUInt32LE(centralDirectory.length, 12);
  eocd.writeUInt32LE(offset, 16);
  eocd.writeUInt16LE(0, 20);
  return Buffer.concat([...localParts, centralDirectory, eocd]);
}

// In-memory https.get replacement: routes is { url: { status, headers, body } }.
// Tests use this so no test ever opens a network connection.
function fakeGet(routes) {
  return (url, callback) => {
    const request = new EventEmitter();
    process.nextTick(() => {
      const route = routes[url];
      if (!route) {
        request.emit('error', new Error(`fakeGet: no route for ${url}`));
        return;
      }
      const response = new EventEmitter();
      response.statusCode = route.status || 200;
      response.headers = route.headers || {};
      response.resume = () => {};
      callback(response);
      process.nextTick(() => {
        if (route.body !== undefined) {
          response.emit('data', Buffer.from(route.body));
        }
        response.emit('end');
      });
    });
    return request;
  };
}

module.exports = {
  makeTarGz,
  makeZip,
  fakeGet,
};
