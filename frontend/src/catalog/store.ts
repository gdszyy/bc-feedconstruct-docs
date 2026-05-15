import type {
  SportRemovedPayload,
  SportUpsertedPayload,
  TournamentRemovedPayload,
  TournamentUpsertedPayload,
} from "@/contract/events";

export interface SportRecord {
  sport_id: string;
  name_translations: Record<string, string>;
  sort_order: number;
}

export interface CategoryRecord {
  category_id: string;
  sport_id: string;
}

export interface TournamentRecord {
  tournament_id: string;
  sport_id: string;
  category_id: string;
  name_translations: Record<string, string>;
}

export interface CatalogSnapshot {
  sports: SportRecord[];
  tournaments: TournamentRecord[];
}

/**
 * Tree-shaped cache for Sport → Category → Tournament.
 *
 * Merge policy:
 * - apply*Upserted is authoritative (last write wins by id).
 * - hydrateSnapshot only ADDS unseen records; it never overrides an entry that
 *   was already populated by an earlier increment. This honours M10's rule
 *   that a snapshot cannot regress past live increments.
 * - sport.removed cascades to that sport's tournaments and implicit
 *   categories to keep the tree free of dangling references.
 */
export class CatalogStore {
  private readonly sports = new Map<string, SportRecord>();
  private readonly tournaments = new Map<string, TournamentRecord>();
  private readonly categories = new Map<string, CategoryRecord>();
  private readonly sportTournaments = new Map<string, Set<string>>();
  private readonly listeners = new Set<() => void>();

  hydrateSnapshot(snap: CatalogSnapshot): void {
    for (const s of snap.sports) this.upsertSport(s, "snapshot");
    for (const t of snap.tournaments) this.upsertTournament(t, "snapshot");
    this.notify();
  }

  applySportUpserted(payload: SportUpsertedPayload): void {
    this.upsertSport(
      {
        sport_id: payload.sport_id,
        name_translations: payload.name_translations,
        sort_order: payload.sort_order,
      },
      "increment",
    );
    this.notify();
  }

  applySportRemoved(payload: SportRemovedPayload): void {
    if (!this.sports.has(payload.sport_id)) return;
    this.sports.delete(payload.sport_id);

    const tournamentIds = this.sportTournaments.get(payload.sport_id);
    if (tournamentIds) {
      for (const tid of tournamentIds) this.tournaments.delete(tid);
      this.sportTournaments.delete(payload.sport_id);
    }

    for (const [cid, cat] of this.categories) {
      if (cat.sport_id === payload.sport_id) this.categories.delete(cid);
    }
    this.notify();
  }

  applyTournamentUpserted(payload: TournamentUpsertedPayload): void {
    this.upsertTournament(
      {
        tournament_id: payload.tournament_id,
        sport_id: payload.sport_id,
        category_id: payload.category_id,
        name_translations: payload.name_translations,
      },
      "increment",
    );
    this.notify();
  }

  applyTournamentRemoved(payload: TournamentRemovedPayload): void {
    const t = this.tournaments.get(payload.tournament_id);
    if (!t) return;
    this.tournaments.delete(payload.tournament_id);
    this.sportTournaments.get(t.sport_id)?.delete(payload.tournament_id);
    this.notify();
  }

  listSports(): SportRecord[] {
    return [...this.sports.values()].sort(
      (a, b) => a.sort_order - b.sort_order,
    );
  }

  listTournaments(sportId: string): TournamentRecord[] {
    const tournamentIds = this.sportTournaments.get(sportId);
    if (!tournamentIds) return [];
    const result: TournamentRecord[] = [];
    for (const tid of tournamentIds) {
      const record = this.tournaments.get(tid);
      if (record) result.push(record);
    }
    return result;
  }

  getSportName(sportId: string, locale: string): string | undefined {
    const sport = this.sports.get(sportId);
    if (!sport) return undefined;
    return sport.name_translations[locale];
  }

  hasCategory(categoryId: string): boolean {
    return this.categories.has(categoryId);
  }

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => this.listeners.delete(handler);
  }

  private upsertSport(
    record: SportRecord,
    source: "snapshot" | "increment",
  ): void {
    if (source === "snapshot" && this.sports.has(record.sport_id)) return;
    this.sports.set(record.sport_id, record);
  }

  private upsertTournament(
    record: TournamentRecord,
    source: "snapshot" | "increment",
  ): void {
    const exists = this.tournaments.has(record.tournament_id);
    if (source === "snapshot" && exists) {
      // Snapshot must not regress an increment — but the category index still
      // needs to know about the (sport, category) pair if it slipped through.
      this.ensureIndices(record);
      return;
    }
    this.tournaments.set(record.tournament_id, record);
    this.ensureIndices(record);
  }

  private ensureIndices(record: TournamentRecord): void {
    if (!this.sportTournaments.has(record.sport_id)) {
      this.sportTournaments.set(record.sport_id, new Set());
    }
    this.sportTournaments.get(record.sport_id)!.add(record.tournament_id);
    if (!this.categories.has(record.category_id)) {
      this.categories.set(record.category_id, {
        category_id: record.category_id,
        sport_id: record.sport_id,
      });
    }
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}
