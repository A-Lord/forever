package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
	// Idol of the Moon
	core.NewItemEffect(23197, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Improved Moonfire",
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask:  DruidSpellMoonfire,
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 33.0,
		}))

		// TODO: this registers the swap proc against 32330, not 23197. That item is not in
		// the Forever database either, so the registration is dead whichever was intended.
		character.ItemSwap.RegisterProc(32330, aura)
	})

	// Howling Idol (272427): its equip spell lowers Shifting Power's cooldown.
	core.NewItemEffect(272427, func(agent core.Agent) {
		spelldata.ParseStatic(agent.GetCharacter(), spelldata.MustFind(1291059))
	})

	// Wolfshead Helm (8345). Effect 1 raises effect 2 of Enrage, its instant rage, which the parser
	// would read as damage. Effect 2 is the energy Shifting Power adds; the row states it as a dummy.
	core.NewItemEffect(8345, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		row := spelldata.MustFind(17768)
		energy := row.EffectN(2).BaseValue()
		enrage := row.EffectN(1).Percent()
		core.MakePermanent(druid.RegisterAura(core.Aura{
			Label:    "Wolfshead Helm",
			ActionID: core.ActionID{SpellID: row.ID},
			OnGain: func(_ *core.Aura, _ *core.Simulation) {
				druid.WolfsheadEnergyBonus += energy
				druid.WolfsheadEnrageBonus += enrage
			},
			OnExpire: func(_ *core.Aura, _ *core.Simulation) {
				druid.WolfsheadEnergyBonus -= energy
				druid.WolfsheadEnrageBonus -= enrage
			},
		}))
	})
}
