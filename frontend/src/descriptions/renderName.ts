// ---------------------------------------------------------------------------
// M12 — renderName
//
// Pure template renderer for market / outcome / group / tab descriptions.
// Substitutes {key} placeholders from a vars map; missing keys are left as
// the literal '{key}' so a description package that hasn't caught up with a
// market's specifier shape stays debuggable instead of silently dropping
// content or exposing vendor ids (M12 principle: 禁止把供应商 ID 直接展示给用户).
//
// Stateless: the description cache lives in DescriptionsStore.
// ---------------------------------------------------------------------------

const PLACEHOLDER = /\{([^{}]+)\}/g;

export type RenderVars = Record<string, string | number>;

export function renderName(template: string, vars: RenderVars = {}): string {
  return template.replace(PLACEHOLDER, (literal, key: string) => {
    if (!Object.prototype.hasOwnProperty.call(vars, key)) return literal;
    const value = vars[key];
    return typeof value === "number" ? String(value) : value;
  });
}
