package database

import (
	"os"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestWeaponMetadataLoadedFromEmbeddedAssets(t *testing.T) {
	// Given
	db := Load()
	expected := map[int32]float64{273456: 14.695897, 7230: 19.808466, 5279: 10.123344, 7005: 1.6533102, 15444: 11.928854051282052, 2028: 9.63904, 6256: 1.0142055, 866: 31.362663}
	// When
	byID := make(map[int32]*proto.UIItem)
	for _, item := range db.Items {
		byID[item.Id] = item
	}
	// Then
	for id, want := range expected {
		item := byID[id]
		if item == nil {
			t.Errorf("item %d missing", id)
			continue
		}
		row := item.ScalingOptions[0]
		if row == nil || row.UnroundedDps == nil {
			t.Errorf("item %d has no unrounded source DPS", id)
			continue
		}
		if *row.UnroundedDps != want {
			t.Errorf("item %d DPS = %.9f, want %.9f", id, *row.UnroundedDps, want)
		}
	}
}

func TestWeaponMetadataEmbeddedBinaryMatchesJSON(t *testing.T) {
	// Given
	expected := &proto.UIDatabase{}
	for _, name := range []string{"db", "leftover_db"} {
		data, err := os.ReadFile(name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		db := &proto.UIDatabase{}
		if err := protojson.Unmarshal(data, db); err != nil {
			t.Fatal(err)
		}
		googleProto.Merge(expected, db)
	}
	// When
	actual := Load()
	// Then
	if !googleProto.Equal(expected, actual) {
		t.Fatal("embedded binary differs from generated JSON")
	}
}
