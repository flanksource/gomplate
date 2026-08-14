/**
 * Introspection of the document an expression is evaluated against.
 *
 * Deliberately free of Monaco and of any UI dependency: the same functions back
 * editor completion and any caller that needs to render a document's shape.
 */

/** One step of a path. A number indexes a list, a string keys a map. */
export type PathSegment = string | number;

/** The JSON shape of a value, as a name a reader recognises. */
export type ValueKind = "string" | "number" | "boolean" | "null" | "object" | "array";

/** One child of a container, as completion and shape views need it. */
export interface EnvironmentEntry {
  /** The key or index, as typed. */
  key: string;
  segment: PathSegment;
  kind: ValueKind;
  /** A short rendering of the value, for a detail column. */
  summary: string;
  /** Whether the value has children of its own. */
  container: boolean;
}

const IDENTIFIER = /^[A-Za-z_][A-Za-z0-9_]*$/;

/** Whether `key` can be written as a bare `.key` rather than a subscript. */
export function isIdentifier(key: string): boolean {
  return IDENTIFIER.test(key);
}

export function kindOf(value: unknown): ValueKind {
  if (value === null || value === undefined) return "null";
  if (Array.isArray(value)) return "array";
  switch (typeof value) {
    case "string":
      return "string";
    case "number":
      return "number";
    case "boolean":
      return "boolean";
    default:
      return "object";
  }
}

const SUMMARY_LIMIT = 32;

/** A one-line rendering of a value: the sample a completion detail shows. */
export function summarize(value: unknown): string {
  switch (kindOf(value)) {
    case "null":
      return "null";
    case "array":
      return plural((value as unknown[]).length, "item");
    case "object":
      return plural(Object.keys(value as Record<string, unknown>).length, "key");
    case "string": {
      const text = value as string;
      return text.length > SUMMARY_LIMIT ? `${JSON.stringify(text.slice(0, SUMMARY_LIMIT))}…` : JSON.stringify(text);
    }
    default:
      return String(value);
  }
}

function plural(count: number, noun: string): string {
  return `${count} ${noun}${count === 1 ? "" : "s"}`;
}

/**
 * Walks `segments` from the root.
 *
 * A numeric segment indexes a list and nothing else; a string segment keys a
 * map. Returns undefined as soon as a step does not apply, so a half-typed path
 * simply yields no completions rather than throwing.
 */
export function resolvePath(environment: unknown, segments: readonly PathSegment[]): unknown {
  let current = environment;
  for (const segment of segments) {
    if (typeof segment === "number") {
      if (!Array.isArray(current)) return undefined;
      current = current[segment];
      continue;
    }
    if (!isRecord(current)) return undefined;
    current = current[segment];
  }
  return current;
}

/**
 * The children of a container, in document order.
 *
 * A list yields its indices rather than the keys of its first element: no
 * language here lets a field follow a list without an index, so offering
 * `containers.name` would complete an expression that cannot evaluate.
 */
export function childEntries(value: unknown): EnvironmentEntry[] {
  if (Array.isArray(value)) {
    return value.map((element, index) => ({
      key: String(index),
      segment: index,
      kind: kindOf(element),
      summary: summarize(element),
      container: isContainer(element),
    }));
  }
  if (!isRecord(value)) return [];
  return Object.entries(value).map(([key, child]) => ({
    key,
    segment: key,
    kind: kindOf(child),
    summary: summarize(child),
    container: isContainer(child),
  }));
}

function isContainer(value: unknown): boolean {
  const kind = kindOf(value);
  return kind === "object" || kind === "array";
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

/**
 * Renders a path in one language's syntax.
 *
 * The single place path syntax is spelled out, so completion and click-to-insert
 * cannot drift apart. Returns null when the path has no expression in that
 * language — a go-template reaches a list element or an awkward key through
 * `index`, which is a call rather than a path.
 */
export function pathExpression(languageId: string, segments: readonly PathSegment[]): string | null {
  const flavour = pathFlavour(languageId);
  if (!flavour) return null;
  if (flavour === "gotemplate") {
    if (segments.some((segment) => typeof segment === "number" || !isIdentifier(segment))) {
      return null;
    }
    return segments.length === 0 ? "." : segments.map((segment) => `.${String(segment)}`).join("");
  }

  const root = flavour === "jsonpath" ? "$" : "";
  let out = root;
  for (const segment of segments) {
    if (typeof segment === "number") {
      out += `[${segment}]`;
    } else if (isIdentifier(segment)) {
      out += out === "" ? segment : `.${segment}`;
    } else {
      out += `[${JSON.stringify(segment)}]`;
    }
  }
  return out;
}

/** Which path syntax a language id uses. */
export function pathFlavour(languageId: string): "cel" | "gotemplate" | "jsonpath" | null {
  if (languageId === "cel") return "cel";
  if (languageId === "jsonpath") return "jsonpath";
  if (languageId === "gomplate" || languageId.endsWith("-gomplate")) return "gotemplate";
  return null;
}
