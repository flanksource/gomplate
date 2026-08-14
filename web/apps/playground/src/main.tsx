import { createRoot } from "react-dom/client";
import { setFallbackIconProvider } from "@flanksource/clicky-ui";
import { clickyIconProvider } from "@flanksource/clicky-ui/icons";
import { App } from "./App";
import "@flanksource/clicky-ui/styles.css";
import "./styles.css";

// Icons referenced by name in schema-driven surfaces are runtime strings, which
// no import can resolve; registering the provider turns them into glyphs.
setFallbackIconProvider(clickyIconProvider());

const root = document.getElementById("app");
if (!root) throw new Error("#app root not found");
createRoot(root).render(<App />);
