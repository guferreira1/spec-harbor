'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('fs');
const path = require('path');

const PACKAGE_ROOT = path.join(__dirname, '..');

function readManifest() {
  return JSON.parse(fs.readFileSync(path.join(PACKAGE_ROOT, 'package.json'), 'utf8'));
}

function sourceFiles() {
  const files = [];
  for (const dir of ['bin', 'lib', 'scripts']) {
    for (const name of fs.readdirSync(path.join(PACKAGE_ROOT, dir))) {
      if (name.endsWith('.js')) {
        files.push(path.join(PACKAGE_ROOT, dir, name));
      }
    }
  }
  return files;
}

test('package is the unscoped specharbor package exposing the specharbor command', () => {
  const manifest = readManifest();
  assert.equal(manifest.name, 'specharbor');
  assert.equal(manifest.bin.specharbor, 'bin/specharbor.js');
  assert.match(manifest.version, /^\d+\.\d+\.\d+$/);
});

test('postinstall download hook is declared and no publish automation exists', () => {
  const manifest = readManifest();
  assert.equal(manifest.scripts.postinstall, 'node scripts/postinstall.js');
  for (const hook of ['publish', 'prepublish', 'prepublishOnly', 'prepack', 'postpublish']) {
    assert.equal(manifest.scripts[hook], undefined, `unexpected ${hook} script`);
  }
});

test('package ships the wrapper, scripts, and both READMEs but no binaries or fixtures', () => {
  const manifest = readManifest();
  const files = manifest.files;
  assert.ok(Array.isArray(files), 'package.json must declare a files array');

  for (const required of ['bin', 'lib', 'scripts', 'README.md', 'README.pt-BR.md', 'package.json']) {
    if (required === 'package.json') {
      // package.json is always included by npm; it need not be listed.
      assert.ok(fs.existsSync(path.join(PACKAGE_ROOT, required)), 'package.json must exist');
      continue;
    }
    assert.ok(files.includes(required), `files must include ${required}`);
    assert.ok(fs.existsSync(path.join(PACKAGE_ROOT, required)), `${required} must exist on disk`);
  }

  // The English README links to the Portuguese README, so the package must
  // ship it; a published tarball that omits it would have a broken link.
  const readme = fs.readFileSync(path.join(PACKAGE_ROOT, 'README.md'), 'utf8');
  assert.ok(readme.includes('README.pt-BR.md'), 'README.md must link to README.pt-BR.md');

  // Downloaded native binaries, installed modules, and test fixtures must never
  // be published.
  for (const forbidden of ['native', 'node_modules', 'test']) {
    assert.ok(!files.includes(forbidden), `files must not include ${forbidden}`);
  }
});

test('package has no runtime dependencies', () => {
  const manifest = readManifest();
  assert.equal(manifest.dependencies, undefined);
  assert.equal(manifest.devDependencies, undefined);
  assert.equal(manifest.optionalDependencies, undefined);
});

test('no lockfile is committed for the wrapper package', () => {
  for (const lockfile of ['package-lock.json', 'npm-shrinkwrap.json', 'yarn.lock']) {
    assert.ok(
      !fs.existsSync(path.join(PACKAGE_ROOT, lockfile)),
      `${lockfile} must not be committed`
    );
  }
});

test('wrapper sources never build shell command strings', () => {
  for (const file of sourceFiles()) {
    const source = fs.readFileSync(file, 'utf8');
    // Excludes `.exec(` (RegExp.prototype.exec); targets child_process exec forms.
    assert.ok(
      !/(^|[^.\w])exec(Sync|File|FileSync)?\s*\(/.test(source),
      `${file} must not use child_process exec functions`
    );
    assert.ok(!/shell\s*:\s*true/.test(source), `${file} must not spawn through a shell`);
  }
});

test('wrapper sources contain no plain-http URLs and no tokens', () => {
  for (const file of sourceFiles()) {
    const source = fs.readFileSync(file, 'utf8');
    assert.ok(!source.includes('http://'), `${file} must not contain plain-http URLs`);
    assert.ok(!/GITHUB_TOKEN|NPM_TOKEN|Authorization/i.test(source), `${file} must not use tokens`);
  }
});

test('the only download host in wrapper sources is the official release URL', () => {
  const urlPattern = /https:\/\/[a-z0-9.-]+\/[^\s'"`]*/gi;
  const allowedPrefixes = [
    'https://github.com/guferreira1/spec-harbor',
  ];
  for (const file of sourceFiles()) {
    const source = fs.readFileSync(file, 'utf8');
    for (const url of source.match(urlPattern) || []) {
      assert.ok(
        allowedPrefixes.some((prefix) => url.startsWith(prefix)),
        `${file} references non-allowlisted URL ${url}`
      );
    }
  }
});
