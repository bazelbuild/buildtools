const dir = require('path').join(
    process.env['TEST_SRCDIR'],
    process.env['BAZEL_WORKSPACE'],
    'buildifier/npm/buildifier');
process.chdir(dir);
const {stdout} = require('child_process').spawnSync(
    process.argv0,
    ['./buildifier.js', '--help'],
    {encoding: 'utf-8'});
if (!/usage: buildifier/.test(stdout)) {
    throw new Error('buildifier --help should include usage: buildifier');
}
