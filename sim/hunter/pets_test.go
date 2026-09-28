package hunter

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// The generated family table and the options enum name the same families: a family the client
// adds needs its enum value, and an enum value the client no longer describes has to go.
func TestPetFamiliesMatchTheOptions(t *testing.T) {
	for name := range petFamilies {
		if _, ok := proto.HunterOptions_PetType_value[name]; !ok {
			t.Errorf("pet family %q has no HunterOptions.PetType value", name)
		}
	}
	for name, value := range proto.HunterOptions_PetType_value {
		if value == 0 {
			continue
		}
		if _, ok := petFamilies[name]; !ok {
			t.Errorf("HunterOptions.PetType %s names no generated pet family", name)
		}
	}
	for name, value := range proto.HunterOptions_PetAttackSpeed_value {
		if value == 0 {
			continue
		}
		if _, ok := petAttackSpeeds[proto.HunterOptions_PetAttackSpeed(value)]; !ok {
			t.Errorf("HunterOptions.PetAttackSpeed %s names no passive", name)
		}
	}
}

// The generated arrow table and the options enum name the same arrows, and every value is its
// arrow's item id, which is what the UI lists them by.
func TestArrowsMatchTheOptions(t *testing.T) {
	for name, arrow := range arrows {
		if value, ok := proto.HunterOptions_Ammo_value[name]; !ok {
			t.Errorf("arrow %q has no HunterOptions.Ammo value", name)
		} else if value != arrow.ItemID {
			t.Errorf("HunterOptions.Ammo %s is %d, not the arrow's item id %d", name, value, arrow.ItemID)
		}
	}
	for name, value := range proto.HunterOptions_Ammo_value {
		if value == 0 {
			continue
		}
		if _, ok := arrows[name]; !ok {
			t.Errorf("HunterOptions.Ammo %s names no generated arrow", name)
		}
	}
}

// The same for the quivers.
func TestQuiversMatchTheOptions(t *testing.T) {
	for name, quiver := range quivers {
		if value, ok := proto.HunterOptions_QuiverBonus_value[name]; !ok {
			t.Errorf("quiver %q has no HunterOptions.QuiverBonus value", name)
		} else if value != quiver.ItemID {
			t.Errorf("HunterOptions.QuiverBonus %s is %d, not the quiver's item id %d", name, value, quiver.ItemID)
		}
	}
	for name, value := range proto.HunterOptions_QuiverBonus_value {
		if value == 0 {
			continue
		}
		if _, ok := quivers[name]; !ok {
			t.Errorf("HunterOptions.QuiverBonus %s names no generated quiver", name)
		}
	}
}
