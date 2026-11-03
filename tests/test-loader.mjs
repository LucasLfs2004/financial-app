import fs from "node:fs";
import path from "node:path";
import { createRequire } from "node:module";
import { fileURLToPath } from "node:url";
import ts from "typescript";
const dependencyRequire = createRequire(import.meta.url);
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");

// Compile the actual TS modules in memory; no additional test runtime dependency.
export function load(relative, stubs = {}, cache = new Map()) {
  const filename = path.resolve(root, relative);
  if (cache.has(filename)) return cache.get(filename).exports;
  const compiledModule = { exports: {} };
  cache.set(filename, compiledModule);
  const code = ts.transpileModule(fs.readFileSync(filename, "utf8"), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX },
  }).outputText;
  const localRequire = (name) => {
    if (Object.hasOwn(stubs, name)) return stubs[name];
    const target = name.startsWith("@/") ? name.slice(2) : name.startsWith(".") ? path.relative(root, path.resolve(path.dirname(filename), name)) : null;
    return target ? load(fs.existsSync(path.resolve(root, `${target}.ts`)) ? `${target}.ts` : `${target}.tsx`, stubs, cache) : dependencyRequire(name);
  };
  new Function("require", "module", "exports", code)(localRequire, compiledModule, compiledModule.exports);
  return compiledModule.exports;
}
