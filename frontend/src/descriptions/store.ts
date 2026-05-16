import type { MarketDescription, OutcomeDescription } from "@/contract/rest";

// ---------------------------------------------------------------------------
// M12 — DescriptionsStore + renderName
//
// Caches market & outcome descriptions per locale, hydrated by REST (P7).
// The store performs no IO; the (future) descriptionService is responsible
// for ETag/If-None-Match coordination on the wire.
//
// Locked decisions (per PR thread):
//   1. Per-locale storage: byLocale buckets coexist; language switch is a
//      pure read of a different bucket (acceptance §1).
//   2. Hydration is a full-snapshot replace per (kind, locale).
//   3. Version semantics: lexicographic compare; only strictly-newer
//      versions are accepted (same-version no-op, lower rejected; P6).
//      The BFF is responsible for monotonic ETags.
//   4. Selector miss → undefined (UI shows skeleton, never raw ID — P7).
//   5. renderName placeholder syntax: `{key}`; any missing/undefined value
//      collapses the render to undefined (UI shows skeleton).
//   6. Numeric ctx values are stringified via String(value).
// ---------------------------------------------------------------------------

type DescKind = "market" | "outcome";

function versionKey(kind: DescKind, locale: string): string {
  return `${kind}:${locale}`;
}

export type RenderContext = Record<string, string | number | undefined>;

export class DescriptionsStore {
  private readonly markets = new Map<string, Map<string, MarketDescription>>();
  private readonly outcomes = new Map<string, Map<string, OutcomeDescription>>();
  private readonly versions = new Map<string, string>();
  private readonly listeners = new Set<() => void>();

  hydrateMarketDescriptions(
    locale: string,
    version: string,
    descriptions: ReadonlyArray<MarketDescription>,
  ): boolean {
    if (!this.acceptVersion("market", locale, version)) return false;
    const bucket = new Map<string, MarketDescription>();
    for (const d of descriptions) bucket.set(d.market_type_id, { ...d });
    this.markets.set(locale, bucket);
    this.versions.set(versionKey("market", locale), version);
    this.notify();
    return true;
  }

  hydrateOutcomeDescriptions(
    locale: string,
    version: string,
    descriptions: ReadonlyArray<OutcomeDescription>,
  ): boolean {
    if (!this.acceptVersion("outcome", locale, version)) return false;
    const bucket = new Map<string, OutcomeDescription>();
    for (const d of descriptions) bucket.set(d.outcome_type_id, { ...d });
    this.outcomes.set(locale, bucket);
    this.versions.set(versionKey("outcome", locale), version);
    this.notify();
    return true;
  }

  selectMarketDescription(
    marketTypeId: string,
    locale: string,
  ): MarketDescription | undefined {
    const d = this.markets.get(locale)?.get(marketTypeId);
    return d ? { ...d } : undefined;
  }

  selectOutcomeDescription(
    outcomeTypeId: string,
    locale: string,
  ): OutcomeDescription | undefined {
    const d = this.outcomes.get(locale)?.get(outcomeTypeId);
    return d ? { ...d } : undefined;
  }

  getMarketDescriptionsVersion(locale: string): string | undefined {
    return this.versions.get(versionKey("market", locale));
  }

  getOutcomeDescriptionsVersion(locale: string): string | undefined {
    return this.versions.get(versionKey("outcome", locale));
  }

  renderOutcomeName(
    outcomeTypeId: string,
    locale: string,
    ctx?: RenderContext,
  ): string | undefined {
    const d = this.outcomes.get(locale)?.get(outcomeTypeId);
    if (!d) return undefined;
    return renderName(d.name, ctx);
  }

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  private acceptVersion(
    kind: DescKind,
    locale: string,
    incoming: string,
  ): boolean {
    const current = this.versions.get(versionKey(kind, locale));
    if (current === undefined) return true;
    return incoming > current;
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

// ---------------------------------------------------------------------------
// renderName — pure template helper
//
// Substitutes `{key}` placeholders from ctx. Returns undefined if ANY
// placeholder lacks a value (so the UI can render a skeleton rather than
// a half-rendered string).
// ---------------------------------------------------------------------------

const PLACEHOLDER_RE = /\{([^}]+)\}/g;

export function renderName(
  template: string,
  ctx?: RenderContext,
): string | undefined {
  let missing = false;
  const out = template.replace(PLACEHOLDER_RE, (_match, key: string) => {
    const value = ctx?.[key];
    if (value === undefined || value === null) {
      missing = true;
      return "";
    }
    return String(value);
  });
  return missing ? undefined : out;
}
