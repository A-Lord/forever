package spelldata

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// Two characters attacking a third, each with its own attack table against it.
func casterScene() (caster, other, target *core.Character) {
	caster, other, target = parseWarrior(), parseWarrior(), parseWarrior()
	target.UnitIndex = 0
	caster.AttackTables = []*core.AttackTable{core.NewAttackTable(&caster.Unit, &target.Unit)}
	other.AttackTables = []*core.AttackTable{core.NewAttackTable(&other.Unit, &target.Unit)}
	return caster, other, target
}

func casterSpell(character *core.Character, id int32, school core.SpellSchool, flags core.ClassFlags) *core.Spell {
	return character.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: id},
		ClassFlags:       flags,
		ProcMask:         core.ProcMaskSpellDamage,
		SpellSchool:      school,
		DamageMultiplier: 1,
	})
}

func takenMultiplier(spell *core.Spell, target *core.Character) float64 {
	return spell.TargetDamageMultiplier(&core.Simulation{}, spell.Unit.AttackTables[target.UnitIndex], false)
}

// Shadow Weaving 15258: 2% more shadow damage from the priest per stack, and nothing from anyone else.
func TestSchoolDamageFromCasterFollowsTheStacks(t *testing.T) {
	sim := &core.Simulation{}
	caster, other, target := casterScene()
	row := &Spell{ID: 15258, Name: "Shadow Weaving", DurationMs: 15000, MaxStack: 5,
		ClassFlags: core.ClassFlags{Family: 6},
		Effects: []Effect{{SpellID: 15258, Type: dbcenums.E_APPLY_AURA,
			Aura: dbcenums.A_MOD_SCHOOL_MASK_DAMAGE_FROM_CASTER, BasePoints: 2, Misc: 32}}}

	shadow := casterSpell(caster, 1, core.SpellSchoolShadow, core.ClassFlags{})
	fire := casterSpell(caster, 2, core.SpellSchoolFire, core.ClassFlags{})
	otherShadow := casterSpell(other, 1, core.SpellSchoolShadow, core.ClassFlags{})

	aura := target.RegisterAura(AuraConfig(row))
	ParseEffects(nil, aura, row, FromCaster(&caster.Unit, 0, 1))

	aura.Activate(sim)
	aura.SetStacks(sim, 3)
	if got := takenMultiplier(shadow, target); math.Abs(got-1.06) > 1e-9 {
		t.Errorf("the caster's shadow spell at three stacks: %v, want 1.06", got)
	}
	if got := takenMultiplier(fire, target); got != 1 {
		t.Errorf("the caster's fire spell: %v, want 1", got)
	}
	if got := takenMultiplier(otherShadow, target); got != 1 {
		t.Errorf("another caster's shadow spell: %v, want 1", got)
	}

	aura.Deactivate(sim)
	if got := takenMultiplier(shadow, target); got != 1 {
		t.Errorf("the caster's shadow spell after the aura dropped: %v, want 1", got)
	}
}

// Wrack 1316697: 10% more damage from the warlock's other shadow dots, which its class mask names,
// and not from Wrack itself.
func TestSpellDamageFromCasterReachesTheNamedSpells(t *testing.T) {
	sim := &core.Simulation{}
	caster, other, target := casterScene()
	row := &Spell{ID: 1316697, Name: "Wrack", DurationMs: 6000,
		ClassFlags: core.ClassFlags{Family: 5, Mask: [4]uint32{1: 262144}},
		Effects: []Effect{{SpellID: 1316697, Index: 1, Type: dbcenums.E_APPLY_AURA,
			Aura: dbcenums.A_MOD_SPELL_DAMAGE_FROM_CASTER, BasePoints: 10,
			ClassFlags: core.ClassFlags{Family: 5, Mask: [4]uint32{0: 1026}}}}}

	corruption := casterSpell(caster, 1, core.SpellSchoolShadow, core.ClassFlags{Family: 5, Mask: [4]uint32{0: 2}})
	wrack := casterSpell(caster, 2, core.SpellSchoolShadow, row.ClassFlags)
	otherCorruption := casterSpell(other, 1, core.SpellSchoolShadow, core.ClassFlags{Family: 5, Mask: [4]uint32{0: 2}})

	aura := target.RegisterAura(AuraConfig(row))
	ParseEffects(nil, aura, row, FromCaster(&caster.Unit, 0, 1))
	aura.Activate(sim)

	if got := takenMultiplier(corruption, target); math.Abs(got-1.1) > 1e-9 {
		t.Errorf("the caster's Corruption: %v, want 1.1", got)
	}
	if got := takenMultiplier(wrack, target); got != 1 {
		t.Errorf("the caster's Wrack: %v, want 1", got)
	}
	if got := takenMultiplier(otherCorruption, target); got != 1 {
		t.Errorf("another caster's Corruption: %v, want 1", got)
	}
}

// Winter's Chill 12579: 2% more crit per stack on the mage's Frostbolt and Ice Lance, which its
// class mask names, and on nothing else.
func TestCritForCasterReachesTheNamedSpells(t *testing.T) {
	sim := &core.Simulation{}
	caster, other, target := casterScene()
	row := &Spell{ID: 12579, Name: "Winter's Chill", DurationMs: 15000, MaxStack: 5,
		ClassFlags: core.ClassFlags{Family: 3, Mask: [4]uint32{2: 1}},
		Effects: []Effect{{SpellID: 12579, Type: dbcenums.E_APPLY_AURA,
			Aura: dbcenums.A_MOD_CRIT_CHANCE_FOR_CASTER_WITH_ABILITIES, BasePoints: 2, Misc: 16,
			ClassFlags: core.ClassFlags{Family: 3, Mask: [4]uint32{0: 0x20020}}}}}

	frostbolt := casterSpell(caster, 1, core.SpellSchoolFrost, core.ClassFlags{Family: 3, Mask: [4]uint32{0: 0x20}})
	blizzard := casterSpell(caster, 2, core.SpellSchoolFrost, core.ClassFlags{Family: 3, Mask: [4]uint32{0: 0x80}})
	otherFrostbolt := casterSpell(other, 1, core.SpellSchoolFrost, core.ClassFlags{Family: 3, Mask: [4]uint32{0: 0x20}})

	before := map[*core.Spell]float64{}
	for _, spell := range []*core.Spell{frostbolt, blizzard, otherFrostbolt} {
		before[spell] = spell.SpellCritChance(&target.Unit)
	}
	physicalBefore := frostbolt.PhysicalCritChance(caster.AttackTables[0])

	aura := target.RegisterAura(AuraConfig(row))
	ParseEffects(nil, aura, row, FromCaster(&caster.Unit, 0, 1))
	aura.Activate(sim)
	aura.SetStacks(sim, 3)

	if got := frostbolt.SpellCritChance(&target.Unit) - before[frostbolt]; math.Abs(got-0.06) > 1e-9 {
		t.Errorf("the caster's Frostbolt at three stacks gains %v, want 0.06", got)
	}
	if got := frostbolt.PhysicalCritChance(caster.AttackTables[0]) - physicalBefore; math.Abs(got-0.06) > 1e-9 {
		t.Errorf("the caster's Frostbolt's physical crit gains %v, want 0.06", got)
	}
	if got := blizzard.SpellCritChance(&target.Unit) - before[blizzard]; got != 0 {
		t.Errorf("the caster's Blizzard gains %v, want 0", got)
	}
	if got := otherFrostbolt.SpellCritChance(&target.Unit) - before[otherFrostbolt]; got != 0 {
		t.Errorf("another caster's Frostbolt gains %v, want 0", got)
	}

	aura.Deactivate(sim)
	if got := frostbolt.SpellCritChance(&target.Unit); got != before[frostbolt] {
		t.Errorf("the caster's Frostbolt after the aura dropped: %v, want %v", got, before[frostbolt])
	}
}

// A row that states a damage effect and a crit effect holds one slot, and each handler answers for
// its own effects.
func TestFromCasterEffectsShareTheParsesSlot(t *testing.T) {
	sim := &core.Simulation{}
	caster, _, target := casterScene()
	flags := core.ClassFlags{Family: 3, Mask: [4]uint32{0: 0x20}}
	row := &Spell{ID: 7000, Name: "Two Effects", DurationMs: 10000, ClassFlags: core.ClassFlags{Family: 3},
		Effects: []Effect{
			{SpellID: 7000, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_SCHOOL_MASK_DAMAGE_FROM_CASTER,
				BasePoints: 10, Misc: 16},
			{SpellID: 7000, Index: 1, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_SPELL_DAMAGE_FROM_CASTER,
				BasePoints: 20, ClassFlags: flags},
			{SpellID: 7000, Index: 2, Type: dbcenums.E_APPLY_AURA,
				Aura: dbcenums.A_MOD_CRIT_CHANCE_FOR_CASTER_WITH_ABILITIES, BasePoints: 5, ClassFlags: flags},
		}}

	frostbolt := casterSpell(caster, 1, core.SpellSchoolFrost, flags)
	aura := target.RegisterAura(AuraConfig(row))
	ParseEffects(nil, aura, row, FromCaster(&caster.Unit, 1, 2))
	aura.Activate(sim)

	table := caster.AttackTables[0]
	if len(table.DamageDoneByCasterExtraMultiplier) != 2 || table.DamageDoneByCasterExtraMultiplier[1] == nil ||
		len(table.CritChanceByCasterBonus) != 2 || table.CritChanceByCasterBonus[1] == nil {
		t.Fatalf("the handlers are not in slot 1 of 2")
	}
	if got := takenMultiplier(frostbolt, target); math.Abs(got-1.1*1.2) > 1e-9 {
		t.Errorf("the caster's Frostbolt: %v, want 1.1 * 1.2", got)
	}

	aura.Deactivate(sim)
	if table.DamageDoneByCasterExtraMultiplier[1] != nil || table.CritChanceByCasterBonus[1] != nil {
		t.Errorf("the handlers stay in their slot after the aura dropped")
	}
}

// Without a caster the rows have no attacks to reach: a raid debuff whose caster the sim does not
// model, and the dry run a generator reads.
func TestFromCasterRowsNeedACaster(t *testing.T) {
	row := &Spell{ID: 22959, Name: "Fire Vulnerability", DurationMs: 30000, MaxStack: 5,
		Effects: []Effect{{SpellID: 22959, Type: dbcenums.E_APPLY_AURA,
			Aura: dbcenums.A_MOD_SCHOOL_MASK_DAMAGE_FROM_CASTER, BasePoints: 3, Misc: 4}}}

	_, _, target := casterScene()
	if parsed := ParseEffects(nil, target.RegisterAura(AuraConfig(row)), row); len(parsed.Applied) != 0 || len(parsed.Skipped) != 1 {
		t.Errorf("without a caster: attached %v and skipped %d, want the effect skipped", appliedKinds(parsed), len(parsed.Skipped))
	}
	if parsed := DryRun(row); len(parsed.Applied) != 0 {
		t.Errorf("the dry run attached %v, want nothing", appliedKinds(parsed))
	}
}
