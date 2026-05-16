import type {
  GetMarketDescriptionsResponse,
  GetOutcomeDescriptionsResponse,
  MarketDescription,
  OutcomeDescription,
} from "@/contract/rest";

// ---------------------------------------------------------------------------
// M12 — DescriptionsStore
//
// Caches market + outcome description bundles per (lang, version+etag). The
// store is the data model only; the REST/ETag conditional fetcher
// (descriptionService) sits above it and feeds these hydrate* methods.
//
// Idempotency mirrors M08/M09/M11:
//   - hydrate* is a no-op when version AND etag both match the existing
//     bundle. Listeners do not fire on no-op hydrates.
//   - A different version (or a fresh etag) replaces the bundle wholesale.
//   - setLang is a no-op when the new lang equals currentLang.
//
// Languages are isolated: hydrating 'zh' does not evict 'en'. This is what
// powers the M12 acceptance "切换语言不重拉描述结构，只换文案" — once a
// language bucket is filled, switching is instantaneous.
// ---------------------------------------------------------------------------

export interface DescriptionsBundleMeta {
  version: string;
  etag?: string;
}

export interface DescriptionsStoreOptions {
  defaultLang: string;
}

interface MarketBundle {
  meta: DescriptionsBundleMeta;
  byId: Map<string, MarketDescription>;
}

interface OutcomeBundle {
  meta: DescriptionsBundleMeta;
  byId: Map<string, OutcomeDescription>;
}

export class DescriptionsStore {
  private readonly markets = new Map<string, MarketBundle>();
  private readonly outcomes = new Map<string, OutcomeBundle>();
  private currentLang: string;
  private readonly listeners = new Set<() => void>();

  constructor(opts: DescriptionsStoreOptions) {
    this.currentLang = opts.defaultLang;
  }

  getLang(): string {
    return this.currentLang;
  }

  setLang(lang: string): boolean {
    if (this.currentLang === lang) return false;
    this.currentLang = lang;
    this.notify();
    return true;
  }

  hydrateMarkets(
    lang: string,
    resp: GetMarketDescriptionsResponse,
    etag?: string,
  ): boolean {
    const existing = this.markets.get(lang);
    if (
      existing &&
      existing.meta.version === resp.version &&
      existing.meta.etag === etag
    ) {
      return false;
    }
    const byId = new Map<string, MarketDescription>();
    for (const d of resp.descriptions) byId.set(d.market_type_id, cloneMarket(d));
    this.markets.set(lang, { meta: buildMeta(resp.version, etag), byId });
    this.notify();
    return true;
  }

  hydrateOutcomes(
    lang: string,
    resp: GetOutcomeDescriptionsResponse,
    etag?: string,
  ): boolean {
    const existing = this.outcomes.get(lang);
    if (
      existing &&
      existing.meta.version === resp.version &&
      existing.meta.etag === etag
    ) {
      return false;
    }
    const byId = new Map<string, OutcomeDescription>();
    for (const d of resp.descriptions) byId.set(d.outcome_type_id, cloneOutcome(d));
    this.outcomes.set(lang, { meta: buildMeta(resp.version, etag), byId });
    this.notify();
    return true;
  }

  selectMarket(marketTypeId: string, lang?: string): MarketDescription | undefined {
    const bundle = this.markets.get(lang ?? this.currentLang);
    const entry = bundle?.byId.get(marketTypeId);
    return entry ? cloneMarket(entry) : undefined;
  }

  selectOutcome(outcomeTypeId: string, lang?: string): OutcomeDescription | undefined {
    const bundle = this.outcomes.get(lang ?? this.currentLang);
    const entry = bundle?.byId.get(outcomeTypeId);
    return entry ? cloneOutcome(entry) : undefined;
  }

  getMarketsMeta(lang?: string): DescriptionsBundleMeta | undefined {
    const bundle = this.markets.get(lang ?? this.currentLang);
    return bundle ? cloneMeta(bundle.meta) : undefined;
  }

  getOutcomesMeta(lang?: string): DescriptionsBundleMeta | undefined {
    const bundle = this.outcomes.get(lang ?? this.currentLang);
    return bundle ? cloneMeta(bundle.meta) : undefined;
  }

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

function buildMeta(version: string, etag?: string): DescriptionsBundleMeta {
  return etag !== undefined ? { version, etag } : { version };
}

function cloneMeta(m: DescriptionsBundleMeta): DescriptionsBundleMeta {
  return m.etag !== undefined ? { version: m.version, etag: m.etag } : { version: m.version };
}

function cloneMarket(d: MarketDescription): MarketDescription {
  return {
    market_type_id: d.market_type_id,
    name: d.name,
    group: d.group,
    tab: d.tab,
  };
}

function cloneOutcome(d: OutcomeDescription): OutcomeDescription {
  return { outcome_type_id: d.outcome_type_id, name: d.name };
}
