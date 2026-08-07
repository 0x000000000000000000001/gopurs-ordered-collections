const fs = require('fs');
const wasmBuffer = fs.readFileSync('../gopurs/bin/ffi_gen.wasm');
const crypto = require('crypto');
global.crypto = crypto; // required for wasm exec

require('../gopurs/bin/wasm_exec.js'); // Assuming standard Go wasm_exec
const go = new Go();
WebAssembly.instantiate(wasmBuffer, go.importObject).then((result) => {
    go.run(result.instance);
    const code = `package Internal
func Size(m interface{}) int { return 1 }
`;
    const res = global.parseFFI(code);
    console.log(res);
});
