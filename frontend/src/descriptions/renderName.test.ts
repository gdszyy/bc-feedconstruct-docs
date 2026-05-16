import { describe, expect, it } from "vitest";

import { renderName } from "./renderName";

// ---------------------------------------------------------------------------
// M12 — renderName
//
// Locked decisions (PR thread):
//   - Templates may contain {key} placeholders (e.g. {player}, {handicap},
//     {minute}). The renderer substitutes from a vars map.
//   - Missing keys are kept as the literal '{key}'. This is the deliberate
//     fallback — never silently drop content, never expose vendor ids, and
//     stay debuggable when a description package hasn't caught up to a
//     market's new specifier shape.
//   - Numeric values are coerced to their decimal string form.
//   - Pure function: no listeners, no caching.
// ---------------------------------------------------------------------------

// =================== No placeholders ===================

// Given a template with no {key} placeholders
// When renderName(template, anyVars) is invoked
// Then the template is returned unchanged
describe("M12 renderName: returns plain template untouched", () => {
  it("when the template has no placeholders then renderName returns the template as-is", () => {
    expect(renderName("Full Time Result", { player: "ignored" })).toBe(
      "Full Time Result",
    );
  });
});

// =================== Single substitution ===================

// Given a template containing one {key} placeholder
// When renderName is invoked with a vars object that supplies that key
// Then the {key} is replaced with the supplied value
describe("M12 renderName: substitutes a single placeholder", () => {
  it("when one placeholder maps to a string in vars then the output replaces it inline", () => {
    expect(renderName("Goalscorer: {player}", { player: "Messi" })).toBe(
      "Goalscorer: Messi",
    );
  });
});

// =================== Multiple substitution ===================

// Given a template with several {key} placeholders
// When renderName is invoked with vars covering each key
// Then all are substituted in one pass
describe("M12 renderName: substitutes multiple placeholders", () => {
  it("when several placeholders are present then renderName substitutes them all", () => {
    expect(
      renderName("{player} scores in minute {minute}", {
        player: "Messi",
        minute: 23,
      }),
    ).toBe("Messi scores in minute 23");
  });
});

// =================== Numeric values ===================

// Given a template with a numeric-typed slot (e.g. {handicap})
// When renderName is invoked with a number value
// Then the number is coerced to its string form (e.g. -1.5 → "-1.5")
describe("M12 renderName: coerces numeric values to strings", () => {
  it("when vars contains a number then renderName stringifies it in the output", () => {
    expect(renderName("Handicap {handicap}", { handicap: -1.5 })).toBe(
      "Handicap -1.5",
    );
    expect(renderName("Over {total}", { total: 2 })).toBe("Over 2");
  });
});

// =================== Missing keys ===================

// Given a template with a {key} placeholder
// When renderName is invoked with vars that does NOT contain the key
// Then the literal '{key}' is preserved in the output (debuggable fallback)
describe("M12 renderName: preserves missing placeholders verbatim", () => {
  it("when a key is missing in vars then renderName keeps the literal {key} in the output", () => {
    expect(renderName("Goalscorer: {player}", {})).toBe("Goalscorer: {player}");
    expect(
      renderName("{player} in minute {minute}", { player: "Messi" }),
    ).toBe("Messi in minute {minute}");
  });
});

// =================== Repeated placeholders ===================

// Given a template with the same {key} occurring more than once
// When renderName is invoked with that key supplied once in vars
// Then every occurrence is replaced
describe("M12 renderName: replaces repeated placeholders globally", () => {
  it("when a key appears multiple times then every occurrence is substituted", () => {
    expect(
      renderName("{player} vs {player} — anytime scorer", { player: "Messi" }),
    ).toBe("Messi vs Messi — anytime scorer");
  });
});

// =================== Empty / no-op inputs ===================

// Given an empty template or empty vars map
// When renderName is invoked
// Then the output is the template unchanged (empty stays empty; no vars yields literal placeholders)
describe("M12 renderName: edge cases for empty inputs", () => {
  it("when the template is empty or vars is empty then renderName returns the template literally", () => {
    expect(renderName("", { player: "Messi" })).toBe("");
    expect(renderName("Just text", {})).toBe("Just text");
    expect(renderName("{a}{b}", {})).toBe("{a}{b}");
  });
});
