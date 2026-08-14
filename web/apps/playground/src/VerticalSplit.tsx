import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
  type ReactNode,
} from "react";

/** Monaco's line height, as clicky-ui's MonacoEditor configures it. */
export const EDITOR_LINE_HEIGHT = 20;

/** Height of an EditorPane header row (text-xs on py-2). */
const PANE_HEADER_HEIGHT = 33;

/** Slack for Monaco's own chrome: the horizontal scrollbar and top padding. */
const EDITOR_CHROME = 12;

/**
 * Pane height that shows `rows` lines of code without scrolling.
 *
 * Expressed in rows because that is how the size is actually reasoned about --
 * a CEL expression is one or two lines, a templated manifest is a screenful --
 * and it keeps the default honest if the editor's line height ever changes.
 */
export function rowsToPaneHeight(rows: number): number {
  return PANE_HEADER_HEIGHT + rows * EDITOR_LINE_HEIGHT + EDITOR_CHROME;
}

export interface VerticalSplitProps {
  top: ReactNode;
  bottom: ReactNode;
  /** Top-pane height in pixels, used until the reader drags the divider. */
  defaultTopHeight: number;
  /** Smallest the top pane may be dragged. Defaults to two rows. */
  minTop?: number;
  /** Smallest the bottom pane may be dragged. */
  minBottom?: number;
  /** localStorage key persisting the dragged height. */
  storageKey?: string;
}

/**
 * A vertically stacked, resizable pair of panes.
 *
 * clicky-ui's SplitPane only splits horizontally, and the two editors stack, so
 * the divider here is its own component. Sizing is in pixels rather than a
 * percentage because the useful size of the expression editor is a number of
 * rows, which a percentage of a varying viewport does not express.
 */
export function VerticalSplit({
  top,
  bottom,
  defaultTopHeight,
  minTop = rowsToPaneHeight(2),
  minBottom = 96,
  storageKey,
}: VerticalSplitProps) {
  const container = useRef<HTMLDivElement>(null);
  const [topHeight, setTopHeight] = useState(() => readStored(storageKey) ?? defaultTopHeight);
  const [dragging, setDragging] = useState(false);

  // Only follow the language's default while the reader has not chosen a size:
  // a dragged divider is a decision, and switching language should not undo it.
  const hasStoredSize = useRef(readStored(storageKey) !== null);
  useEffect(() => {
    if (!hasStoredSize.current) setTopHeight(defaultTopHeight);
  }, [defaultTopHeight]);

  const clamp = useCallback(
    (height: number) => {
      const available = container.current?.getBoundingClientRect().height ?? 0;
      const upperBound = Math.max(minTop, available - minBottom);
      return Math.round(Math.max(minTop, Math.min(upperBound, height)));
    },
    [minTop, minBottom],
  );

  const commit = useCallback(
    (height: number) => {
      const next = clamp(height);
      setTopHeight(next);
      hasStoredSize.current = true;
      if (storageKey) window.localStorage.setItem(storageKey, String(next));
    },
    [clamp, storageKey],
  );

  const onPointerDown = useCallback(
    (event: React.PointerEvent<HTMLDivElement>) => {
      event.preventDefault();
      const rect = container.current?.getBoundingClientRect();
      if (!rect) return;

      setDragging(true);
      const onMove = (moveEvent: PointerEvent) => commit(moveEvent.clientY - rect.top);
      const onUp = () => {
        setDragging(false);
        document.removeEventListener("pointermove", onMove);
        document.removeEventListener("pointerup", onUp);
      };
      document.addEventListener("pointermove", onMove);
      document.addEventListener("pointerup", onUp);
    },
    [commit],
  );

  // A separator that can only be dragged is unusable without a mouse, and this
  // one gates access to the input editor.
  const onKeyDown = useCallback(
    (event: ReactKeyboardEvent<HTMLDivElement>) => {
      const step = event.shiftKey ? EDITOR_LINE_HEIGHT * 5 : EDITOR_LINE_HEIGHT;
      if (event.key === "ArrowUp") {
        event.preventDefault();
        commit(topHeight - step);
      } else if (event.key === "ArrowDown") {
        event.preventDefault();
        commit(topHeight + step);
      } else if (event.key === "Home") {
        event.preventDefault();
        commit(minTop);
      }
    },
    [commit, topHeight, minTop],
  );

  return (
    <div ref={container} className="flex h-full min-h-0 flex-col">
      <div className="min-h-0 shrink-0 overflow-hidden" style={{ height: topHeight }}>
        {top}
      </div>

      <div
        role="separator"
        aria-orientation="horizontal"
        aria-label="Resize the expression editor"
        aria-valuenow={topHeight}
        tabIndex={0}
        onPointerDown={onPointerDown}
        onKeyDown={onKeyDown}
        className={`group relative h-1.5 shrink-0 cursor-row-resize bg-border transition-colors hover:bg-primary/40 focus-visible:bg-primary/60 focus-visible:outline-none ${
          dragging ? "bg-primary/60" : ""
        }`}
      >
        {/* A 6px strip is a small target; widen the grab area without moving
            anything by overflowing an invisible band above and below it. */}
        <span className="absolute inset-x-0 -top-1 -bottom-1" />
      </div>

      <div className="min-h-0 flex-1 overflow-hidden">{bottom}</div>
    </div>
  );
}

function readStored(key: string | undefined): number | null {
  if (!key || typeof window === "undefined") return null;
  const raw = window.localStorage.getItem(key);
  if (raw === null) return null;
  const parsed = Number.parseInt(raw, 10);
  return Number.isFinite(parsed) ? parsed : null;
}
