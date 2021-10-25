#!/usr/bin/env node
// Copyright 2020 The Bazel Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
'use strict';

// This package inspired by
// https://github.com/angular/clang-format/blob/master/index.js
const os = require('os');
const path = require('path');
const spawn = require('child_process').spawn;

function getNativeBinary() {
  const arch = {
    'arm64' : 'arm64',
    'x64' : 'amd64',
  }[os.arch()];
  // Filter the platform based on the platforms that are build/included.
  const platform = {
    'darwin' : 'darwin',
    'linux' : 'linux',
    'win32' : 'windows',
  }[os.platform()];
  const extension = {
    'darwin' : '',
    'linux' : '',
    'win32' : '.exe',
  }[os.platform()];

  if (arch == undefined || platform == undefined || (arch == 'arm64' && platform == 'windows')) {
    console.error(`FATAL: Your platform/architecture combination ${
        os.platform()} - ${os.arch()} is not yet supported.
    See instructions at https://github.com/bazelbuild/buildtools/blob/master/_TOOL_/README.md.`);
    return Promise.resolve(1);
  }

  const binary =
      path.join(__dirname, `_TOOL_-${platform}_${arch}${extension}`);
  return binary;
}

function main(args) {
  const binary = getNativeBinary();
  const ps = spawn(binary, args, {stdio : 'inherit'});

  function shutdown() {
    ps.kill("SIGTERM")
    process.exit();
  }

  process.on("SIGINT", shutdown);
  process.on("SIGTERM", shutdown);

  ps.on('close', e => process.exitCode = e);
}

if (require.main === module) {
  main(process.argv.slice(2));
}

module.exports = {
  getNativeBinary,
};
