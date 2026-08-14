#!/usr/bin/env node
'use strict';

const path = require('path');
const bin = require('./binary.js');
const pkg = require('./package.json');

const version = bin.resolveVersion(pkg.version);

bin
  .installBinary(path.join(__dirname, 'vendor'), version)
  .then((dest) => console.log(`envx: installed v${version} → ${dest}`))
  .catch((err) => {
    console.error(`envx: install failed: ${err.message}`);
    process.exit(1);
  });
