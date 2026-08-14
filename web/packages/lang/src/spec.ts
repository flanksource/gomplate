import { spec } from "./generated";
import type { GomplateSpec, SpecFunction, SpecMacro } from "./types";

export { spec };
export type { GomplateSpec, SpecFunction, SpecMacro };

/** Looks up a CEL function by its fully qualified name. */
export function celFunction(name: string): SpecFunction | undefined {
  return spec.cel.functions.find((fn) => fn.name === name);
}

/** Looks up a go-template function by its fully qualified name. */
export function goTemplateFunction(name: string): SpecFunction | undefined {
  return spec.gotemplate.functions.find((fn) => fn.name === name);
}

/** Every CEL function in a namespace, e.g. all of `k8s.*`. */
export function celNamespace(namespace: string): SpecFunction[] {
  return spec.cel.functions.filter((fn) => fn.namespace === namespace);
}

/** Every go-template function in a namespace. */
export function goTemplateNamespace(namespace: string): SpecFunction[] {
  return spec.gotemplate.functions.filter((fn) => fn.namespace === namespace);
}
