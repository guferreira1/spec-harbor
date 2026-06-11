'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');

const { extractTarGzEntry, extractZipEntry, extractEntry } = require('../lib/archive');
const { makeTarGz, makeZip } = require('./helpers');

test('extracts the specharbor binary from a tar.gz archive', () => {
  const archive = makeTarGz([
    { name: 'LICENSE', data: 'license text' },
    { name: 'specharbor', data: 'binary-bytes' },
  ]);
  assert.equal(extractTarGzEntry(archive, 'specharbor').toString('utf8'), 'binary-bytes');
});

test('tar.gz extraction fails when the binary entry is missing', () => {
  const archive = makeTarGz([{ name: 'LICENSE', data: 'license text' }]);
  assert.throws(() => extractTarGzEntry(archive, 'specharbor'), /not found/);
});

test('extracts specharbor.exe from a stored zip archive', () => {
  const archive = makeZip([
    { name: 'LICENSE', data: 'license text' },
    { name: 'specharbor.exe', data: 'exe-bytes' },
  ]);
  assert.equal(extractZipEntry(archive, 'specharbor.exe').toString('utf8'), 'exe-bytes');
});

test('extracts specharbor.exe from a deflate zip archive', () => {
  const archive = makeZip([{ name: 'specharbor.exe', data: 'deflated-exe-bytes' }], {
    method: 8,
  });
  assert.equal(extractZipEntry(archive, 'specharbor.exe').toString('utf8'), 'deflated-exe-bytes');
});

test('zip extraction fails when the binary entry is missing', () => {
  const archive = makeZip([{ name: 'LICENSE', data: 'license text' }]);
  assert.throws(() => extractZipEntry(archive, 'specharbor.exe'), /not found/);
});

test('extractEntry dispatches by archive format and rejects unknown formats', () => {
  const tarGz = makeTarGz([{ name: 'specharbor', data: 'tar-bytes' }]);
  const zip = makeZip([{ name: 'specharbor.exe', data: 'zip-bytes' }]);
  assert.equal(extractEntry(tarGz, 'tar.gz', 'specharbor').toString('utf8'), 'tar-bytes');
  assert.equal(extractEntry(zip, 'zip', 'specharbor.exe').toString('utf8'), 'zip-bytes');
  assert.throws(() => extractEntry(tarGz, 'rar', 'specharbor'), /unsupported archive format/);
});
