package catalog_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/catalog"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// 验收 4 — 主数据（M03）— Competition 上行处理
//
// Given a catalog.competition (objectType=3) delivery with id / region_id /
//       sport_id / name and the parent region + sport already exist
// When the catalog handler processes it
// Then a competitions row is upserted with region_id / sport_id / name
//      AND is_active=true.
func TestGiven_NewCompetitionMessage_When_Handled_Then_CompetitionRowUpserted(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	ctx := context.Background()

	require.NoError(t, h.HandleSport(ctx, feed.MsgCatalogSport,
		feed.Envelope{Payload: []byte(`{"id":1,"name":"Football"}`)}, [16]byte{}))
	require.NoError(t, h.HandleRegion(ctx, feed.MsgCatalogRegion,
		feed.Envelope{Payload: []byte(`{"id":11,"sportId":1,"name":"Spain"}`)}, [16]byte{}))

	require.NoError(t, h.HandleCompetition(ctx, feed.MsgCatalogComp,
		feed.Envelope{Payload: []byte(`{"id":111,"regionId":11,"sportId":1,"name":"La Liga"}`)}, [16]byte{}))

	_, _, comps, _, _ := repo.snapshot()
	require.Equal(t, catalog.Competition{
		ID: 111, RegionID: 11, SportID: 1, Name: "La Liga", IsActive: true,
	}, comps[111])
	require.Equal(t, 1, repo.compUpsertCnt)
}

// Given a catalog.competition delivery where sport_id is missing on the
//       payload but region_id is present and the regions row already
//       carries a sport_id
// When the catalog handler processes it
// Then the competitions.sport_id is filled from regions.sport_id
//      AND the row is inserted without an FK violation.
func TestGiven_CompetitionMissingSport_When_Handled_Then_SportInheritedFromRegion(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	ctx := context.Background()

	require.NoError(t, h.HandleSport(ctx, feed.MsgCatalogSport,
		feed.Envelope{Payload: []byte(`{"id":1,"name":"Football"}`)}, [16]byte{}))
	require.NoError(t, h.HandleRegion(ctx, feed.MsgCatalogRegion,
		feed.Envelope{Payload: []byte(`{"id":11,"sportId":1,"name":"Spain"}`)}, [16]byte{}))

	require.NoError(t, h.HandleCompetition(ctx, feed.MsgCatalogComp,
		feed.Envelope{Payload: []byte(`{"id":111,"regionId":11,"name":"La Liga"}`)}, [16]byte{}))

	_, _, comps, _, _ := repo.snapshot()
	require.Equal(t, int32(1), comps[111].SportID, "sport_id must be inherited from the parent region")
	require.Equal(t, int32(11), comps[111].RegionID)
}

// Given an existing competitions row
// When a later catalog.competition delivery arrives renaming it
// Then competitions.name is updated AND updated_at advances AND the
//      region_id / sport_id linkage is unchanged.
func TestGiven_ExistingCompetition_When_NameChanges_Then_RowRenamedLinkagePreserved(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	ctx := context.Background()

	require.NoError(t, h.HandleSport(ctx, feed.MsgCatalogSport,
		feed.Envelope{Payload: []byte(`{"id":1,"name":"Football"}`)}, [16]byte{}))
	require.NoError(t, h.HandleRegion(ctx, feed.MsgCatalogRegion,
		feed.Envelope{Payload: []byte(`{"id":11,"sportId":1,"name":"Spain"}`)}, [16]byte{}))
	require.NoError(t, h.HandleCompetition(ctx, feed.MsgCatalogComp,
		feed.Envelope{Payload: []byte(`{"id":111,"regionId":11,"sportId":1,"name":"La Liga"}`)}, [16]byte{}))

	require.NoError(t, h.HandleCompetition(ctx, feed.MsgCatalogComp,
		feed.Envelope{Payload: []byte(`{"id":111,"regionId":11,"sportId":1,"name":"Primera División"}`)}, [16]byte{}))

	_, _, comps, _, _ := repo.snapshot()
	require.Equal(t, "Primera División", comps[111].Name)
	require.Equal(t, int32(11), comps[111].RegionID)
	require.Equal(t, int32(1), comps[111].SportID)
	require.Equal(t, 2, repo.compUpsertCnt)
}
