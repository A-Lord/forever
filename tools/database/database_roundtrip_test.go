package database

import "testing"

func TestDatabaseReadRetainsEncounters(t *testing.T) {
	// Given
	input := `{"encounters":[{"path":"Encounter"}]}`
	// When
	db := ReadDatabaseFromJson(input)
	// Then
	if len(db.Encounters) != 1 || db.Encounters[0].Path != "Encounter" {
		t.Fatal("database read discards encounter data")
	}
}
