import { spec } from "./generated";
import { pathFlavour } from "./environment";
import type { PathSegment } from "./environment";

/** The subset of a Monaco model this module reads. */
export interface PrefixModel {
  getLineContent(line: number): string;
}

export interface PrefixPosition {
  lineNumber: number;
  column: number;
}

/** A path expression under construction at the cursor. */
export interface EnvironmentPrefix {
  /** The segments already committed, left of the trailing dot. */
  segments: PathSegment[];
  /** The partially typed leaf, empty right after a dot. */
  leaf: string;
  /**
   * 1-based columns spanning the whole expression typed so far, root marker
   * included. A completion replaces this range with a freshly rendered path, so
   * the inserted text is always well formed rather than glued onto what is
   * already there.
   */
  startColumn: number;
  endColumn: number;
  /** The text in that range, which Monaco filters candidates against. */
  typed: string;
}

const IDENT = /[A-Za-z0-9_]/;
const SUBSCRIPT = /^(?:(\d+)|"((?:[^"\\]|\\.)*)"|'((?:[^'\\]|\\.)*)')$/;

/**
 * Reads the path expression the cursor sits in, or null when it sits somewhere
 * the document's keys have no meaning.
 *
 * `dottedWordAt` in `hover.ts` cannot serve here: it scans past the cursor and
 * strips the trailing dot, and the trailing dot is precisely the signal that a
 * child is being completed.
 */
export function environmentPrefixAt(
  model: PrefixModel,
  position: PrefixPosition,
  languageId: string,
): EnvironmentPrefix | null {
  const flavour = pathFlavour(languageId);
  if (!flavour) return null;

  const line = model.getLineContent(position.lineNumber);
  const before = line.slice(0, position.column - 1);

  if (flavour === "gotemplate" && !insideAction(before)) return null;

  let leafStart = before.length;
  while (leafStart > 0 && IDENT.test(before[leafStart - 1]!)) leafStart--;
  const leaf = before.slice(leafStart);
  const head = before.slice(0, leafStart);

  const chain = parseChain(head);
  if (!chain) return null;
  // Without a dot the leaf can only be a root name, so anything that looks like
  // the tail of a path before it (`items[0]name`) is malformed, not a prefix.
  if (leaf !== "" && !chain.dotted && chain.segments.length > 0) return null;

  // `chain.start` already sits on the leading `.` when there is one, so a go
  // template needs no adjustment; JSONPath's `$` sits one place further left.
  let start = chain.start;
  if (flavour === "gotemplate" && !chain.rooted) return null;
  if (flavour === "jsonpath") {
    const marker = head[chain.start - 1];
    if (marker !== "$" && marker !== "@") return null;
    start -= 1;
  }
  if (flavour === "cel" && chain.rooted) return null; // a leading `.` is not CEL

  return {
    segments: chain.segments,
    leaf,
    startColumn: start + 1,
    endColumn: position.column,
    typed: before.slice(start),
  };
}

interface Chain {
  segments: PathSegment[];
  /** Index in `head` where the segment list starts, root marker excluded. */
  start: number;
  /** Whether `head` ends with the dot that opens a new segment. */
  dotted: boolean;
  /** Whether the outermost segment is preceded by a `.`, as go templates need. */
  rooted: boolean;
}

/**
 * Parses the trailing path of `head` backwards.
 *
 * Backwards because a path has no left boundary of its own: it ends where the
 * expression around it begins, and that is only knowable by walking off it.
 */
function parseChain(head: string): Chain | null {
  let pos = head.length;
  const dotted = pos > 0 && head[pos - 1] === ".";
  if (dotted) pos--;

  const reversed: PathSegment[] = [];
  let rooted = false;

  for (;;) {
    if (pos > 0 && head[pos - 1] === "]") {
      const open = head.lastIndexOf("[", pos - 2);
      if (open < 0) return null;
      const segment = parseSubscript(head.slice(open + 1, pos - 1));
      if (segment === null) return null;
      reversed.push(segment);
      pos = open;
      rooted = false;
      continue;
    }
    if (pos > 0 && IDENT.test(head[pos - 1]!)) {
      let start = pos;
      while (start > 0 && IDENT.test(head[start - 1]!)) start--;
      reversed.push(head.slice(start, pos));
      pos = start;
      rooted = pos > 0 && head[pos - 1] === ".";
      if (rooted) {
        pos--;
        continue;
      }
      break;
    }
    break;
  }

  // Only the trailing dot was consumed, so that dot is itself the root marker.
  if (reversed.length === 0 && dotted) rooted = true;

  return { segments: reversed.reverse(), start: pos, dotted, rooted };
}

function parseSubscript(inner: string): PathSegment | null {
  const match = SUBSCRIPT.exec(inner.trim());
  if (!match) return null;
  if (match[1] !== undefined) return Number(match[1]);
  const quoted = match[2] ?? match[3]!;
  try {
    return JSON.parse(`"${quoted.replace(/\\'/g, "'")}"`) as string;
  } catch {
    return null;
  }
}

/**
 * Whether the cursor sits between template delimiters. Outside them the text is
 * literal output, where a key path means nothing.
 */
function insideAction(before: string): boolean {
  const { left, right, leftComment, rightComment } = spec.gotemplate.delimiters;

  const open = before.lastIndexOf(left);
  if (open < 0) return false;
  if (before.lastIndexOf(right) > open) return false;

  // `{{/*` opens with the ordinary left delimiter, so an unterminated comment
  // reads as an open action unless it is checked for separately.
  const comment = before.lastIndexOf(leftComment);
  if (comment >= open && before.lastIndexOf(rightComment) < comment) return false;

  return true;
}
