package shared

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// An item proc whose spell grants the wearer extra attacks: Flurry Axe's 18797 and Thrash Blade's
// 21919 each grant one. Each is a main-hand swing tagged with the granting spell, so the metrics tell
// it from the swing that procced it.
func NewSpellDataExtraAttackProc(cfg SpellDataProc, variants []ItemVariant) {
	forEachSpellDataVariant(cfg, variants, registerSpellDataExtraAttackProc)
}

func registerSpellDataExtraAttackProc(cfg SpellDataProc) {
	registerSpellDataRowProc(cfg, nil, func(agent core.Agent, source effectSource, trigger *spelldata.Spell, row *spelldata.Spell) {
		applySpellDataExtraAttackProc(agent, cfg, source, trigger, row)
	})
}

func applySpellDataExtraAttackProc(agent core.Agent, cfg SpellDataProc, source effectSource, trigger *spelldata.Spell, row *spelldata.Spell) {
	character := agent.GetCharacter()
	count := int(row.ExtraAttacksEffect().BasePoints)

	// The swing it grants is not a hit the proc hears, or one proc would swing forever.
	var extraAttack *core.Spell
	config := spellDataProcListener(character, cfg, source, trigger, nil)
	config.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spell != extraAttack
	}
	config.Handler = func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		for range count {
			character.AutoAttacks.MaybeReplaceMHSwing(sim, extraAttack).Cast(sim, result.Target)
		}
	}
	config.TriggerImmediately = true

	// The main-hand swing is configured after the item effects register, so the copy waits for init.
	source.registerTrigger(character, config).ApplyOnInit(func(_ *core.Aura, _ *core.Simulation) {
		swing := *character.AutoAttacks.MHConfig()
		swing.ActionID = swing.ActionID.WithTag(row.ID)
		extraAttack = character.GetOrRegisterSpell(swing)
	})
}
