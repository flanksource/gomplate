import type { LanguageId } from "@flanksource/gomplate-lang";
import {
  UiBraces,
  UiCode2,
  UiFileCode,
  UiFileJson,
  UiFileText,
  UiFunction,
} from "@flanksource/clicky-ui/icons";
import type { StaticIconComponent } from "@flanksource/clicky-ui";

/** An evaluator on the Go side. */
export type EvalLanguage = "cel" | "gotemplate" | "jsonpath" | "javascript";

/** One entry in the playground's language rail. */
export interface PlaygroundLanguage {
  id: string;
  label: string;
  icon: StaticIconComponent;
  /** Rail section this language belongs to. */
  section: "Expressions" | "Templates";
  /** Monaco language for the source editor. */
  editorLanguage: LanguageId | "javascript";
  /** Which evaluator the Go server should run. */
  evalLanguage: EvalLanguage;
  description: string;
  /**
   * Rows the expression editor opens at, before the reader drags the divider.
   *
   * An expression is a line or two, so five rows leaves the input document the
   * rest of the pane. A template is a whole document, and starting it at five
   * rows would hide most of what is being written.
   */
  editorRows: number;
}

/** Default rows for a language whose source is an expression. */
const EXPRESSION_ROWS = 5;

/** Default rows for a language whose source is a document. */
const TEMPLATE_ROWS = 14;

export const LANGUAGES: PlaygroundLanguage[] = [
  {
    id: "cel",
    label: "CEL",
    icon: UiFunction,
    section: "Expressions",
    editorLanguage: "cel",
    evalLanguage: "cel",
    description: "Common Expression Language, with gomplate's k8s, aws and math helpers",
    editorRows: EXPRESSION_ROWS,
  },
  {
    id: "jsonpath",
    label: "JSONPath",
    icon: UiBraces,
    section: "Expressions",
    editorLanguage: "jsonpath",
    evalLanguage: "jsonpath",
    description: "JSONPath, as evaluated by ojg",
    editorRows: EXPRESSION_ROWS,
  },
  {
    id: "javascript",
    label: "JavaScript",
    icon: UiCode2,
    section: "Expressions",
    editorLanguage: "javascript",
    evalLanguage: "javascript",
    description: "JavaScript, as evaluated by otto",
    editorRows: EXPRESSION_ROWS,
  },
  {
    id: "gomplate",
    label: "Go template",
    icon: UiFileCode,
    section: "Templates",
    editorLanguage: "gomplate",
    evalLanguage: "gotemplate",
    description: "Go text/template with gomplate's function library",
    editorRows: TEMPLATE_ROWS,
  },
  {
    id: "yaml-gomplate",
    label: "YAML + template",
    icon: UiFileText,
    section: "Templates",
    editorLanguage: "yaml-gomplate",
    evalLanguage: "gotemplate",
    description: "YAML with templates embedded in it, as real configuration is written",
    editorRows: TEMPLATE_ROWS,
  },
  {
    id: "json-gomplate",
    label: "JSON + template",
    icon: UiFileJson,
    section: "Templates",
    editorLanguage: "json-gomplate",
    evalLanguage: "gotemplate",
    description: "JSON with templates embedded in it",
    editorRows: TEMPLATE_ROWS,
  },
];

/** Rail section order. */
export const SECTIONS = ["Expressions", "Templates"] as const;

export function languageById(id: string): PlaygroundLanguage {
  const found = LANGUAGES.find((language) => language.id === id);
  if (!found) throw new Error(`unknown playground language "${id}"`);
  return found;
}
