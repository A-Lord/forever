package spelldata

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func weaponDamageDoneRow(misc int32, subclass int32) *Spell {
	return &Spell{ID: 1500, Name: "Weapon Specialization", DurationMs: -1,
		EquipClass: dbcenums.ITEM_CLASS_WEAPON, EquipSubclass: subclass,
		Effects: []Effect{{SpellID: 1500, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_DAMAGE_PERCENT_DONE,
			Misc: misc, BasePoints: 10, Target: [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_CASTER}}}}
}

func swordWarrior(hand proto.HandType) *core.Character {
	character := parseWarrior()
	character.Equipment[proto.ItemSlot_ItemSlotMainHand] = core.Item{
		ID: 1, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: hand}
	return character
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

func physicalMultiplier(character *core.Character) float64 {
	return character.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]
}

// One-Handed Weapon Specialization 20196 names the Physical school and a melee weapon: its bonus also
// reaches the melee-defense spells of the other schools, and a magic-defense spell of those schools
// takes none of it. Physical keeps the school multiplier alone.
func TestPhysicalMeleeWeaponRowReachesMeleeDefenseSpells(t *testing.T) {
	character := swordWarrior(proto.HandType_HandTypeOneHand)
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
	if got := physicalMultiplier(character); math.Abs(got-1.1) > 1e-9 {
		t.Errorf("the physical school multiplier: %v, want 1.1", got)
	}
}

// A row that names every school or a ranged weapon adds nothing past its school multiplier, and a row
// that names no weapon holds whatever the character wears.
func TestOtherDamageDoneRowsStayOnTheirSchools(t *testing.T) {
	noWeapon := weaponDamageDoneRow(1, 0)
	noWeapon.EquipClass = 0

	for name, row := range map[string]*Spell{
		"every school": weaponDamageDoneRow(127, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H),
		"ranged":       weaponDamageDoneRow(1, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_BOW),
		"no weapon":    noWeapon,
	} {
		character := swordWarrior(proto.HandType_HandTypeOneHand)
		holyMelee := defenseSpell(character, 1, core.SpellSchoolHoly, core.DefenseTypeMelee)

		ParseStatic(character, row)

		if got := holyMelee.DamageMultiplier; got != 1 {
			t.Errorf("%s: a melee-defense holy spell's own multiplier: %v, want 1", name, got)
		}
	}

	character := parseWarrior()
	ParseStatic(character, noWeapon)
	if got := physicalMultiplier(character); math.Abs(got-1.1) > 1e-9 {
		t.Errorf("no weapon, bare hands: the physical school multiplier: %v, want 1.1", got)
	}
}

// A row that names a weapon holds only while a matching weapon is equipped: One-Handed Weapon
// Specialization gives nothing with bare hands or a two-hander.
func TestWeaponDamageDoneRowNeedsItsWeapon(t *testing.T) {
	for name, character := range map[string]*core.Character{
		"bare hands": parseWarrior(),
		"two-hander": swordWarrior(proto.HandType_HandTypeTwoHand),
	} {
		holyMelee := defenseSpell(character, 1, core.SpellSchoolHoly, core.DefenseTypeMelee)

		ParseStatic(character, weaponDamageDoneRow(1, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H))

		if got := physicalMultiplier(character); got != 1 {
			t.Errorf("%s: the physical school multiplier: %v, want 1", name, got)
		}
		if got := holyMelee.DamageMultiplier; got != 1 {
			t.Errorf("%s: a melee-defense holy spell: %v, want 1", name, got)
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
		hand proto.HandType
		want float64
	}{
		{Talent(20196, 3).Rank(3), proto.HandType_HandTypeOneHand, 1.10},
		{Talent(20111, 3).Rank(3), proto.HandType_HandTypeTwoHand, 1.06},
	} {
		character := swordWarrior(tc.hand)
		holyMelee := defenseSpell(character, 1, core.SpellSchoolHoly, core.DefenseTypeMelee)
		holyMagic := defenseSpell(character, 2, core.SpellSchoolHoly, core.DefenseTypeMagic)

		ParseStatic(character, tc.row)

		if got := holyMelee.DamageMultiplier; math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("%s: a melee-defense holy spell: %v, want %v", tc.row.Name, got, tc.want)
		}
		if got := holyMagic.DamageMultiplier; got != 1 {
			t.Errorf("%s: a magic-defense holy spell: %v, want 1", tc.row.Name, got)
		}
		if got := physicalMultiplier(character); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("%s: the physical school multiplier: %v, want %v", tc.row.Name, got, tc.want)
		}
	}
}

// Applied names the melee-defense part beside the school multiplier, as a factor.
func TestWeaponDamageDoneRowReportsItsMeleeDefensePart(t *testing.T) {
	defer generatedStore()()

	applied := DryRun(Talent(20196, 3).Rank(3)).Applied
	if len(applied) != 2 {
		t.Fatalf("applied %+v, want the school multiplier and the melee-defense part", applied)
	}
	for i, want := range []string{"damage-dealt-by-school", "damage-dealt-melee"} {
		if applied[i].Kind != want || math.Abs(applied[i].Value-1.1) > 1e-9 {
			t.Errorf("applied[%d] = %s %v, want %s 1.1", i, applied[i].Kind, applied[i].Value, want)
		}
	}
}
