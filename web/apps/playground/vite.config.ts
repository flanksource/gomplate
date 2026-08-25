import { existsSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig, type UserConfig } from "vite";

import { evalServer, findAvailablePort } from "./plugins/eval-server";

const root = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(root, "../../..");

const EVAL_HOST = "127.0.0.1";
const EVAL_PORT = 8321;
const DEV_PORT = 5280;

// A sibling clicky-ui checkout is aliased to its sources during `vite dev`, so
// UI changes show up without a publish round-trip. Absent (CI, a clean clone),
// the published package is used instead.
// Tests are excluded on purpose: they assert against the surface the package
// publishes, not against whatever a sibling checkout happens to have mid-edit.
const clickySrc = resolve(repoRoot, "../clicky-ui/packages/ui/src");
const clickySourceAvailable = existsSync(clickySrc) && process.env.GOMPLATE_CLICKY_SOURCE !== "0";

// Every subpath the app imports has to be aliased, not just the ones that are
// convenient: mixing aliased sources with the published bundle loads clicky-ui
// twice, and React context does not cross module instances -- a RouterProvider
// from one copy is invisible to an AppShell from the other.
const clickyAliases = [
  { find: /^@flanksource\/clicky-ui\/styles\.css$/, replacement: resolve(clickySrc, "styles/full.css") },
  { find: /^@flanksource\/clicky-ui\/tailwind-preset$/, replacement: resolve(clickySrc, "tailwind-preset.ts") },
  { find: /^@flanksource\/clicky-ui\/monaco$/, replacement: resolve(clickySrc, "monaco.ts") },
  { find: /^@flanksource\/clicky-ui\/icons$/, replacement: resolve(clickySrc, "icons.ts") },
  { find: /^@flanksource\/clicky-ui\/data$/, replacement: resolve(clickySrc, "data.ts") },
  { find: /^@flanksource\/clicky-ui\/components$/, replacement: resolve(clickySrc, "components.ts") },
  { find: /^@flanksource\/clicky-ui\/hooks$/, replacement: resolve(clickySrc, "hooks.ts") },
  { find: /^@flanksource\/clicky-ui\/rpc$/, replacement: resolve(clickySrc, "rpc.ts") },
  { find: /^@flanksource\/clicky-ui$/, replacement: resolve(clickySrc, "index.ts") },
];

interface PlaygroundConfigOptions {
  command: "build" | "serve";
  mode: string;
  clickySourceAvailable: boolean;
  evalPort: number;
}

export function createPlaygroundConfig({
  command,
  mode,
  clickySourceAvailable,
  evalPort,
}: PlaygroundConfigOptions): UserConfig {
  // Vitest also runs in `serve`, and it is the one mode that must not alias:
  // tests assert against the surface the package publishes, not against
  // whatever a sibling checkout happens to have mid-edit.
  const useClickySource = command === "serve" && clickySourceAvailable && mode !== "test";

  return {
    plugins: [react(), tailwindcss(), evalServer({ repoRoot, host: EVAL_HOST, port: evalPort })],
    resolve: {
      dedupe: ["react", "react-dom"],
      alias: command === "serve" && useClickySource ? clickyAliases : [],
    },
    server: {
      port: DEV_PORT,
      strictPort: true,
      proxy: {
        "/api": {
          target: `http://${EVAL_HOST}:${evalPort}`,
          changeOrigin: false,
        },
      },
      // Vite refuses to serve files outside the project root unless told to.
      fs: { allow: [root, resolve(root, "../.."), ...(useClickySource ? [clickySrc] : [])] },
    },
    optimizeDeps: {
      // The sibling lockfile is outside Vite's cache inputs, so dependency
      // paths can otherwise remain stale after clicky-ui upgrades a package.
      force: useClickySource,
      exclude: [
        "@flanksource/gomplate-lang",
        ...(useClickySource ? ["@flanksource/clicky-ui"] : []),
      ],
    },
  };
}

export default defineConfig(async ({ command, mode }) => {
  const managesEvalServer = process.env.GOMPLATE_PLAYGROUND_SERVER !== "0";
  const evalPort =
    command === "serve" && mode !== "test" && managesEvalServer
      ? await findAvailablePort({ host: EVAL_HOST, preferredPort: EVAL_PORT })
      : EVAL_PORT;

  return createPlaygroundConfig({ command, mode, clickySourceAvailable, evalPort });
});
