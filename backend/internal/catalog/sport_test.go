package catalog_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/catalog"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// 验收 4 — 主数据（M03）— Sport 上行处理
//
// Given a catalog.sport (objectType=1) delivery with id / name
// When the catalog handler processes it
// Then a sports row is upserted with id / name / is_active=true
//      AND updated_at is set to now().
func TestGiven_NewSportMessage_When_Handled_Then_SportRowUpserted(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)

	payload := []byte(`{"id": 1, "name": "Football"}`)
	env := feed.Envelope{ObjectType: 1, Payload: payload}

	require.NoError(t, h.HandleSport(context.Background(), feed.MsgCatalogSport, env, [16]byte{}))

	sports, _, _, _, _ := repo.snapshot()
	require.Contains(t, sports, int32(1))
	require.Equal(t, "Football", sports[1].Name)
	require.True(t, sports[1].IsActive)
	require.Equal(t, 1, repo.sportUpsertCnt)
}

// Given an existing sports row with name = "Football"
// When a later catalog.sport delivery arrives with the same id but
//      name = "Soccer"
// Then sports.name is updated to "Soccer" AND updated_at advances.
func TestGiven_ExistingSport_When_NameChanges_Then_RowRenamed(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	ctx := context.Background()

	require.NoError(t, h.HandleSport(ctx, feed.MsgCatalogSport,
		feed.Envelope{ObjectType: 1, Payload: []byte(`{"id":1,"name":"Football"}`)}, [16]byte{}))
	require.NoError(t, h.HandleSport(ctx, feed.MsgCatalogSport,
		feed.Envelope{ObjectType: 1, Payload: []byte(`{"id":1,"name":"Soccer"}`)}, [16]byte{}))

	sports, _, _, _, _ := repo.snapshot()
	require.Equal(t, "Soccer", sports[1].Name)
	require.Equal(t, 2, repo.sportUpsertCnt, "rename must trigger a fresh upsert")
}

// Given a sport.removed signal (catalog.sport payload with removed=true)
// When the catalog handler processes it
// Then sports.is_active is flipped to false (soft delete) AND existing
//      regions / competitions / matches referencing that sport are NOT
//      cascaded — only the visibility flag changes.
func TestGiven_SportRemoved_When_Handled_Then_SoftDeleted(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	ctx := context.Background()

	// Seed sport + a region + a competition linked to it.
	require.NoError(t, h.HandleSport(ctx, feed.MsgCatalogSport,
		feed.Envelope{Payload: []byte(`{"id":1,"name":"Football"}`)}, [16]byte{}))
	require.NoError(t, h.HandleRegion(ctx, feed.MsgCatalogRegion,
		feed.Envelope{Payload: []byte(`{"id":11,"sportId":1,"name":"Spain"}`)}, [16]byte{}))
	require.NoError(t, h.HandleCompetition(ctx, feed.MsgCatalogComp,
		feed.Envelope{Payload: []byte(`{"id":111,"regionId":11,"sportId":1,"name":"La Liga"}`)}, [16]byte{}))

	require.NoError(t, h.HandleSport(ctx, feed.MsgCatalogSport,
		feed.Envelope{Payload: []byte(`{"id":1,"name":"Football","removed":true}`)}, [16]byte{}))

	sports, regions, comps, _, _ := repo.snapshot()
	require.False(t, sports[1].IsActive, "sport must be marked inactive")
	require.True(t, regions[11].IsActive, "region must NOT be cascaded")
	require.True(t, comps[111].IsActive, "competition must NOT be cascaded")
}

// Given two concurrent catalog.sport deliveries for the same id
// When both are handled in parallel
// Then exactly one row exists and last-writer-wins on name AND no
//      "duplicate key" error escapes the handler.
func TestGiven_ConcurrentSportUpserts_When_Handled_Then_SingleRow(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	ctx := context.Background()

	var wg sync.WaitGroup
	for _, name := range []string{"Football", "Soccer", "FÚtbol", "Calcio"} {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			err := h.HandleSport(ctx, feed.MsgCatalogSport, feed.Envelope{
				Payload: []byte(`{"id":1,"name":"` + n + `"}`),
			}, [16]byte{})
			require.NoError(t, err)
		}(name)
	}
	wg.Wait()

	sports, _, _, _, _ := repo.snapshot()
	require.Len(t, sports, 1, "exactly one sports row must remain")
	require.True(t, sports[1].IsActive)
	require.Equal(t, 4, repo.sportUpsertCnt, "every concurrent call must reach the repo")
}
