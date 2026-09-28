package spelldata

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/stats"
)

func weaponDamageDoneRow(misc int32, subclass int32) *Spell {
	return &Spell{ID: 1500, Name: "Weapon Specialization", DurationMs: -1,
		EquipClass: dbcenums.ITEM_CLASS_WEAPON, EquipSubclass: subclass,
		Effects: []Effect{{SpellID: 1500, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_DAMAGE_PERCENT_DONE,
			Misc: misc, BasePoints: 10, Target: [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_CASTER}}}}
}

func defenseSpell(character *core.Character, id int32, school core.SpellSchool, defense core.DefenseType) *core.Spell {
	return character.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: id},
		SpellSchool:      school,
		DefenseType:      defense,
		ProcMask:         core.ProcMaskMeleeMHSpecial,
		DamageMultiplier: 1,
	})
}

// One-Handed Weapon Specialization 20196 names the Physical school and a melee weapon: its bonus also
// reaches the melee-defense spells of the other schools, and a magic-defense spell of those schools
// takes none of it. Physical keeps the school multiplier alone.
func TestPhysicalMeleeWeaponRowReachesMeleeDefenseSpells(t *testing.T) {
	character := parseWarrior()
	holyMelee := defenseSpell(character, 1, core.SpellSchoolHoly, core.DefenseTypeMelee)
	holyMagic := defenseSpell(character, 2, core.SpellSchoolHoly, core.DefenseTypeMagic)
	physical := defenseSpell(character, 3, core.SpellSchoolPhysical, core.DefenseTypeMelee)

	ParseStatic(character, weaponDamageDoneRow(1, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H))

	if got := holyMelee.DamageMultiplier; math.Abs(got-1.1) > 1e-9 {
		t.Errorf("a melee-defense holy spell: %v, want 1.1", got)
	}
	if got := holyMagic.DamageMultiplier; got != 1 {
		t.Errorf("a magic-defense holy spell: %v, want 1", got)
	}
	if got := physical.DamageMultiplier; got != 1 {
		t.Errorf("a physical spell's own multiplier: %v, want 1, the school multiplier carries it", got)
	}
	if got := character.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]; math.Abs(got-1.1) > 1e-9 {
		t.Errorf("the physical school multiplier: %v, want 1.1", got)
	}
}

// A row that names every school, a ranged weapon or no weapon at all adds nothing past its school
// multiplier.
func TestOtherDamageDoneRowsStayOnTheirSchools(t *testing.T) {
	noWeapon := weaponDamageDoneRow(1, 0)
	noWeapon.EquipClass = 0

	for name, row := range map[string]*Spell{
		"every school": weaponDamageDoneRow(127, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H),
		"ranged":       weaponDamageDoneRow(1, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_BOW),
		"no weapon":    noWeapon,
	} {
		character := parseWarrior()
		holyMelee := defenseSpell(character, 1, core.SpellSchoolHoly, core.DefenseTypeMelee)

		ParseStatic(character, row)

		if got := holyMelee.DamageMultiplier; got != 1 {
			t.Errorf("%s: a melee-defense holy spell's own multiplier: %v, want 1", name, got)
		}
	}
}

// One-Handed Weapon Specialization 20196 at 3/3 gives 10% and Two-Handed Weapon Specialization 20111
// at 3/3 gives 6%, on Physical and on the melee-defense Holy spells such as the Seal of Righteousness
// hit, and nothing on a magic-defense Holy spell such as Consecration.
func TestWeaponSpecializationsReachMeleeDefenseHolySpells(t *testing.T) {
	defer generatedStore()()

	for _, tc := range []struct {
		row  *Spell
		want float64
	}{
		{Talent(20196, 3).Rank(3), 1.10},
		{Talent(20111, 3).Rank(3), 1.06},
	} {
		character := parseWarrior()
		holyMelee := defenseSpell(character, 1, core.SpellSchoolHoly, core.DefenseTypeMelee)
		holyMagic := defenseSpell(character, 2, core.SpellSchoolHoly, core.DefenseTypeMagic)

		ParseStatic(character, tc.row)

		if got := holyMelee.DamageMultiplier; math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("%s: a melee-defense holy spell: %v, want %v", tc.row.Name, got, tc.want)
		}
		if got := holyMagic.DamageMultiplier; got != 1 {
			t.Errorf("%s: a magic-defense holy spell: %v, want 1", tc.row.Name, got)
		}
		if got := character.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]; math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("%s: the physical school multiplier: %v, want %v", tc.row.Name, got, tc.want)
		}
	}
}
