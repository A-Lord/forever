package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shieldBlockRank = spellData.ShieldBlock.Highest()

func (warrior *Warrior) registerShieldBlock() {
	aura := warrior.RegisterAura(spelldata.AuraConfig(shieldBlockRank))
	spelldata.ParseEffects(&warrior.Character, aura, shieldBlockRank)

	// The block that spends a charge is an outcome no proc mask states, so the listener is the
	// caller's; the row states the two charges the aura starts with.
	aura.AttachProcTrigger(core.ProcTrigger{
		Name:               "Shield Block - Consume",
		TriggerImmediately: true,
		Outcome:            core.OutcomeBlock,
		Callback:           core.CallbackOnSpellHitTaken,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.RemoveStack(sim)
		},
	})

	config := spelldata.SpellConfig(&warrior.Unit, shieldBlockRank, spelldata.Flags(core.SpellFlagAPL))

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return warrior.PseudoStats.CanBlock
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		aura.Activate(sim)
		aura.SetStacks(sim, aura.MaxStacks)
	}

	config.RelatedSelfBuff = aura

	warrior.RegisterSpell(config)

	warrior.deactivateWithoutShield(aura)
}

func (warrior *Warrior) deactivateWithoutShield(aura *core.Aura) {
	warrior.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand}, func(sim *core.Simulation, _ proto.ItemSlot) {
		if !warrior.PseudoStats.CanBlock {
			aura.Deactivate(sim)
		}
	})
}
