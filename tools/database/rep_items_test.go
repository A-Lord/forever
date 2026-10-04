package database

// Runs the rep item query on a fixture shaped like the game files, holding only the columns it reads,
// so it needs no game files database.

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

const repItemsFixture = `
CREATE TABLE Faction (ID INTEGER PRIMARY KEY, Name_lang TEXT, ReputationRaceMasks0_0 INTEGER);
CREATE TABLE ItemSparse (ID INTEGER PRIMARY KEY, MinFactionID INTEGER, MinReputation INTEGER);
INSERT INTO Faction (ID, Name_lang, ReputationRaceMasks0_0) VALUES
	(509, 'The League of Arathor', -1321907123), (510, 'The Defilers', 1309324210),
	(270, 'Zandalar Tribe', -1), (910, 'Brood of Nozdormu', -1), (2758, 'Nightclaw Druids', 0);
INSERT INTO ItemSparse (ID, MinFactionID, MinReputation) VALUES
	(20070, 509, 7), (20041, 509, 5), (20068, 510, 7), (19764, 270, 6), (21196, 910, 3), (276929, 2758, 4), (720, 0, 0);
`

func TestRepItemsReadFactionStandingAndSide(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "client.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(repItemsFixture); err != nil {
		t.Fatalf("building the fixture: %v", err)
	}
	sources, names, err := LoadRepItems(&DBHelper{db: db})
	if err != nil {
		t.Fatal(err)
	}

	want := map[int]*proto.RepSource{
		20070:  {RepFactionId: 509, RepLevel: proto.RepLevel_RepLevelExalted, FactionId: proto.Faction_Alliance},
		20041:  {RepFactionId: 509, RepLevel: proto.RepLevel_RepLevelHonored, FactionId: proto.Faction_Alliance},
		20068:  {RepFactionId: 510, RepLevel: proto.RepLevel_RepLevelExalted, FactionId: proto.Faction_Horde},
		19764:  {RepFactionId: 270, RepLevel: proto.RepLevel_RepLevelRevered, FactionId: proto.Faction_Unknown},
		21196:  {RepFactionId: 910, RepLevel: proto.RepLevel_RepLevelNeutral, FactionId: proto.Faction_Unknown},
		276929: {RepFactionId: 2758, RepLevel: proto.RepLevel_RepLevelFriendly, FactionId: proto.Faction_Unknown},
	}
	if len(sources) != len(want) {
		t.Errorf("got rep sources for %d items, want %d", len(sources), len(want))
	}
	for itemID, w := range want {
		got := sources[itemID]
		if len(got) != 1 {
			t.Errorf("item %d: got %d rep sources, want 1", itemID, len(got))
			continue
		}
		if got[0].RepFactionId != w.RepFactionId || got[0].RepLevel != w.RepLevel || got[0].FactionId != w.FactionId {
			t.Errorf("item %d: got faction %d %v side %v, want faction %d %v side %v", itemID,
				got[0].RepFactionId, got[0].RepLevel, got[0].FactionId, w.RepFactionId, w.RepLevel, w.FactionId)
		}
	}
	if names[509] != "The League of Arathor" || names[270] != "Zandalar Tribe" || len(names) != 5 {
		t.Errorf("faction names: got %v", names)
	}
}
