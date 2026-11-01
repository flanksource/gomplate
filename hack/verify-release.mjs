import assert from "node:assert/strict";
import { readFileSync, readdirSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { createRequire } from "node:module";
const require = createRequire(new URL("../.tmp/release-check/package.json", import.meta.url));
const { parse } = require("yaml");
const semver = require("semver");
const { analyzeCommits } = await import(require.resolve("@semantic-release/commit-analyzer"));
const config = parse(readFileSync(".releaserc", "utf8"));
const options = config.plugins[0][1];
const logger = { log() {} };
for (const [message, expected] of [
 ["feat!: migrate gomplate and CEL module paths\n\nBREAKING CHANGE: use /v4 and cel.dev/cel-go imports", "major"],
 ["fix(security): remediate dependency advisory", "patch"],
]) {
 const kind = await analyzeCommits(options, { commits: [{hash:"example",message}], logger, cwd:process.cwd() });
 assert.equal(kind, expected);
 assert.equal(semver.inc("3.24.90", kind), kind === "major" ? "4.0.0" : "3.24.91");
}
for (const file of ["test.yml", "lint.yml", "web.yml", "release.yml", "security.yml"]) {
 const workflow = parse(readFileSync(".github/workflows/" + file, "utf8"));
 assert(workflow.on.push.branches.includes("3.x"), file);
}
const release = parse(readFileSync(".github/workflows/release.yml", "utf8"));
for (const file of readdirSync(".github/workflows").filter(file => file.endsWith(".yml"))) {
 const workflow = parse(readFileSync(".github/workflows/" + file, "utf8"));
 for (const job of Object.values(workflow.jobs)) {
  for (const step of job.steps ?? []) {
   if (step.run) execFileSync("bash", ["-n"], { input: step.run });
  }
 }
}
assert(release.jobs["bump-clients"].if.includes("startsWith(needs.semantic-release.outputs.release-version, '3.')"));
process.stdout.write("Release versions, workflow YAML, and downstream guard passed.\n");
