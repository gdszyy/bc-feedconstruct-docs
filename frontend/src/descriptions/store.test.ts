import { describe, expect, it, vi } from "vitest";

import type {
  MarketDescription,
  OutcomeDescription,
} from "@/contract/rest";

import { DescriptionsStore, renderName } from "./store";

// ---------------------------------------------------------------------------
// M12 — DescriptionsStore + renderName
//
// Locked decisions (per PR thread):
//   1. Multi-locale coexistence (acceptance §1)
//   2. Full-snapshot replace on hydrate
//   3. Lexicographic version compare; same → no-op, lower → reject (P6)
//   4. Selector miss returns undefined (UI shows skeleton — P7)
//   5. renderName: {key} placeholders; any missing ctx → undefined
//   6. Numeric ctx values are stringified
// ---------------------------------------------------------------------------

function marketDesc(
  market_type_id: string,
  name: string,
  group = "main",
  tab = "default",
): MarketDescription {
  return { market_type_id, name, group, tab };
}

function outcomeDesc(
  outcome_type_id: string,
  name: string,
): OutcomeDescription {
  return { outcome_type_id, name };
}

// =================== Empty store ===================

describe("M12 descriptions baseline: empty store", () => {
  it("when no hydration has occurred then selectors return undefined", () => {
    const store = new DescriptionsStore();
    expect(store.selectMarketDescription("1", "en")).toBeUndefined();
    expect(store.selectOutcomeDescription("home", "en")).toBeUndefined();
    expect(store.getMarketDescriptionsVersion("en")).toBeUndefined();
    expect(store.getOutcomeDescriptionsVersion("en")).toBeUndefined();
  });
});

// =================== Market hydration ===================

describe("M12 descriptions hydrate markets: first hydration per locale", () => {
  it("when hydrateMarketDescriptions runs for a locale then that locale's selectors return the entries; sibling locales remain empty", () => {
    const store = new DescriptionsStore();
    const ok = store.hydrateMarketDescriptions("en", "v1", [
      marketDesc("1", "1X2"),
      marketDesc("18", "Total"),
    ]);
    expect(ok).toBe(true);
    expect(store.selectMarketDescription("1", "en")?.name).toBe("1X2");
    expect(store.selectMarketDescription("18", "en")?.name).toBe("Total");
    expect(store.selectMarketDescription("1", "fr")).toBeUndefined();
    expect(store.getMarketDescriptionsVersion("en")).toBe("v1");
  });
});

// =================== Outcome hydration ===================

describe("M12 descriptions hydrate outcomes: first hydration per locale", () => {
  it("when hydrateOutcomeDescriptions runs for a locale then that locale's selectors return the entries; sibling locales remain empty", () => {
    const store = new DescriptionsStore();
    const ok = store.hydrateOutcomeDescriptions("en", "v1", [
      outcomeDesc("home", "Home"),
      outcomeDesc("away", "Away"),
    ]);
    expect(ok).toBe(true);
    expect(store.selectOutcomeDescription("home", "en")?.name).toBe("Home");
    expect(store.selectOutcomeDescription("away", "en")?.name).toBe("Away");
    expect(store.selectOutcomeDescription("home", "fr")).toBeUndefined();
    expect(store.getOutcomeDescriptionsVersion("en")).toBe("v1");
  });
});

// =================== Multi-locale coexistence ===================

describe("M12 descriptions multi-locale coexistence", () => {
  it("when two locales are hydrated for the same market then each locale returns its own text", () => {
    const store = new DescriptionsStore();
    store.hydrateMarketDescriptions("en", "v1", [marketDesc("1", "1X2")]);
    store.hydrateMarketDescriptions("fr", "v1", [marketDesc("1", "1N2")]);
    expect(store.selectMarketDescription("1", "en")?.name).toBe("1X2");
    expect(store.selectMarketDescription("1", "fr")?.name).toBe("1N2");
  });
});

// =================== Version semantics ===================

describe("M12 descriptions version: higher-version hydration replaces", () => {
  it("when a newer version hydration arrives then the locale's entries are replaced", () => {
    const store = new DescriptionsStore();
    store.hydrateMarketDescriptions("en", "v1", [marketDesc("1", "1X2")]);
    const ok = store.hydrateMarketDescriptions("en", "v2", [
      marketDesc("1", "Match Result"),
    ]);
    expect(ok).toBe(true);
    expect(store.selectMarketDescription("1", "en")?.name).toBe("Match Result");
    expect(store.getMarketDescriptionsVersion("en")).toBe("v2");
  });
});

describe("M12 descriptions version: same-version re-hydration is no-op", () => {
  it("when an identical-version hydration repeats then the call is a no-op and listeners do not fire", () => {
    const store = new DescriptionsStore();
    store.hydrateMarketDescriptions("en", "v2", [marketDesc("1", "1X2")]);
    const listener = vi.fn();
    store.subscribe(listener);
    const ok = store.hydrateMarketDescriptions("en", "v2", [
      marketDesc("1", "Different Name"),
    ]);
    expect(ok).toBe(false);
    expect(listener).not.toHaveBeenCalled();
    expect(store.selectMarketDescription("1", "en")?.name).toBe("1X2");
  });
});

describe("M12 descriptions version: lower-version hydration rejected", () => {
  it("when a stale hydration arrives then the existing newer version is preserved", () => {
    const store = new DescriptionsStore();
    store.hydrateMarketDescriptions("en", "v5", [marketDesc("1", "Newer")]);
    const ok = store.hydrateMarketDescriptions("en", "v3", [
      marketDesc("1", "Older"),
    ]);
    expect(ok).toBe(false);
    expect(store.selectMarketDescription("1", "en")?.name).toBe("Newer");
    expect(store.getMarketDescriptionsVersion("en")).toBe("v5");
  });
});

// =================== Per-locale version isolation ===================

describe("M12 descriptions per-locale version isolation", () => {
  it("when one locale advances version then sibling locales' versions are unaffected", () => {
    const store = new DescriptionsStore();
    store.hydrateMarketDescriptions("en", "v5", [marketDesc("1", "EN")]);
    store.hydrateMarketDescriptions("fr", "v2", [marketDesc("1", "FR")]);
    const ok = store.hydrateMarketDescriptions("fr", "v3", [
      marketDesc("1", "FR-3"),
    ]);
    expect(ok).toBe(true);
    expect(store.getMarketDescriptionsVersion("en")).toBe("v5");
    expect(store.getMarketDescriptionsVersion("fr")).toBe("v3");
    expect(store.selectMarketDescription("1", "en")?.name).toBe("EN");
    expect(store.selectMarketDescription("1", "fr")?.name).toBe("FR-3");
  });
});

// =================== renderName: literal templates ===================

describe("M12 renderName literal: no placeholders", () => {
  it("when the template has no placeholders then renderName returns the template verbatim", () => {
    expect(renderName("1X2")).toBe("1X2");
    expect(renderName("Match Result", {})).toBe("Match Result");
  });
});

// =================== renderName: substitution ===================

describe("M12 renderName substitution: single placeholder", () => {
  it("when the template contains a single placeholder and ctx supplies it then renderName substitutes it", () => {
    expect(renderName("Goalscorer: {player}", { player: "Messi" })).toBe(
      "Goalscorer: Messi",
    );
  });
});

describe("M12 renderName substitution: numeric ctx values are stringified", () => {
  it("when ctx supplies a numeric value for a placeholder then renderName stringifies it", () => {
    expect(renderName("Over {handicap}", { handicap: 2.5 })).toBe("Over 2.5");
  });
});

describe("M12 renderName substitution: multiple placeholders", () => {
  it("when the template contains multiple placeholders then all are substituted", () => {
    expect(
      renderName("{player} over {handicap}", {
        player: "Haaland",
        handicap: "1.5",
      }),
    ).toBe("Haaland over 1.5");
  });
});

// =================== renderName: missing context ===================

describe("M12 renderName missing context: returns undefined", () => {
  it("when ctx lacks a required placeholder then renderName returns undefined", () => {
    expect(renderName("Goalscorer: {player}", {})).toBeUndefined();
    expect(renderName("Goalscorer: {player}")).toBeUndefined();
  });
});

describe("M12 renderName missing context: explicit undefined is missing", () => {
  it("when ctx supplies undefined for a placeholder then renderName returns undefined", () => {
    expect(renderName("Foo {player} bar", { player: undefined })).toBeUndefined();
  });
});

// =================== renderOutcomeName integration ===================

describe("M12 renderOutcomeName: pulls template from store and substitutes", () => {
  it("when a hydrated outcome name carries a template then renderOutcomeName substitutes via ctx", () => {
    const store = new DescriptionsStore();
    store.hydrateOutcomeDescriptions("en", "v1", [
      outcomeDesc("scorer", "Score by {player}"),
    ]);
    expect(
      store.renderOutcomeName("scorer", "en", { player: "Foden" }),
    ).toBe("Score by Foden");
  });
});

describe("M12 renderOutcomeName: missing description returns undefined", () => {
  it("when the outcome is not hydrated for the locale then renderOutcomeName returns undefined", () => {
    const store = new DescriptionsStore();
    store.hydrateOutcomeDescriptions("en", "v1", [outcomeDesc("home", "Home")]);
    expect(store.renderOutcomeName("home", "fr")).toBeUndefined();
    expect(store.renderOutcomeName("unknown", "en")).toBeUndefined();
  });
});

// =================== Listeners ===================

describe("M12 descriptions listeners: fire only on real mutations", () => {
  it("when descriptions state actually changes the listener fires; otherwise it does not", () => {
    const store = new DescriptionsStore();
    const listener = vi.fn();
    store.subscribe(listener);

    store.hydrateMarketDescriptions("en", "v1", [marketDesc("1", "1X2")]);
    expect(listener).toHaveBeenCalledTimes(1);

    store.hydrateMarketDescriptions("en", "v1", [marketDesc("1", "Other")]);
    expect(listener).toHaveBeenCalledTimes(1);

    store.hydrateMarketDescriptions("en", "v0", [marketDesc("1", "Old")]);
    expect(listener).toHaveBeenCalledTimes(1);

    store.hydrateMarketDescriptions("en", "v2", [marketDesc("1", "New")]);
    expect(listener).toHaveBeenCalledTimes(2);

    store.hydrateOutcomeDescriptions("en", "v1", [outcomeDesc("home", "Home")]);
    expect(listener).toHaveBeenCalledTimes(3);
  });
});

// =================== Skeleton invariant ===================

describe("M12 skeleton invariant: missing descriptions never expose raw IDs", () => {
  it("when a description is missing then the selector returns undefined (never the raw ID)", () => {
    const store = new DescriptionsStore();
    store.hydrateMarketDescriptions("en", "v1", [marketDesc("1", "1X2")]);
    const result = store.selectMarketDescription("99", "en");
    expect(result).toBeUndefined();
    expect(result).not.toEqual(expect.objectContaining({ name: "99" }));
  });
});
