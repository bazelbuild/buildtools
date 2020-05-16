const dir = require('path').join(
    process.env['TEST_SRCDIR'],
    process.env['BAZEL_WORKSPACE'],
    'buildozer/npm/buildozer');
process.chdir(dir);
const {stderr} = require('child_process').spawnSync(
    process.argv0,
    ['./buildozer.js', '--help'],
    {encoding: 'utf-8'});
if (!/Usage of .*buildozer/.test(stderr)) {
    throw new Error('buildozer --help should include usage: buildifier');
}

process.chdir(process.env['TEST_TMPDIR']);
const buildozer = require(dir);
const fs = require('fs');
fs.mkdirSync('foo');
fs.writeFileSync('foo/BUILD', '');
buildozer.run({commands: ['new_load //:some.bzl some_rule'], targets: ['//foo:__pkg__']});
const content = fs.readFileSync('foo/BUILD', 'utf-8');
if (!content.includes('load("//:some.bzl", "some_rule")')) {
    throw new Error('buildozer generated file should include load statement');
}
