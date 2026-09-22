package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var overpowerRank = spellData.Overpower.ByID(11585)
var overpowerBaseDamage = overpowerRank.DamageEffect().Average(core.CharacterLevel)

// The window a dodge opens: the aura Offensive State (DND) fires on the hit.
var overpowerWindow = spellData.OffensiveStateTriggered.Highest()

func (warrior *Warrior) registerOverpower() {
	warrior.OverpowerAura = warrior.RegisterAura(core.Aura{
		ActionID: core.ActionID{SpellID: overpowerWindow.ID},
		Label:    "Overpower Aura",
		Duration: overpowerWindow.Duration(),
	})

	warrior.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Overpower - Trigger",
		TriggerImmediately: true,
		Outcome:            core.OutcomeDodge,
		Callback:           core.CallbackOnSpellHitDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			warrior.OverpowerAura.Activate(sim)
		},
	})

	config := spelldata.SpellConfig(&warrior.Unit, overpowerRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	// TODO: Ingame validation needed
	config.ThreatMultiplier = 1

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return warrior.StanceMatches(BattleStance) && warrior.OverpowerAura.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := overpowerBaseDamage + spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialNoBlockDodgeParry)
		warrior.OverpowerAura.Deactivate(sim)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
