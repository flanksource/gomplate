export interface Example {
  name: string;
  source: string;
  input: string;
}

const POD_INPUT = `pod:
  metadata:
    name: web-7d4f
    labels:
      app: web
      tier: frontend
  spec:
    containers:
      - name: web
        resources:
          limits:
            cpu: 500m
            memory: 1Gi
  status:
    phase: Running
replicas: 3
env: production
`;

/**
 * Starting points per language. Each one is chosen to exercise something the
 * highlighter has to get right as well as something the evaluator does.
 */
export const EXAMPLES: Record<string, Example[]> = {
  cel: [
    {
      name: "Field access and optionals",
      source: `pod.metadata.labels.app + "/" + pod.status.?phase.orValue("unknown")`,
      input: POD_INPUT,
    },
    {
      name: "Kubernetes helpers",
      source: `{
  "cpu": k8s.cpuAsMillicores(pod.spec.containers[0].resources.limits.cpu),
  "memory": k8s.memoryAsBytes(pod.spec.containers[0].resources.limits.memory),
}`,
      input: POD_INPUT,
    },
    {
      name: "Macros and collections",
      source: `[1, 2, 3, 4].filter(e, e % 2 == 0).map(e, e * 10)`,
      input: "",
    },
    {
      name: "fold",
      source: `["a", "b", "c"].fold(e, acc, acc + e)`,
      input: "",
    },
    {
      name: "String literal forms",
      source: `[
  "plain",
  """triple "quoted" """,
  r"raw\\dstring",
  "unicode \\U0001F600",
]`,
      input: "",
    },
    {
      name: "Strings and time",
      source: `"hello world".upperAscii() + " @ " + string(time.Now().getFullYear())`,
      input: "",
    },
  ],
  // Keyed by the playground language id, not the evaluator name: the go
  // template language is `gomplate` here and `gotemplate` on the wire.
  gomplate: [
    {
      name: "Pipelines",
      source: `{{ .pod.metadata.name | strings.ToUpper }}`,
      input: POD_INPUT,
    },
    {
      name: "Control flow and variables",
      source: `{{- $name := .pod.metadata.name -}}
{{ if eq .env "production" }}PROD: {{ $name }}{{ else }}dev: {{ $name }}{{ end }}`,
      input: POD_INPUT,
    },
    {
      name: "Collections",
      source: `{{ coll.Dict "app" .pod.metadata.labels.app "replicas" .replicas | toJSON }}`,
      input: POD_INPUT,
    },
    {
      name: "Comments and trim markers",
      source: `{{/* not rendered */}}
{{- range $i, $v := coll.Slice "a" "b" "c" }}
{{ $i }}={{ $v }}
{{- end }}`,
      input: "",
    },
  ],
  "yaml-gomplate": [
    {
      name: "Templated manifest",
      source: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: "{{ .pod.metadata.labels.app }}-deployment"
  labels:
    app: {{ .pod.metadata.labels.app }}
spec:
  replicas: {{ .replicas }}
  template:
    spec:
      containers:
        - name: {{ .pod.metadata.labels.app }}
          image: "registry.example.com/{{ .pod.metadata.labels.app }}:latest"
`,
      input: POD_INPUT,
    },
  ],
  "json-gomplate": [
    {
      name: "Templated JSON",
      source: `{
  "app": "{{ .pod.metadata.labels.app }}",
  "replicas": {{ .replicas }},
  "env": "{{ .env }}"
}`,
      input: POD_INPUT,
    },
  ],
  jsonpath: [
    { name: "Field path", source: `$.pod.metadata.name`, input: POD_INPUT },
    { name: "Recursive descent", source: `$..name`, input: POD_INPUT },
    { name: "Filter", source: `$.pod.spec.containers[?(@.name == "web")]`, input: POD_INPUT },
  ],
  javascript: [
    { name: "Arithmetic", source: `replicas * 2`, input: POD_INPUT },
    { name: "Object access", source: `pod.metadata.labels.app`, input: POD_INPUT },
  ],
};

export function examplesFor(languageId: string): Example[] {
  return EXAMPLES[languageId] ?? [];
}

export function defaultExample(languageId: string): Example {
  const [first] = examplesFor(languageId);
  return first ?? { name: "Empty", source: "", input: "" };
}
