package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: Not registered, and NewShadowfiend is not called. 401977 is a Season of Discovery rune: only
// Engrave Pants - Shadowfiend (402861) grants it, with no trainer level and no Forever talent. Wire it
// and the pet in if Forever adds the ability.
func (priest *Priest) registerShadowfiendSpell() {
	rank := spellData.Shadowfiend.Highest()

	// The pet's lifetime on the timeline. The pet leaves when this duration runs out, so a mod on the
	// buff's duration keeps it longer.
	priest.ShadowfiendAura = priest.RegisterAura(spelldata.AuraConfig(rank))

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = PriestSpellShadowFiend
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		priest.ShadowfiendPet.EnableWithTimeout(sim, priest.ShadowfiendPet, spell.RelatedSelfBuff.Duration)
		spell.RelatedSelfBuff.Activate(sim)
	}
	config.RelatedSelfBuff = priest.ShadowfiendAura

	priest.Shadowfiend = priest.RegisterSpell(config)
}
