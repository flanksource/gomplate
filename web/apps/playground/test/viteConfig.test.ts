import { describe, expect, it } from "vitest";

import { createPlaygroundConfig } from "../vite.config";

const SELECTED_EVAL_PORT = 49_153;

describe("playground Vite configuration", () => {
  it.each([
    { command: "serve" as const, mode: "development", clickySourceAvailable: true, expected: true },
    { command: "serve" as const, mode: "test", clickySourceAvailable: true, expected: false },
    { command: "serve" as const, mode: "development", clickySourceAvailable: false, expected: false },
    { command: "build" as const, mode: "production", clickySourceAvailable: true, expected: false },
  ])(
    "sets dependency re-optimization to $expected for command=$command mode=$mode source=$clickySourceAvailable",
    ({ command, mode, clickySourceAvailable, expected }) => {
      const config = createPlaygroundConfig({
        command,
        mode,
        clickySourceAvailable,
        evalPort: SELECTED_EVAL_PORT,
      });

      expect(config.optimizeDeps?.force).toBe(expected);
    },
  );

  it("proxies API requests to the selected eval-server port", () => {
    const config = createPlaygroundConfig({
      command: "serve",
      mode: "development",
      clickySourceAvailable: false,
      evalPort: SELECTED_EVAL_PORT,
    });

    expect(config.server?.proxy?.["/api"]).toMatchObject({
      target: `http://127.0.0.1:${SELECTED_EVAL_PORT}`,
    });
  });

});
