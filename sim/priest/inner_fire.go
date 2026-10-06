package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerInnerFire(rank *spelldata.Spell) {
	auraConfig := spelldata.AuraConfig(rank, spelldata.Label(rank.Name+" "+rank.Rank))
	auraConfig.Tag = rank.Name
	if rank == spellData.InnerFire.Highest() && priest.SelfBuffs.Armor == proto.PriestOptions_InnerFire {
		auraConfig.OnReset = func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
			aura.SetStacks(sim, aura.MaxStacks)
		}
	}
	aura := priest.RegisterAura(auraConfig)

	// Improved Inner Fire states its armor bonus as SPELLMOD_ALL_EFFECTS, which the talent's own parse
	// files as damage done, so the armor is priced with it here. Its extra charges reach the aura as a
	// spell mod through RelatedSelfBuff.
	improvedInnerFire := spellData.ImprovedInnerFire.Rank(priest.Talents.ImprovedInnerFire).EffectN(1)
	spelldata.ParseEffects(&priest.Character, aura, rank, spelldata.ScaledBy(improvedInnerFire))

	aura.AttachProcTrigger(spelldata.ProcTrigger(&priest.Character, rank,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.RemoveStack(sim)
		}))

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = PriestSpellInnerFire
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		activateRank(sim, aura)
		aura.SetStacks(sim, aura.MaxStacks)
	}
	config.RelatedSelfBuff = aura

	priest.RegisterSpell(config)
}

// Casting one rank of a buff replaces whichever other rank of it is up. The ranks share the aura's Tag.
func activateRank(sim *core.Simulation, aura *core.Aura) {
	for _, other := range aura.Unit.GetAurasWithTag(aura.Tag) {
		if other != aura {
			other.Deactivate(sim)
		}
	}
	aura.Activate(sim)
}
