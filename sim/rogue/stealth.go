package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var stealthRank = spellData.Stealth.Highest()

func (rogue *Rogue) registerStealthAura() {
	config := spelldata.SpellConfig(&rogue.Unit, stealthRank)
	rogue.StealthAura = rogue.RegisterAura(spelldata.AuraConfig(stealthRank))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.StealthAura.Activate(sim)
	}

	// The TBC implementation, kept for the port:
	// rogue.StealthAura = rogue.RegisterAura(core.Aura{
	// 	Label:    "Stealth",
	// 	ActionID: core.ActionID{SpellID: stealthRank.ID},
	// 	Duration: core.NeverExpires,
	// 	OnGain: func(aura *core.Aura, sim *core.Simulation) {
	// 		// Stealth triggered auras
	// 		if rogue.MasterOfSubtletyAura != nil {
	// 			rogue.MasterOfSubtletyAura.Duration = core.NeverExpires
	// 			rogue.MasterOfSubtletyAura.Activate(sim)
	// 		}
	// 	},
	// 	OnExpire: func(aura *core.Aura, sim *core.Simulation) {
	// 		if rogue.MasterOfSubtletyAura != nil {
	// 			rogue.MasterOfSubtletyAura.Deactivate(sim)
	// 			rogue.MasterOfSubtletyAura.Duration = time.Second * 6
	// 			rogue.MasterOfSubtletyAura.Activate(sim)
	// 		}
	// 	},
	// 	// Stealth breaks on damage taken (if not absorbed)
	// 	// This may be desirable later, but not applicable currently
	// })
	//
	// rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: stealthRank.ID},
	// 	SpellSchool:    stealthRank.SpellSchool(),
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: RogueSpellStealth,
	//
	// 	Cast: core.CastConfig{
	// 		CD: core.Cooldown{
	// 			Timer:    rogue.NewTimer(),
	// 			Duration: max(stealthRank.Cooldown(), stealthRank.CategoryCooldown()),
	// 		},
	// 	},
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		return sim.CurrentTime < 0
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		spell.RelatedSelfBuff.Activate(sim)
	// 	},
	// 	RelatedSelfBuff: rogue.StealthAura,
	// })
}
