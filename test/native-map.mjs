// Uses the existing compiled gopurs FFI generator; no PureScript rebuild required.
// node test/native-map.mjs [--race] [--bench | --bench-union] [--curried-baseline]
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { prepareFfi } from "../../gopurs/output/Gopurs.FfiSupport/index.js";
import { generateFfiBridge } from "../../gopurs/output/Gopurs.FfiBridge/index.js";
import { Nothing } from "../../gopurs/output/Data.Maybe/index.js";
import { Tuple } from "../../gopurs/output/Data.Tuple/index.js";

const baseline = process.argv.includes("--curried-baseline");
const benchUnion = process.argv.includes("--bench-union");
const directory = mkdtempSync(join(tmpdir(), "gopurs-map-ffi-"));
try {
  let source = readFileSync(new URL("../src/Data/Map/Internal.go", import.meta.url), "utf8");
  if (baseline) {
    // Isolate only the previous callback ABI, keeping today's identical B-tree implementation.
    const signature = "compare func(interface{}, interface{}) interface{}";
    assert.equal(source.split(signature).length - 1, 7);
    source = source.replaceAll(signature, "compare func(interface{}) func(interface{}) interface{}")
      .replaceAll("compare(a, b)", "compare(a)(b)");
  } else {
    mkdirSync(join(directory, "native"));
    writeFileSync(join(directory, "native/Map.go"), source.replace("package Data_Map_Internal", 'package Data_Map_Internal\nimport gopurs_runtime "gopurs/output/runtime"'));
    copyFileSync(new URL("./native-map_test.go", import.meta.url), join(directory, "native/map_test.go"));
  }
  mkdirSync(join(directory, "bridge"));
  mkdirSync(join(directory, "runtime"));
  copyFileSync(new URL("../../gopurs/runtime/runtime.go", import.meta.url), join(directory, "runtime/runtime.go"));
  writeFileSync(join(directory, "go.mod"), "module gopurs/output\n\ngo 1.22\n");
  const prepared = prepareFfi({ moduleName: "Data.Map.Internal", path: fileURLToPath(new URL("../src/Data/Map/Internal.go", import.meta.url)) })("Map_")(source)();
  const names = ["insertImpl", "insertWithImpl", "lookupImpl", "deleteImpl", "unionWithImpl", "intersectionWithImpl", "differenceImpl"];
  const wrappers = generateFfiBridge("Map")([])(prepared.decls)(names.map(name => new Tuple(name, Nothing.value)));
  writeFileSync(join(directory, "bridge/Map.go"), 'package mapbridge\nimport gopurs_runtime "gopurs/output/runtime"\n' + prepared.content + "\n" + wrappers);
  copyFileSync(new URL("./native-map-bridge_test.go", import.meta.url), join(directory, "bridge/map_test.go"));
  const args = ["test", "-count=1", "-v"];
  if ((!process.argv.includes("--bench") && !benchUnion) || process.argv.includes("--race")) args.push("-race");
  if (benchUnion) args.push("-bench", "BenchmarkMapUnion", "-benchtime=30x");
  if (process.argv.includes("--bench")) args.push("-bench", "BenchmarkMapLookup", "-benchtime=10000x");
  args.push(baseline ? "./bridge" : "./...");
  console.log(baseline ? "Previous curried callback ABI, same Map source" : "Native Map and generated binary callback bridge");
  execFileSync("go", args, { cwd: directory, stdio: "inherit", timeout: 60_000, env: { ...process.env, GOWORK: "off" } });
} finally {
  rmSync(directory, { recursive: true, force: true });
}
