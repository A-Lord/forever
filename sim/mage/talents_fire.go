package mage

import (
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (mage *Mage) registerFireTalents() {
	// Tier 1
	mage.registerWakeOfFire()
	mage.registerIncineration()
	mage.registerImprovedFireball()

	// Tier 2
	mage.registerIgnite()
	mage.registerFlameThrowing()
	mage.registerImpact()

	// Tier 3
	mage.registerBurningSoul()
	mage.registerImprovedFlamestrike()
	// Pyroblast implemented in pyroblast.go, gated on the talent there.

	// Tier 4
	// Improved Scorch implemented in scorch.go
	mage.registerImprovedFireWard()
	mage.registerHotStreak()
	mage.registerMasterOfElements()

	// Tier 5
	mage.registerCriticalMass()
	// Blast Wave implemented in blast_wave.go; registered unconditionally
	// from registerSpells (its own Talents.BlastWave guard is inside that file).

	// Tier 6
	mage.registerFirePower()

	// Tier 7
	// Combustion implemented in combustion.go; registered unconditionally
	// from registerSpells (its own Talents.Combustion guard is inside that file).
}

// Wake of Fire (11078): the Blast Wave cooldown reduction is static; its other half, a
// +50% Blast Wave crit buff after a killing blow (1312934), never fires against a raid
// boss and is not modelled.
func (mage *Mage) registerWakeOfFire() {
	if mage.Talents.WakeOfFire == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.WakeOfFire.Rank(mage.Talents.WakeOfFire),
		spelldata.Effects(1))
}

func (mage *Mage) registerIncineration() {
	if mage.Talents.Incineration == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.Incineration.Rank(mage.Talents.Incineration))
}

func (mage *Mage) registerImprovedFireball() {
	if mage.Talents.ImprovedFireball == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedFireball.Rank(mage.Talents.ImprovedFireball))
}

var igniteBuff = spellData.IgniteTriggered.Highest()

// Ignite's dot sits on an A_PERIODIC_DUMMY, so it stays hand-written: the ticks carry a
// rolling share of each crit rather than a client-stated amount.
func (mage *Mage) registerIgnite() {
	if mage.Talents.Ignite == 0 {
		return
	}

	igniteDamageMultiplier := spellData.Ignite.FractionAt(mage.Talents.Ignite)
	tickLength := igniteBuff.EffectN(1).Period()

	igniteSpell := mage.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: igniteBuff.ID},
		SpellSchool:      igniteBuff.SpellSchool(),
		ProcMask:         core.ProcMaskSpellDamage,
		ClassSpellMask:   MageSpellIgnite,
		Flags:            core.SpellFlagIgnoreModifiers | core.SpellFlagNoSpellMods | core.SpellFlagNoOnCastComplete | core.SpellFlagIgnoreResists | core.SpellFlagProc,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:     "Ignite",
				Tag:       "IgniteDot",
				MaxStacks: math.MaxInt32,
			},
			NumberOfTicks: int32(igniteBuff.Duration() / tickLength),
			TickLength:    tickLength,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Spell.CalcAndDealPeriodicDamage(sim, target, dot.SnapshotBaseDamage, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).Apply(sim)
		},
	})

	refreshIgnite := func(sim *core.Simulation, target *core.Unit, damagePerTick float64) {
		dot := igniteSpell.Dot(target)
		igniteSpell.Cast(sim, target)
		dot.SnapshotBaseDamage = damagePerTick
		dot.Aura.SetStacks(sim, int32(dot.SnapshotBaseDamage))
	}

	procTrigger := core.ProcTrigger{
		Name:               "Ignite Talent",
		CanProcFromProcs:   true, // 11119, 11120, 12846-12848 carry the bit.
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskSpellDamage,
		ClassSpellMask:     FireSpellIgnitable,
		Outcome:            core.OutcomeCrit,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			target := result.Target
			dot := igniteSpell.Dot(target)
			outstandingDamage := dot.OutstandingDmg()
			if dot.RemainingTicks() <= 0 {
				outstandingDamage = 0
			}

			newDamage := result.Damage * igniteDamageMultiplier
			totalDamage := outstandingDamage + newDamage
			damagePerTick := totalDamage / float64(dot.BaseTickCount)

			refreshIgnite(sim, target, damagePerTick)
		},
	}
	igniteSpell.Unit.MakeProcTriggerAura(procTrigger)
	mage.Ignite = igniteSpell

	// This is needed because we want to listen for the spell "cast" event that refreshes the Dot
	mage.Ignite.Flags ^= core.SpellFlagNoOnCastComplete
}

func (mage *Mage) registerFlameThrowing() {
	if mage.Talents.FlameThrowing == 0 {
		return
	}

	// A range modifier; the sim ignores range but the mod is the row's own.
	spelldata.ParseStatic(&mage.Character, spellData.FlameThrowing.Rank(mage.Talents.FlameThrowing))
}

// Impact (11103) stuns the target for 2 s on a fire spell hit, which does nothing to a
// raid boss.
func (mage *Mage) registerImpact() {
	if mage.Talents.Impact == 0 {
		return
	}
}

func (mage *Mage) registerBurningSoul() {
	if mage.Talents.BurningSoul == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.BurningSoul.Rank(mage.Talents.BurningSoul))
}

func (mage *Mage) registerImprovedFlamestrike() {
	if mage.Talents.ImprovedFlamestrike == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedFlamestrike.Rank(mage.Talents.ImprovedFlamestrike))
}

func (mage *Mage) registerImprovedFireWard() {
	if mage.Talents.ImprovedFireWard == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedFireWard.Rank(mage.Talents.ImprovedFireWard))
}

var hotStreakRank = spellData.HotStreak.Highest()
var hotStreakBuff = spellData.HotStreakTriggered.Highest()

func (mage *Mage) registerHotStreak() {
	if !mage.Talents.HotStreak {
		return
	}

	// 400625 stacks to three, each stack -25% Pyroblast cast time; the modifier is the
	// row's own and follows the stacks.
	mage.HotStreakAura = mage.RegisterAura(spelldata.AuraConfig(hotStreakBuff))
	spelldata.ParseEffects(&mage.Character, mage.HotStreakAura, hotStreakBuff)

	mage.HotStreakAura.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.ClassSpellMask == MageSpellPyroblast {
			aura.Deactivate(sim)
		}
	}

	trigger := spelldata.ProcTrigger(&mage.Character, hotStreakRank,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			mage.HotStreakAura.Activate(sim)
			mage.HotStreakAura.AddStack(sim)
		})
	trigger.Name = "Hot Streak - Trigger"
	// The row hears every harmful spell crit; the tooltip names the direct fire spells.
	trigger.ClassSpellMask = FireSpellIgnitable | MageSpellFireBlast | MageSpellFrostfireBolt
	mage.MakeProcTriggerAura(trigger)
}

func (mage *Mage) registerMasterOfElements() {
	if mage.Talents.MasterOfElements == 0 {
		return
	}

	refundCoeff := spellData.MasterOfElements.FractionAt(mage.Talents.MasterOfElements)
	manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: spellData.MasterOfElements.Highest().ID})

	mage.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Master of Elements",
		Duration:       core.NeverExpires,
		ClassSpellMask: MageSpellFire | MageSpellFrost,
		Outcome:        core.OutcomeCrit,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.CurCast.Cost == 0 {
				return
			}
			mage.AddMana(sim, spell.DefaultCast.Cost*refundCoeff, manaMetrics)
		},
	})
}

func (mage *Mage) registerCriticalMass() {
	if mage.Talents.CriticalMass == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.CriticalMass.Rank(mage.Talents.CriticalMass))
}

func (mage *Mage) registerFirePower() {
	if mage.Talents.FirePower == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.FirePower.Rank(mage.Talents.FirePower))
}

// ------ FROST TALENTS ------
