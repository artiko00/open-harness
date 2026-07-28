#!/usr/bin/env node
'use strict';
// `open-harness init` delega en el init de cada tool: por cada uno crea su
// archivo de config (<tool>.json) en el directorio actual. No sobrescribe:
// si el archivo ya existe, lo reporta y sigue. El init de cada tool sigue
// siendo el dueño de generar su propio JSON; aquí solo lo orquestamos.

const { spawnSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const TOOLS = [
  { name: 'linelens', config: 'linelens.json' },
  { name: 'dupelens', config: 'dupelens.json' },
  { name: 'secretlens', config: 'secretlens.json' },
  { name: 'testlens', config: 'testlens.json' },
  { name: 'scopelens', config: 'scopelens.json' },
];

// Resuelve el bin (wrapper .js) del tool desde su propio paquete.
function toolBin(name) {
  return require.resolve(`@open_harness/${name}/bin/${name}.js`);
}

function runInit() {
  const cwd = process.cwd();
  const created = [];
  const existed = [];
  const failed = [];

  for (const tool of TOOLS) {
    const target = path.join(cwd, tool.config);
    if (fs.existsSync(target)) {
      existed.push(tool.config);
      continue;
    }

    let bin;
    try {
      bin = toolBin(tool.name);
    } catch {
      failed.push(`${tool.config} (paquete @open_harness/${tool.name} no instalado)`);
      continue;
    }

    // stdout ignorado para dar un reporte unificado; stderr se propaga.
    const res = spawnSync(process.execPath, [bin, 'init'], {
      stdio: ['inherit', 'ignore', 'inherit'],
    });

    if (res.status === 0 && fs.existsSync(target)) {
      created.push(tool.config);
    } else {
      failed.push(tool.config);
    }
  }

  console.log('open-harness init:');
  console.log(`  creados:    ${created.length ? created.join(', ') : '(ninguno)'}`);
  console.log(`  existentes: ${existed.length ? existed.join(', ') : '(ninguno)'}`);
  if (failed.length) {
    console.log(`  fallidos:   ${failed.join(', ')}`);
  }

  return failed.length ? 1 : 0;
}

function usage() {
  console.log('uso: open-harness <comando>');
  console.log('');
  console.log('comandos:');
  console.log('  init    crea los archivos de config de cada tool en el directorio actual');
  console.log('');
  console.log('tools: linelens, dupelens, secretlens, testlens, scopelens');
}

const sub = process.argv[2];
if (sub === 'init') {
  process.exit(runInit());
} else {
  usage();
  process.exit(sub === undefined || sub === '-h' || sub === '--help' ? 0 : 2);
}
