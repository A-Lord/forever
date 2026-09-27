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
	mage.registerPyroblastTalent()

	// Tier 4
	// Improved Scorch implemented in scorch.go
	mage.registerImprovedFireWard()
	mage.registerHotStreak()
	mage.registerMasterOfElements()

	// Tier 5
	mage.registerCriticalMass()
	mage.registerBlastWaveTalent()

	// Tier 6
	mage.registerFirePower()

	// Tier 7
	mage.registerCombustionTalent()
}

// registerWakeOfFire implements Wake of Fire, new in Forever.
func (mage *Mage) registerWakeOfFire() {
	if mage.Talents.WakeOfFire == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.WakeOfFire.Rank(mage.Talents.WakeOfFire))
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

// registerIgnite implements Ignite: a Fire spell's crit burns the target for a share of its damage
// (11119, 8% per point) over 4 sec. The DoT (412538) stacks to 1, so it rolls: a new crit adds its
// share to what the current DoT has not ticked yet and restarts it over the full duration.
func (mage *Mage) registerIgnite() {
	if mage.Talents.Ignite == 0 {
		return
	}

	igniteFraction := spellData.Ignite.FractionAt(mage.Talents.Ignite)
	igniteDot := spellData.IgniteTriggered.Highest()

	// The damage is already the crit's, after every modifier and resist, so the ticks take none again.
	config := spelldata.SpellConfig(&mage.Unit, igniteDot, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(),
		spelldata.Flags(core.SpellFlagIgnoreModifiers|core.SpellFlagNoSpellMods|core.SpellFlagIgnoreResists|core.SpellFlagProc))
	config.ClassSpellMask = MageSpellIgnite

	config.Dot = spelldata.DotConfig(igniteDot, igniteDot.EffectN(1))
	// The stacks show the damage per tick, which an APL can read; the row's 1 is the rolling cap.
	config.Dot.Aura.MaxStacks = math.MaxInt32
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, dot.SnapshotBaseDamage, dot.OutcomeTick)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}

	mage.Ignite = mage.RegisterSpell(config)

	trigger := spelldata.ProcTrigger(&mage.Character, spellData.Ignite.Rank(mage.Talents.Ignite),
		func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			target := result.Target
			dot := mage.Ignite.Dot(target)
			totalDamage := dot.OutstandingDmg() + result.Damage*igniteFraction

			mage.Ignite.Cast(sim, target)
			dot.SnapshotBaseDamage = totalDamage / float64(dot.BaseTickCount)
			dot.Aura.SetStacks(sim, int32(dot.SnapshotBaseDamage))
		})
	trigger.Name = "Ignite - Trigger"
	trigger.TriggerImmediately = true
	// "Your critical strikes from Fire damage spells": the school, so Frostfire Bolt counts.
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spell.SpellSchool.Matches(core.SpellSchoolFire)
	}

	mage.MakeProcTriggerAura(trigger)
}

// registerFlameThrowing implements Flame Throwing, new in Forever.
func (mage *Mage) registerFlameThrowing() {
	if mage.Talents.FlameThrowing == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.FlameThrowing.Rank(mage.Talents.FlameThrowing))
}

// registerImpact implements Impact, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
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

// registerPyroblastTalent implements the Forever talent gate for Pyroblast.
func (mage *Mage) registerPyroblastTalent() {
	if !mage.Talents.Pyroblast {
		return
	}

	mage.registerPyroblastSpell()
}

// registerBlastWaveTalent implements the talent gate for Blast Wave (blast_wave.go).
func (mage *Mage) registerBlastWaveTalent() {
	if !mage.Talents.BlastWave {
		return
	}

	mage.registerBlastWaveSpell()
}

// registerCombustionTalent implements the talent gate for Combustion (combustion.go).
func (mage *Mage) registerCombustionTalent() {
	if !mage.Talents.Combustion {
		return
	}

	mage.registerCombustionSpell()
}

// registerImprovedFireWard implements Improved Fire Ward, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerImprovedFireWard() {
	if mage.Talents.ImprovedFireWard == 0 {
		return
	}
}

// registerHotStreak implements Hot Streak, new in Forever.
func (mage *Mage) registerHotStreak() {
	if !mage.Talents.HotStreak {
		return
	}

	hotStreakBuff := spellData.HotStreakTriggered.Highest()

	config := spelldata.AuraConfig(hotStreakBuff)

	// UNVERIFIED: whether a Pyroblast consumes every stack or only one has not been checked in game.
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.Matches(MageSpellPyroblast) {
			aura.Deactivate(sim)
		}
	}

	hotStreakAura := mage.RegisterAura(config)

	spelldata.ParseEffects(&mage.Character, hotStreakAura, hotStreakBuff, spelldata.Effects(1))

	trigger := spelldata.ProcTrigger(&mage.Character,
		spellData.HotStreak.Rank(1),
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			hotStreakAura.Activate(sim)
			hotStreakAura.AddStack(sim)
		})
	trigger.Name = "Hot Streak - Trigger"
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spellData.Fireball.Contains(spell) || spellData.FrostfireBolt.Contains(spell) ||
			spellData.FireBlast.Contains(spell) || spellData.Scorch.Contains(spell)
	}

	mage.MakeProcTriggerAura(trigger)
}

// registerMasterOfElements implements Master of Elements.
func (mage *Mage) registerMasterOfElements() {
	if mage.Talents.MasterOfElements == 0 {
		return
	}

	refundFraction := spellData.MasterOfElements.FractionAt(mage.Talents.MasterOfElements)
	manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: 29077})

	trigger := spelldata.ProcTrigger(&mage.Character,
		spellData.MasterOfElements.Rank(mage.Talents.MasterOfElements),
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			mage.AddMana(sim, float64(spell.Cost.BaseCost)*refundFraction, manaMetrics)
		})
	trigger.Name = "Master of Elements - Trigger"

	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		if !spell.SpellSchool.Matches(core.SpellSchoolFire | core.SpellSchoolFrost) {
			return false
		}
		if spell.Cost == nil || spell.Cost.BaseCost == 0 {
			return false
		}

		return spell.CurCast.Cost > 0
	}

	mage.MakeProcTriggerAura(trigger)
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
