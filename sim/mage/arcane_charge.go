package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Arcane Charges, the buff Arcane Blast stacks (400573): "the damage of all your other spells is
// increased by 10% and the mana cost of Arcane Blast is increased by 175%. Effect stacks up to 4
// times and lasts 8 sec or until any other damage spell is cast."
var arcaneChargesRank = spellData.ArcaneBlastTriggered.Highest()

// The damage bonus per charge. The row states it twice, SPELLMOD_DAMAGE (effect 1) and
// SPELLMOD_DOT (effect 3), both 10%; the sim applies the one value to hits and DoTs alike.
var arcaneChargeDamagePerStack = arcaneChargesRank.Effect(dbcenums.A_ADD_PCT_MODIFIER, 0).Percent()

const (
	// Damage that lands after the cast that spent the charges. It carries the bonus that cast
	// locked in (arcaneChargesLock) and never the live one, so charges gained meanwhile do not
	// raise it.
	arcaneChargesLocked = MageSpellFireballDot | MageSpellPyroblastDot | MageSpellFlamestrikeDot |
		MageSpellArcaneMissilesTick | MageSpellBlizzard

	// The casts that spend the charges: any damage spell but Arcane Blast. A missile of the
	// channel is not a cast of its own.
	arcaneChargesConsumers = (MageSpellsAllDamaging | MageSpellArcaneMissilesCast) &^
		(MageSpellArcaneBlast | MageSpellArcaneMissilesTick)
)

func (mage *Mage) registerArcaneCharges() {
	// The charges raise the damage of every damage spell except Arcane Blast itself.
	// Verified in game on Arcane Missiles, Blizzard and Flamestrike, which the row's
	// SPELLMOD_DAMAGE and SPELLMOD_DOT masks (effects 1 and 3) leave out, so this mod takes the
	// sim's own mask instead of ParseEffects'.
	damageMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask: MageSpellsAllDamaging &^ (MageSpellArcaneBlast | arcaneChargesLocked),
		Kind:      core.SpellMod_DamageDone_Pct,
	})

	config := spelldata.AuraConfig(arcaneChargesRank)

	config.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		damageMod.Activate()
	}

	config.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		damageMod.Deactivate()
	}

	config.OnStacksChange = func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
		damageMod.UpdateFloatValue(arcaneChargeDamagePerStack * float64(newStacks))
	}

	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.Matches(arcaneChargesConsumers) {
			aura.Deactivate(sim)
		}
	}

	mage.ArcaneChargesAura = core.BlockPrepull(mage.RegisterAura(config))

	// Arcane Blast's cost, per stack, on the row's own mask.
	spelldata.ParseEffects(&mage.Character, mage.ArcaneChargesAura, arcaneChargesRank, spelldata.Effects(2))
}

// The charges' damage bonus as it stands. A spell whose damage lands later reads it in
// ApplyEffects, before OnCastComplete spends the charges, and hands it to an arcaneChargesLock.
func (mage *Mage) arcaneChargesBonus() float64 {
	return arcaneChargeDamagePerStack * float64(mage.ArcaneChargesAura.GetStacks())
}

// The Arcane Charges bonus a cast locked in, per target, for the damage it deals after the charges
// are gone: a DoT's ticks, a channel's missiles. It holds until the next lock on that target.
type arcaneChargesLock map[*core.Unit]float64

func (lock arcaneChargesLock) set(target *core.Unit, bonus float64) {
	lock[target] = bonus
}

// deal runs a damage calculation of spell on target with the locked bonus applied.
func (lock arcaneChargesLock) deal(spell *core.Spell, target *core.Unit, calc func()) {
	multiplier := 1 + lock[target]
	spell.DamageMultiplier *= multiplier
	calc()
	spell.DamageMultiplier /= multiplier
}

// onTick applies the locked bonus to each of a DoT's ticks.
func (lock arcaneChargesLock) onTick(tick core.OnTick) core.OnTick {
	return func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		lock.deal(dot.Spell, target, func() { tick(sim, target, dot) })
	}
}
