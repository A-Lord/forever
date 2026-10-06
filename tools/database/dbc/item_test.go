package dbc

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// ItemSparse.DamageType numbers the schools 0 physical, 1 holy, 2 fire, 3 nature, 4 frost, 5 shadow,
// 6 arcane, an order the proto enum does not share.
func TestDamageTypeMapsToSpellSchool(t *testing.T) {
	for damageType, want := range map[int]proto.SpellSchool{
		0: proto.SpellSchool_SpellSchoolPhysical,
		1: proto.SpellSchool_SpellSchoolHoly,
		2: proto.SpellSchool_SpellSchoolFire,
		3: proto.SpellSchool_SpellSchoolNature,
		4: proto.SpellSchool_SpellSchoolFrost,
		5: proto.SpellSchool_SpellSchoolShadow,
		6: proto.SpellSchool_SpellSchoolArcane,
	} {
		if got := (&Item{DamageType: damageType}).GetSpellSchool(); got != want {
			t.Errorf("DamageType %d = %v, want %v", damageType, got, want)
		}
	}
}
