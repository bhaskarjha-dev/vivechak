#!/usr/bin/env node
const path = require('path');
const fs = require('fs');

// Try resolving vivechak from local sibling (in monorepo) or standard require resolution
let runnerPath = path.resolve(__dirname, '..', '..', 'vivechak', 'bin', 'runner.js');
if (!fs.existsSync(runnerPath)) {
  try {
    runnerPath = require.resolve('vivechak/bin/runner.js');
  } catch (_) {
    runnerPath = path.resolve(__dirname, '..', 'node_modules', 'vivechak', 'bin', 'runner.js');
  }
}

require(runnerPath);
