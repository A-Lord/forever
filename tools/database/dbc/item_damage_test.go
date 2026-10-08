package dbc

import (
	"os"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

func TestUnroundedDPS(t *testing.T) {
	once.Do(func() {})
	previous := dbcInstance
	dbcInstance = NewDBC()
	t.Cleanup(func() { dbcInstance = previous })
	for name, value := range map[string]float64{
		"ItemDamageOneHand": 14.695897, "ItemDamageTwoHand": 19.808466,
		"ItemDamageOneHandCaster": 9.63904, "ItemDamageTwoHandCaster": 11.927145,
	} {
		dbcInstance.ItemDamageTable[name] = map[int]ItemDamageTable{
			22: {Quality: []float64{0, 0, 0, value, 30.25, 0, 0}},
			23: {Quality: make([]float64, 7)},
		}
	}
	cases := []struct {
		name                    string
		inventory, class, level int
		quality                 ItemQuality
		flags                   ItemStaticFlags1
		dps, low, high          float64
	}{
		{"one hand", INVTYPE_WEAPON, ITEM_CLASS_WEAPON, 22, 3, 0, 14.695897, 23, 35},
		{"two hand", INVTYPE_2HWEAPON, ITEM_CLASS_WEAPON, 22, 3, 0, 19.808466, 31, 48},
		{"caster one hand", INVTYPE_WEAPON, ITEM_CLASS_WEAPON, 22, 3, CASTER_WEAPON, 9.63904, 15, 23},
		{"caster two hand", INVTYPE_2HWEAPON, ITEM_CLASS_WEAPON, 22, 3, CASTER_WEAPON, 11.927145, 19, 29},
		{"heirloom", INVTYPE_WEAPON, ITEM_CLASS_WEAPON, 22, 7, 0, 30.25, 48, 73},
		{"zero", INVTYPE_WEAPON, ITEM_CLASS_WEAPON, 23, 3, 0, 0, 0, 0},
		{"nonweapon", 0, 0, 22, 0, 0, 0, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			item := Item{Id: 1, ItemClass: c.class, InventoryType: c.inventory, ItemDelay: 2000,
				ItemLevel: c.level, OverallQuality: c.quality, Flags1: c.flags, DmgVariance: 0.4}
			generated := item.ToUIItem().ScalingOptions[0]
			field := generated.ProtoReflect().Descriptor().Fields().ByName("unrounded_dps")
			if field == nil {
				t.Fatal("scaling properties discard unrounded weapon table DPS")
			}
			if got := generated.ProtoReflect().Has(field); got != (c.class == ITEM_CLASS_WEAPON) {
				t.Fatalf("DPS presence = %v for item class %d", got, c.class)
			}
			if got := generated.ProtoReflect().Get(field).Float(); got != c.dps {
				t.Fatalf("unrounded DPS = %.9f, want %.9f", got, c.dps)
			}
			if generated.WeaponDamageMin != c.low || generated.WeaponDamageMax != c.high {
				t.Fatalf("equipped endpoints = %v-%v, want %v-%v", generated.WeaponDamageMin, generated.WeaponDamageMax, c.low, c.high)
			}
			for _, format := range []string{"json", "binary"} {
				t.Run(format, func(t *testing.T) {
					decoded := &proto.ScalingItemProperties{}
					var data []byte
					var err error
					if format == "json" {
						data, err = protojson.Marshal(generated)
					} else {
						data, err = googleProto.Marshal(generated)
					}
					if err != nil {
						t.Fatal(err)
					}
					if format == "json" {
						err = protojson.Unmarshal(data, decoded)
					} else {
						err = googleProto.Unmarshal(data, decoded)
					}
					if err != nil {
						t.Fatal(err)
					}
					if !googleProto.Equal(generated, decoded) {
						t.Fatalf("%s round trip changes scaling properties", format)
					}
				})
			}
		})
	}
}

func TestWeaponDPSPresenceWhenSourceTableIsMissing(t *testing.T) {
	// Given
	once.Do(func() {})
	previous := dbcInstance
	dbcInstance = NewDBC()
	t.Cleanup(func() { dbcInstance = previous })
	item := Item{ItemClass: ITEM_CLASS_WEAPON, InventoryType: INVTYPE_WEAPON, OverallQuality: 2, ItemLevel: 18}
	// When
	row := item.ToUIItem().ScalingOptions[0]
	// Then
	if row.UnroundedDps != nil {
		t.Fatal("missing source row must preserve absent table DPS")
	}
}

func TestWeaponDamageRange(t *testing.T) {
	t.Chdir("../../..")
	if _, err := os.Stat("./assets/db_inputs/dbc/spells.json"); err != nil {
		t.Skip("no game files DBC inputs in assets/db_inputs/dbc - run `make db` from a local WoW install to enable this test")
	}
	previous := dbcInstance
	t.Cleanup(func() { dbcInstance = previous })
	if err := InitDBC(); err != nil {
		t.Fatal(err)
	}
	items := dbcInstance.Items

	for _, tc := range []struct {
		name     string
		id       int
		min, max float64
	}{
		{"Death Speaker Scepter, one-hand caster", 2816, 28, 52},
		{"Simone's Cultivating Hammer, one-hand caster", 22380, 35, 67},
		{"Benediction, two-hand caster with a QualityModifier", 18608, 89, 155},
		{"Bearded Boneaxe, not a caster weapon", 2878, 25, 47},
	} {
		item, ok := items[tc.id]
		if !ok {
			t.Errorf("%s: item %d is not in the DBC inputs", tc.name, tc.id)
			continue
		}
		if got, want := [2]float64{item.WeaponDmgMin(0), item.WeaponDmgMax(0)}, [2]float64{tc.min, tc.max}; got != want {
			t.Errorf("%s: damage %v-%v, want %v-%v", tc.name, got[0], got[1], want[0], want[1])
		}
	}
}
