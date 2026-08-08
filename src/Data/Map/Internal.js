export const empty = {};
export const isEmpty = function(a) { return a; };
export const singleton = function(a, b) { return [a, b]; };
export const insertImpl = function(a, b, c, d, e) { return [a, b, c, d, e]; };
export const insertWithImpl = function(a, b, c, d, e, f) { return [a, b, c, d, e, f]; };
export const lookupImpl = function(a, b, c, d, e, f) { return [a, b, c, d, e, f]; };
export const deleteImpl = function(a, b, c, d) { return [a, b, c, d]; };
export const keysImpl = function(a) { return a; };
export const valuesImpl = function(a) { return a; };
export const unionWithImpl = function(a, b, c, d, e) { return [a, b, c, d, e]; };
export const intersectionWithImpl = function(a, b, c, d, e) { return [a, b, c, d, e]; };
export const differenceImpl = function(a, b, c, d) { return [a, b, c, d]; };
export const sizeImpl = function(a) { return a; };

export const filterKeysImpl = function(a, b) { return [a, b]; };
export const mapImpl = function(a, b) { return [a, b]; };
export const foldlImpl = function(a, b, c) { return [a, b, c]; };
export const foldrImpl = function(a, b, c) { return [a, b, c]; };
