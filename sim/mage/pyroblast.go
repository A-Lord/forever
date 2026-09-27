package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var pyroblastRank = spellData.Pyroblast.Highest()
var pyroblastBaseDamage = pyroblastRank.DamageEffect().Average(core.CharacterLevel)

// Pyroblast
// https://www.wowhead.com/forever/spell=18809
//
// The periodic damage is a separate spell (MageSpellPyroblastDot) so that class-mask mods can
// tell it from the hit. Arcane Charges (400573) is the case that needs it: its SPELLMOD_DAMAGE
// mask covers Pyroblast's family flags (0x400000), but its SPELLMOD_DOT effect names only
// Fireball and Frostfire Bolt, so the charges raise the hit and never the DoT.
func (mage *Mage) registerPyroblastSpell() {
	dotConfig := spelldata.SpellConfig(&mage.Unit, pyroblastRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellPyroblastDot

	tick := pyroblastRank.PeriodicEffect()
	dotConfig.Dot = spelldata.DotConfig(pyroblastRank, tick)
	// The DoT only goes out on a landed hit, so its ticks roll crit (the row states Periodic Can
	// Crit) but not hit.
	dotConfig.Dot.OnTick = spelldata.PeriodicDamageTick(tick, pyroblastRank.TickOutcomeHitRolled)

	dotConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}

	pyroblastDot := mage.RegisterSpell(dotConfig)

	config := spelldata.SpellConfig(&mage.Unit, pyroblastRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellPyroblast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, pyroblastBaseDamage, spell.OutcomeMagicHitAndCrit)
		spell.WaitTravelTime(sim, func(s *core.Simulation) {
			spell.DealDamage(sim, result)
			if result.Landed() {
				pyroblastDot.Cast(sim, target)
			}
		})
	}

	mage.Pyroblast = mage.RegisterSpell(config)
	mage.Pyroblast.RelatedDotSpell = pyroblastDot
}
