import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { runtimeGoCode } from '../../gopurs/output/Gopurs.Runtime/index.js';

// Run from this checkout after building the adjacent gopurs Node backend.
// An optional source path supports checking the same regression before a fix.
const source = process.argv[2] ?? fileURLToPath(new URL('../src/Data/Map/Internal.go', import.meta.url));
const workspace = mkdtempSync(join(tmpdir(), 'gopurs-map-race-'));
try {
    mkdirSync(join(workspace, 'gopurs_runtime'));
    writeFileSync(join(workspace, 'go.mod'), 'module gopurs/output\n\ngo 1.22\n');
    writeFileSync(join(workspace, 'gopurs_runtime/runtime.go'), runtimeGoCode);
    // The FFI emitter normally supplies this runtime import.
    writeFileSync(join(workspace, 'map.go'), readFileSync(source, 'utf8').replace(
        'package Data_Map_Internal', 'package Data_Map_Internal\nimport "gopurs/output/gopurs_runtime"'));
    copyFileSync(new URL('./native-map_test.go', import.meta.url), join(workspace, 'map_test.go'));
    const result = spawnSync('go', ['test', '-race', '-count=1', '.'], {
        cwd: workspace, stdio: 'inherit', timeout: 60_000,
        env: { ...process.env, GOWORK: 'off' },
    });
    assert.ifError(result.error);
    assert.equal(result.status, 0, 'native Map concurrency regression');
} finally {
    rmSync(workspace, { recursive: true, force: true });
}
