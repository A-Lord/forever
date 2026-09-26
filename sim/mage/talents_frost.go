package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (mage *Mage) registerFrostTalents() {
	// Tier 1
	mage.registerFrostWarding()
	mage.registerImprovedFrostbolt()
	mage.registerElementalPrecision()

	// Tier 2
	mage.registerIceShards()
	mage.registerPermafrost()
	mage.registerImprovedFrostNova()
	mage.registerFrostbite()

	// Tier 3
	mage.registerPiercingIce()
	mage.registerFrostChanneling()
	// Ice Lance implemented in ice_lance.go, gated on the talent there.
	mage.registerImprovedBlizzard()

	// Tier 4
	mage.registerArcticReach()
	// Ice Block implemented in defensives.go, gated on the talent there.
	// Shatter attaches to the Fingers of Frost aura below.

	// Tier 5
	mage.registerImprovedConeOfCold()
	// Cold Snap implemented in cold_snap.go; registered unconditionally
	// from registerSpells (its own Talents.ColdSnap guard is inside that file).
	mage.registerFingersOfFrost()

	// Tier 6
	mage.registerWinterChill()

	// Tier 7
	// Ice Barrier implemented in defensives.go, gated on the talent there.
}

func (mage *Mage) registerFrostWarding() {
	if mage.Talents.FrostWarding == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.FrostWarding.Rank(mage.Talents.FrostWarding))
}

func (mage *Mage) registerImprovedFrostbolt() {
	if mage.Talents.ImprovedFrostbolt == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedFrostbolt.Rank(mage.Talents.ImprovedFrostbolt))
}

func (mage *Mage) registerElementalPrecision() {
	if mage.Talents.ElementalPrecision == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ElementalPrecision.Rank(mage.Talents.ElementalPrecision))
}

func (mage *Mage) registerIceShards() {
	if mage.Talents.IceShards == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.IceShards.Rank(mage.Talents.IceShards))
}

func (mage *Mage) registerPermafrost() {
	if mage.Talents.Permafrost == 0 {
		return
	}

	// Chill duration and slow amount; movement is not simulated but the mods are the row's own.
	spelldata.ParseStatic(&mage.Character, spellData.Permafrost.Rank(mage.Talents.Permafrost))
}

func (mage *Mage) registerImprovedFrostNova() {
	if mage.Talents.ImprovedFrostNova == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedFrostNova.Rank(mage.Talents.ImprovedFrostNova))
}

// Frostbite (11071) roots the target on a chill, and a raid boss is immune to roots. The
// frozen state it would enable is carried by Fingers of Frost instead.
func (mage *Mage) registerFrostbite() {
	if mage.Talents.Frostbite == 0 {
		return
	}
}

func (mage *Mage) registerPiercingIce() {
	if mage.Talents.PiercingIce == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.PiercingIce.Rank(mage.Talents.PiercingIce))
}

func (mage *Mage) registerFrostChanneling() {
	if mage.Talents.FrostChanneling == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.FrostChanneling.Rank(mage.Talents.FrostChanneling))
}

// Improved Blizzard (11185) adds a chill to Blizzard's ticks; the slow itself does nothing
// to a raid boss, and its Fingers of Frost synergy is read in registerFingersOfFrost.
func (mage *Mage) registerImprovedBlizzard() {
	if mage.Talents.ImprovedBlizzard == 0 {
		return
	}
}

func (mage *Mage) registerArcticReach() {
	if mage.Talents.ArcticReach == 0 {
		return
	}

	// Range and radius modifiers; the sim ignores both but the mods are the row's own.
	spelldata.ParseStatic(&mage.Character, spellData.ArcticReach.Rank(mage.Talents.ArcticReach))
}

func (mage *Mage) registerImprovedConeOfCold() {
	if mage.Talents.ImprovedConeOfCold == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedConeOfCold.Rank(mage.Talents.ImprovedConeOfCold))
}

var fingersOfFrostBuff = spellData.FingersOfFrostTriggered.Highest()

func (mage *Mage) registerFingersOfFrost() {
	if mage.Talents.FingersOfFrost == 0 {
		return
	}

	fofRank := spellData.FingersOfFrost.Rank(mage.Talents.FingersOfFrost)

	// The charge count is the talent's first effect (1/2); 400669 states the duration.
	config := spelldata.AuraConfig(fingersOfFrostBuff)
	config.MaxStacks = int32(fofRank.EffectN(1).BaseValue())
	mage.FingersOfFrostAura = mage.RegisterAura(config)

	// Shatter's crit against frozen targets rides the same aura: while Fingers of Frost
	// treats the target as frozen, frost spells crit for its ladder more (17/33/50).
	if mage.Talents.Shatter > 0 {
		shatterCritMod := mage.AddDynamicMod(core.SpellModConfig{
			School:     core.SpellSchoolFrost,
			FloatValue: spellData.Shatter.ValueAt(mage.Talents.Shatter),
			Kind:       core.SpellMod_BonusCrit_Percent,
		})
		mage.FingersOfFrostAura.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
			shatterCritMod.Activate()
		})
		mage.FingersOfFrostAura.ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
			shatterCritMod.Deactivate()
		})
	}

	trigger := spelldata.ProcTrigger(&mage.Character, fofRank,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			mage.FingersOfFrostAura.Activate(sim)
			mage.FingersOfFrostAura.SetStacks(sim, mage.FingersOfFrostAura.MaxStacks)
		})
	trigger.Name = "Fingers of Frost - Trigger"
	// The row hears every harmful spell; the tooltip means the chill effects.
	trigger.ClassSpellMask = MageSpellFrostbolt | MageSpellConeOfCold | MageSpellFrostfireBolt |
		core.TernaryInt64(mage.Talents.ImprovedBlizzard > 0, MageSpellBlizzard, 0)
	mage.MakeProcTriggerAura(trigger)
}

var wintersChillDebuff = spellData.WintersChillTriggered.Highest()

// The per-stack crit Winter's Chill grants its caster, a percent on the debuff row (2).
var wintersChillCritPerStack = wintersChillDebuff.EffectN(1).BaseValue()

func (mage *Mage) registerWinterChill() {
	if mage.Talents.WintersChill == 0 {
		return
	}

	// The debuff 12579 is A_MOD_CRIT_CHANCE_FOR_CASTER_WITH_ABILITIES, so it raises this
	// mage's crit alone; the attack-table bonus is the closest core mechanism and covers
	// every school where the row names frost spells only.
	wcAuras := mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		var appliedCrit float64
		aura := target.GetOrRegisterAura(spelldata.AuraConfig(wintersChillDebuff,
			spelldata.Label("Winter's Chill"+mage.Label)))
		aura.ApplyOnStacksChange(func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
			table := mage.AttackTables[target.UnitIndex]
			table.BonusSpellCritPercent -= appliedCrit
			appliedCrit = wintersChillCritPerStack * float64(newStacks)
			table.BonusSpellCritPercent += appliedCrit
		})
		aura.ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
			mage.AttackTables[target.UnitIndex].BonusSpellCritPercent -= appliedCrit
			appliedCrit = 0
		})
		return aura
	})

	mage.Env.RegisterPreFinalizeEffect(func() {
		for _, spell := range mage.GetSpellsMatchingSchool(core.SpellSchoolFrost) {
			spell.RelatedAuraArrays = spell.RelatedAuraArrays.Append(wcAuras)
		}
	})

	// Forever states a flat SpellAuraOptions.ProcChance of 100 on the talent spell and puts
	// the real per-rank chance on effect 2 (20..100).
	procChance := spellData.WintersChill.EffectAt(2).FractionAt(mage.Talents.WintersChill)

	mage.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Winters Chill Talent",
		Callback:       core.CallbackOnSpellHitDealt,
		Outcome:        core.OutcomeLanded,
		ClassSpellMask: MageSpellFrost,
		ProcChance:     procChance,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura := wcAuras.Get(result.Target)
			aura.Activate(sim)
			aura.AddStack(sim)
		},
	})
}
