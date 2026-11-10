import assert from "node:assert/strict";
import { readFileSync, readdirSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { createRequire } from "node:module";
const require = createRequire(new URL("../.tmp/release-check/package.json", import.meta.url));
const { parse } = require("yaml");
const semver = require("semver");
const { analyzeCommits } = await import(require.resolve("@semantic-release/commit-analyzer"));
const normalize = await import(require.resolve("semantic-release/lib/branches/normalize.js"));
const definitions = await import(require.resolve("semantic-release/lib/definitions/branches.js"));
const config = parse(readFileSync(".releaserc", "utf8"));
const options = config.plugins[0][1];
const logger = { log() {} };
const git = (...args) => execFileSync("git", args, { encoding: "utf8" }).trim();
const target = process.argv[2] ?? "main";
const revision = process.argv[3] ?? "HEAD";
const branches = config.branches.map((branch) => {
 const options = typeof branch === "string" ? { name: branch } : branch;
 const tags = git("tag", "--merged", `origin/${options.name}`).split("\n")
  .map((gitTag) => ({ gitTag, version: semver.valid(gitTag), channels: [null] }))
  .filter(({ version }) => version && !semver.prerelease(version));
 return { ...options, tags };
});
function normalizeBranches(branches) {
 const groups = {
  maintenance: branches.filter(definitions.maintenance.filter),
  release: branches.filter(definitions.release.filter),
 };
 return [...normalize.maintenance(groups), ...normalize.release(groups)];
}
const normalized = normalizeBranches(branches);
const fixtureTags = (versions) => versions.map((version) => ({ version, gitTag: `v${version}`, channels: [null] }));
for (const scenario of [
 { main: ["3.24.90"], maintenance: ["3.24.90"], target: "main", candidate: "3.24.91", allowed: true },
 { main: ["3.24.90"], maintenance: ["3.24.90"], target: "3.x", candidate: "3.24.91", allowed: false },
 { main: ["3.24.90"], maintenance: ["3.24.90"], target: "main", candidate: "4.0.0", allowed: true },
 { main: ["3.24.90", "3.24.91", "4.0.0"], maintenance: ["3.24.90", "3.24.91"], target: "3.x", candidate: "3.24.92", allowed: true },
]) {
 const fixtureBranches = branches.map((branch) => ({ ...branch,
  tags: fixtureTags(branch.name === "main" ? scenario.main : scenario.maintenance),
 }));
 const branch = normalizeBranches(fixtureBranches).find(({ name }) => name === scenario.target);
 assert.equal(semver.satisfies(scenario.candidate, branch.range), scenario.allowed, JSON.stringify(scenario));
}
const branch = normalized.find(({ name }) => name === target);
assert(branch, `Unknown release branch: ${target}`);
const lastVersion = semver.rsort(branch.tags.map(({ version }) => version))[0];
assert(lastVersion, `No published release found on ${target}`);
const lastTag = branch.tags.find(({ version }) => version === lastVersion).gitTag;
const commits = git("log", "--format=%B%x00", `${lastTag}..${revision}`).split("\0")
 .map((message, index) => ({ hash: String(index), message: message.trim() }))
 .filter(({ message }) => message);
const kind = await analyzeCommits(options, { commits, logger, cwd: process.cwd() });
if (kind) {
 const candidate = semver.inc(lastVersion, kind);
 assert(semver.satisfies(candidate, branch.range),
  `${target} cannot publish ${candidate}: allowed range is ${branch.range}. Publish v3 fixes from main until main publishes v4.`);
 const moduleMajor = readFileSync("go.mod", "utf8").match(/^module github.com\/flanksource\/gomplate\/v(\d+)$/m);
 assert(moduleMajor, "Cannot determine gomplate module major");
 assert.equal(semver.major(candidate), Number(moduleMajor[1]), "Release major must match the Go module");
 process.stdout.write(`${target} can publish ${candidate} (${branch.range})\n`);
} else {
 process.stdout.write(`${target}: no unreleased conventional commits\n`);
}
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
assert.equal(release.jobs["semantic-release"].steps.find(({ id }) => id === "semantic").with.semantic_version,
 require("semantic-release/package.json").version, "Verification and publishing must use the same semantic-release version");
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
