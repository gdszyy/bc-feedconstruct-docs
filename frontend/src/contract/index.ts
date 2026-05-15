// Barrel for the BFF↔frontend contract.
// All cross-module imports must come through this entry point so future
// scope-specific event/rest files can be added without each track
// updating its own import paths.

export * from "./events";
export * from "./rest";
