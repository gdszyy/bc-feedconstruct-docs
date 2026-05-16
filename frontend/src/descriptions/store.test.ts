import { describe, expect, it, vi } from "vitest";

import { DescriptionsStore } from "./store";

// ---------------------------------------------------------------------------
// M12 — DescriptionsStore
//
// Locked decisions (PR thread):
//   - Caches market + outcome descriptions per (lang, version+etag)
//   - "切换语言不重拉描述结构，只换文案" — lookups go through the current
//     lang's bundle; hydrating a new lang doesn't evict other langs.
//   - hydrate* is idempotent on (version + etag): no-op + no listener call
//   - Newer version replaces the lang's bundle wholesale
//   - Markets and outcomes carry independent versions/etags per lang
// ---------------------------------------------------------------------------

const en1 = {
  version: "v1",
  descriptions: [
    { market_type_id: "1", name: "Full Time Result", group: "Main", tab: "Main" },
    { market_type_id: "18", name: "Total Goals", group: "Goals", tab: "Goals" },
  ],
};

const en2 = {
  version: "v2",
  descriptions: [
    { market_type_id: "1", name: "1X2", group: "Main", tab: "Main" },
    { market_type_id: "19", name: "Both Teams to Score", group: "Goals", tab: "Goals" },
  ],
};

const zh1 = {
  version: "v1",
  descriptions: [
    { market_type_id: "1", name: "全场赛果", group: "主要", tab: "主要" },
  ],
};

const outcomes_en1 = {
  version: "v1",
  descriptions: [
    { outcome_type_id: "1", name: "Home" },
    { outcome_type_id: "2", name: "Draw" },
    { outcome_type_id: "3", name: "Away" },
  ],
};

// =================== Empty store ===================

// Given a fresh DescriptionsStore with defaultLang='en'
// When selectMarket / selectOutcome / getMarketsMeta / getOutcomesMeta are queried
// Then they return undefined; getLang() returns 'en'
describe("M12 descriptions baseline: empty store yields nothing", () => {
  it("when nothing has been hydrated then selectors return undefined and meta is undefined", () => {
    const store = new DescriptionsStore({ defaultLang: "en" });
    expect(store.getLang()).toBe("en");
    expect(store.selectMarket("1")).toBeUndefined();
    expect(store.selectOutcome("1")).toBeUndefined();
    expect(store.getMarketsMeta()).toBeUndefined();
    expect(store.getOutcomesMeta()).toBeUndefined();
  });
});

// =================== Hydrate markets ===================

// Given an empty store
// When hydrateMarkets('en', {version: 'v1', descriptions: [...]}, etag: 'W/"abc"') is invoked
// Then selectMarket('1') returns the description; getMarketsMeta('en') returns {version, etag}; listener fires
describe("M12 markets hydrate: first hydrate seeds the bundle", () => {
  it("when markets are hydrated for a lang then lookups + meta become available and the listener fires", () => {
    const store = new DescriptionsStore({ defaultLang: "en" });
    const listener = vi.fn();
    store.subscribe(listener);

    const ok = store.hydrateMarkets("en", en1, 'W/"abc"');
    expect(ok).toBe(true);
    expect(listener).toHaveBeenCalledTimes(1);
    expect(store.selectMarket("1")).toEqual({
      market_type_id: "1",
      name: "Full Time Result",
      group: "Main",
      tab: "Main",
    });
    expect(store.getMarketsMeta()).toEqual({ version: "v1", etag: 'W/"abc"' });
  });
});

// Given an existing bundle at version='v1' / etag='W/"abc"'
// When hydrateMarkets is called again with the SAME version + etag
// Then the call is a no-op (returns false); listener does NOT fire
describe("M12 markets hydrate: idempotent on identical version + etag", () => {
  it("when the same markets bundle is re-applied then the store skips it and no listener fires", () => {
    const store = new DescriptionsStore({ defaultLang: "en" });
    store.hydrateMarkets("en", en1, 'W/"abc"');
    const listener = vi.fn();
    store.subscribe(listener);

    const ok = store.hydrateMarkets("en", en1, 'W/"abc"');
    expect(ok).toBe(false);
    expect(listener).not.toHaveBeenCalled();
  });
});

// Given an existing bundle at version='v1'
// When hydrateMarkets is called with version='v2' (newer payload)
// Then the bundle replaces wholesale; listener fires; old ids dropped if not in new payload
describe("M12 markets hydrate: newer version replaces wholesale", () => {
  it("when a newer markets version arrives then the cache replaces and stale ids are dropped", () => {
    const store = new DescriptionsStore({ defaultLang: "en" });
    store.hydrateMarkets("en", en1);

    const ok = store.hydrateMarkets("en", en2);
    expect(ok).toBe(true);
    expect(store.getMarketsMeta()).toEqual({ version: "v2" });
    expect(store.selectMarket("1")?.name).toBe("1X2");
    expect(store.selectMarket("18")).toBeUndefined();
    expect(store.selectMarket("19")?.name).toBe("Both Teams to Score");
  });
});

// =================== Hydrate outcomes ===================

// Given an empty store
// When hydrateOutcomes('en', {version, descriptions}) is invoked
// Then selectOutcome works and getOutcomesMeta exposes version/etag; listener fires
describe("M12 outcomes hydrate: first hydrate seeds the outcomes bundle", () => {
  it("when outcomes are hydrated for a lang then the outcome lookups return the data", () => {
    const store = new DescriptionsStore({ defaultLang: "en" });
    const ok = store.hydrateOutcomes("en", outcomes_en1, 'W/"out-1"');
    expect(ok).toBe(true);
    expect(store.selectOutcome("1")).toEqual({ outcome_type_id: "1", name: "Home" });
    expect(store.selectOutcome("3")).toEqual({ outcome_type_id: "3", name: "Away" });
    expect(store.getOutcomesMeta()).toEqual({ version: "v1", etag: 'W/"out-1"' });
  });
});

// Given outcomes already hydrated at version='v1'
// When hydrateOutcomes is called again with the same version
// Then it's a no-op; listener does not fire
describe("M12 outcomes hydrate: idempotent on identical version", () => {
  it("when the same outcomes bundle is re-applied then the store skips it", () => {
    const store = new DescriptionsStore({ defaultLang: "en" });
    store.hydrateOutcomes("en", outcomes_en1, 'W/"out-1"');
    const listener = vi.fn();
    store.subscribe(listener);

    const ok = store.hydrateOutcomes("en", outcomes_en1, 'W/"out-1"');
    expect(ok).toBe(false);
    expect(listener).not.toHaveBeenCalled();
  });
});

// =================== Per-lang isolation ===================

// Given markets hydrated for 'en'
// When markets are hydrated for 'zh' (different content)
// Then both lang buckets coexist and lookups disambiguate by lang
describe("M12 lang isolation: 'en' and 'zh' bundles coexist", () => {
  it("when two languages are hydrated then each lang's selectors return its own data", () => {
    const store = new DescriptionsStore({ defaultLang: "en" });
    store.hydrateMarkets("en", en1);
    store.hydrateMarkets("zh", zh1);

    expect(store.selectMarket("1", "en")?.name).toBe("Full Time Result");
    expect(store.selectMarket("1", "zh")?.name).toBe("全场赛果");
    expect(store.selectMarket("18", "zh")).toBeUndefined();
    expect(store.selectMarket("18", "en")?.name).toBe("Total Goals");
  });
});

// =================== setLang ===================

// Given a store at currentLang='en' with both 'en' and 'zh' hydrated
// When setLang('zh') is invoked
// Then getLang() reports 'zh'; selectMarket(id) (no explicit lang) reads from 'zh'; listener fires
describe("M12 setLang: changes default lookup language", () => {
  it("when setLang switches then selectMarket without an explicit lang reads the new lang", () => {
    const store = new DescriptionsStore({ defaultLang: "en" });
    store.hydrateMarkets("en", en1);
    store.hydrateMarkets("zh", zh1);

    const listener = vi.fn();
    store.subscribe(listener);

    const ok = store.setLang("zh");
    expect(ok).toBe(true);
    expect(store.getLang()).toBe("zh");
    expect(store.selectMarket("1")?.name).toBe("全场赛果");
    expect(listener).toHaveBeenCalledTimes(1);
  });
});

// Given a store at currentLang='en'
// When setLang('en') is invoked again with the same value
// Then it's a no-op; listener does not fire
describe("M12 setLang: idempotent on identical lang", () => {
  it("when setLang is called with the current lang then no listener fires", () => {
    const store = new DescriptionsStore({ defaultLang: "en" });
    const listener = vi.fn();
    store.subscribe(listener);
    expect(store.setLang("en")).toBe(false);
    expect(listener).not.toHaveBeenCalled();
  });
});

// =================== Explicit-lang selector ===================

// Given currentLang='en' with both 'en' and 'zh' hydrated
// When selectMarket(id, 'zh') is invoked
// Then it returns the 'zh' entry regardless of currentLang
describe("M12 selectors: explicit lang overrides currentLang", () => {
  it("when an explicit lang is provided then the selector reads that lang's bundle", () => {
    const store = new DescriptionsStore({ defaultLang: "en" });
    store.hydrateMarkets("en", en1);
    store.hydrateMarkets("zh", zh1);

    expect(store.getLang()).toBe("en");
    expect(store.selectMarket("1", "zh")?.name).toBe("全场赛果");
    expect(store.selectMarket("1")?.name).toBe("Full Time Result");
  });
});

// =================== Snapshot for missing ids ===================

// Given a hydrated 'en' bundle
// When selectMarket queries an id that is NOT in the bundle
// Then it returns undefined (UI is expected to render a skeleton, never an ID)
describe("M12 selectors: unknown ids return undefined", () => {
  it("when a market id is absent from the hydrated bundle then selectMarket returns undefined", () => {
    const store = new DescriptionsStore({ defaultLang: "en" });
    store.hydrateMarkets("en", en1);
    expect(store.selectMarket("999")).toBeUndefined();
    expect(store.selectOutcome("999")).toBeUndefined();
  });
});

// =================== Listener notifications ===================

// Given a listener subscribed to the store
// When hydrate* or setLang actually changes state
// Then the listener fires once per real change
// And idempotent re-hydrates / no-op setLang do NOT fire
describe("M12 listeners: notified only on real state changes", () => {
  it("when hydrate or setLang change state the listener fires; duplicates do not", () => {
    const store = new DescriptionsStore({ defaultLang: "en" });
    const listener = vi.fn();
    store.subscribe(listener);

    store.hydrateMarkets("en", en1, 'W/"abc"');
    expect(listener).toHaveBeenCalledTimes(1);

    store.hydrateMarkets("en", en1, 'W/"abc"'); // dedup
    expect(listener).toHaveBeenCalledTimes(1);

    store.hydrateOutcomes("en", outcomes_en1);
    expect(listener).toHaveBeenCalledTimes(2);

    store.hydrateOutcomes("en", outcomes_en1); // dedup
    expect(listener).toHaveBeenCalledTimes(2);

    store.setLang("zh");
    expect(listener).toHaveBeenCalledTimes(3);

    store.setLang("zh"); // dedup
    expect(listener).toHaveBeenCalledTimes(3);
  });
});
