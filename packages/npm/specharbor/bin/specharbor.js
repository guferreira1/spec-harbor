#!/usr/bin/env node
'use strict';

const { spawnSync } = require('child_process');

const { ensureBinary } = require('../lib/install');
const { INSTALL_DOCS_URL } = require('../lib/platform');

async function main() {
  let binary;
  try {
    // First-run fallback: when postinstall was skipped (--ignore-scripts),
    // this performs the same checksum-verified download before executing.
    binary = await ensureBinary();
  } catch (error) {
    process.stderr.write(`specharbor: ${error.message}\n`);
    process.stderr.write(`specharbor: manual installation options: ${INSTALL_DOCS_URL}\n`);
    process.exit(1);
  }

  // Arguments are forwarded as an array; no shell is involved, so there is
  // no shell-injection surface.
  const result = spawnSync(binary, process.argv.slice(2), { stdio: 'inherit' });
  if (result.error) {
    process.stderr.write(`specharbor: failed to run native binary: ${result.error.message}\n`);
    process.exit(1);
  }
  if (result.signal) {
    process.kill(process.pid, result.signal);
    return;
  }
  process.exit(result.status === null ? 1 : result.status);
}

main();
