import { createServer } from "node:net";
import { describe, expect, it } from "vitest";

import { findAvailablePort } from "../plugins/eval-server";
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

  it("selects another loopback port when the preferred port is occupied", async () => {
    const occupied = createServer();
    await new Promise<void>((resolve, reject) => {
      occupied.once("error", reject);
      occupied.listen(0, "127.0.0.1", resolve);
    });

    try {
      const address = occupied.address();
      if (!address || typeof address === "string") throw new Error("test listener has no TCP port");
      const selectedPort = await findAvailablePort({
        host: "127.0.0.1",
        preferredPort: address.port,
      });

      expect(selectedPort).not.toBe(address.port);
    } finally {
      await new Promise<void>((resolve, reject) => {
        occupied.close((error) => (error ? reject(error) : resolve()));
      });
    }
  });
});
