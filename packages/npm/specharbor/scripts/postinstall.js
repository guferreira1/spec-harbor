'use strict';

const { ensureBinary } = require('../lib/install');
const { INSTALL_DOCS_URL } = require('../lib/platform');

ensureBinary()
  .then((binary) => {
    process.stdout.write(`specharbor: native binary ready at ${binary}\n`);
  })
  .catch((error) => {
    process.stderr.write(`specharbor: postinstall failed: ${error.message}\n`);
    process.stderr.write(`specharbor: manual installation options: ${INSTALL_DOCS_URL}\n`);
    process.exit(1);
  });
