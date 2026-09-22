package warrior

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: Manual review needed -- spell 25288 states "a high amount of threat" with no number;
// none is modelled until measured in game.
var revengeRank = spellData.Revenge.Highest()

func (warrior *Warrior) registerRevenge() {
	actionID := core.ActionID{SpellID: revengeRank.ID}

	// TODO: In-game test needed
	aura := warrior.RegisterAura(core.Aura{
		Label:    "Revenge",
		Duration: 5 * time.Second,
		ActionID: actionID,
	})

	warrior.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Revenge - Trigger",
		TriggerImmediately: true,
		Outcome:            core.OutcomeBlock | core.OutcomeDodge | core.OutcomeParry,
		Callback:           core.CallbackOnSpellHitTaken,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura.Activate(sim)
		},
	})

	config := spelldata.SpellConfig(&warrior.Unit, revengeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.FlatThreatBonus = 0

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return warrior.StanceMatches(DefensiveStance) && aura.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := revengeRank.DamageEffect().Roll(sim, core.CharacterLevel)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		aura.Deactivate(sim)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	config.RelatedSelfBuff = aura

	warrior.RegisterSpell(config)
}
