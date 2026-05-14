package catalog_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/catalog"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// 验收 4 — 主数据（M03）— Region 上行处理
//
// Given a catalog.region (objectType=2) delivery with id / sport_id / name
//       and the parent sport already exists
// When the catalog handler processes it
// Then a regions row is upserted referencing sport_id AND is_active=true.
func TestGiven_NewRegionMessage_When_Handled_Then_RegionRowUpserted(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	ctx := context.Background()

	require.NoError(t, h.HandleSport(ctx, feed.MsgCatalogSport,
		feed.Envelope{Payload: []byte(`{"id":1,"name":"Football"}`)}, [16]byte{}))

	require.NoError(t, h.HandleRegion(ctx, feed.MsgCatalogRegion,
		feed.Envelope{Payload: []byte(`{"id":11,"sportId":1,"name":"Spain"}`)}, [16]byte{}))

	_, regions, _, _, _ := repo.snapshot()
	require.Equal(t, catalog.Region{ID: 11, SportID: 1, Name: "Spain", IsActive: true}, regions[11])
	require.Equal(t, 1, repo.regionUpsertCnt)
	// Parent sport was already there — the region handler must not have
	// auto-stubbed an extra sports row.
	require.Equal(t, 1, repo.sportUpsertCnt)
}

// Given a catalog.region delivery whose sport_id does NOT yet exist in sports
// When the catalog handler processes it
// Then a stub sports row is auto-inserted (id + placeholder name)
//      AND the regions row is created without violating the FK constraint.
func TestGiven_RegionBeforeSport_When_Handled_Then_StubSportAutoInserted(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)

	require.NoError(t, h.HandleRegion(context.Background(), feed.MsgCatalogRegion,
		feed.Envelope{Payload: []byte(`{"id":11,"sportId":7,"name":"Spain"}`)}, [16]byte{}))

	sports, regions, _, _, _ := repo.snapshot()
	require.Contains(t, sports, int32(7), "stub sport must be auto-inserted")
	require.True(t, sports[7].IsActive)
	require.Equal(t, "", sports[7].Name, "stub name stays empty until the real sport delivery arrives")
	require.Equal(t, int32(7), regions[11].SportID)
}

// Given an existing regions row with name = "Spain"
// When a later catalog.region delivery arrives with name = "España"
// Then regions.name is updated to "España" AND updated_at advances.
func TestGiven_ExistingRegion_When_NameChanges_Then_RowRenamed(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	ctx := context.Background()

	require.NoError(t, h.HandleSport(ctx, feed.MsgCatalogSport,
		feed.Envelope{Payload: []byte(`{"id":1,"name":"Football"}`)}, [16]byte{}))
	require.NoError(t, h.HandleRegion(ctx, feed.MsgCatalogRegion,
		feed.Envelope{Payload: []byte(`{"id":11,"sportId":1,"name":"Spain"}`)}, [16]byte{}))
	require.NoError(t, h.HandleRegion(ctx, feed.MsgCatalogRegion,
		feed.Envelope{Payload: []byte(`{"id":11,"sportId":1,"name":"España"}`)}, [16]byte{}))

	_, regions, _, _, _ := repo.snapshot()
	require.Equal(t, "España", regions[11].Name)
	require.Equal(t, 2, repo.regionUpsertCnt)
}
