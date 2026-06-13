#!/usr/bin/env node
'use strict';
// Meta shim: re-export the secretlens wrapper bin so that `secretlens` is available
// in node_modules/.bin/ even when @open_harness/open-harness is installed as the
// only direct dependency. Without this, package managers that don't hoist
// transitive bins (pnpm, yarn Berry) would not expose the command.
require('@open_harness/secretlens/bin/secretlens.js');
