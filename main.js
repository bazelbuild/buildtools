const { getNativeBinary } = require("./launcher");
function main(args) {
  const binary = getNativeBinary();
  const ps = spawn(binary, args, { stdio: 'inherit' });
  function shutdown() {
    ps.kill("SIGTERM");
    process.exit();
  }
  process.on("SIGINT", shutdown);
  process.on("SIGTERM", shutdown);
  ps.on('close', e => process.exitCode = e);
}
exports.main = main;
