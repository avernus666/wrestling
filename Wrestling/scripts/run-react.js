const { spawnSync } = require("child_process");
const path = require("path");

const command = process.argv[2];
const allowedCommands = new Set(["start", "build"]);

if (!allowedCommands.has(command)) {
  console.error("Usage: node scripts/run-react.js <start|build>");
  process.exit(1);
}

const reactScripts = require.resolve("react-scripts/bin/react-scripts.js");
const nodeOptions = process.env.NODE_OPTIONS ? `${process.env.NODE_OPTIONS} --openssl-legacy-provider` : "--openssl-legacy-provider";
const result = spawnSync(process.execPath, [reactScripts, command], {
  stdio: "inherit",
  env: { ...process.env, NODE_OPTIONS: nodeOptions },
  cwd: path.resolve(__dirname, ".."),
});

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status ?? 1);
