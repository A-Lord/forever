package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Swift Judgement (talent)
// https://www.wowhead.com/forever/spell=1310994
//
// Finishes the remaining cooldown on your Judgement ability and reduces the Mana cost of your next
// Judgement by 100%.
func (paladin *Paladin) registerSwiftJudgement() {
	rank := spellData.SwiftJudgement.Highest()
	actionID := core.ActionID{SpellID: rank.ID}

	var freeJudgement *core.Aura
	freeJudgement = paladin.RegisterAura(core.Aura{
		Label:    "Swift Judgement" + paladin.Label,
		ActionID: actionID,
		Duration: core.NeverExpires,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct_Add,
		ClassMask:  SpellMaskJudgement,
		FloatValue: rank.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_COST)).Percent(),
	}).AttachProcTrigger(core.ProcTrigger{
		Callback:           core.CallbackOnCastComplete,
		ClassSpellMask:     SpellMaskJudgement,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			freeJudgement.Deactivate(sim)
		},
	})

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = SpellMaskSwiftJudgement
	// Free and off the global cooldown, so the row leaves the cast empty.
	config.Cast.DefaultCast.NonEmpty = true
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		paladin.Judgement.CD.Reset()
		spell.RelatedSelfBuff.Activate(sim)
	}
	config.RelatedSelfBuff = freeJudgement
	paladin.RegisterSpell(config)
}
