package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ShadowWordPainRankMap = spellData.ShadowWordPain

// TODO: To be implemented. Shadow Word: Pain already has a full Forever rank ladder
// (spellData.ShadowWordPain); the TBC body needs review before it's uncommented.
func (priest *Priest) registerShadowWordPainSpell(rank *spelldata.Spell) {
	// panic("To be implemented")

	swpTick := rank.PeriodicEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ActionID = core.ActionID{SpellID: rank.ID}
	config.ClassSpellMask = PriestSpellShadowWordPain

	config.Dot = spelldata.DotConfig(rank, swpTick)
	config.Dot.Aura.Tag = "Holy Fire"
	config.Dot.OnTick = spelldata.PeriodicDamageTick(swpTick, rank.TickOutcomeHitRolled)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
	}

	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, swpTick.Average(), spell.OutcomeExpectedMagicAlwaysHit)
	}

	priest.RegisterSpell(config)
}
