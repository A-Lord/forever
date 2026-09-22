package hunter

import (
	"github.com/wowsims/forever/sim/core"
)

// TODO: To be implemented. spellData.RaptorStrike holds the eight trainer ranks, 2973 to 14266.
// The client also carries two Season of Discovery ladders under the same name: 415335 to
// 415343, which the rune passive Melee Specialist (415352) swaps onto the action bar, and
// 409691 to 409755, a mana-less copy nothing references. The generator drops the first by
// its override link and the second because only the trainer rank carries a SpellPower row.
func (hunter *Hunter) registerRaptorStrikeSpell() {
	panic("To be implemented")

	// hunter.RaptorStrike = hunter.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: raptorStrikeRank.ID},
	// 	SpellSchool:    raptorStrikeRank.SpellSchool(),
	// 	DefenseType:    raptorStrikeRank.DefenseTypeCore(),
	// 	ClassSpellMask: HunterSpellRaptorStrike,
	// 	ProcMask:       core.ProcMaskMeleeMH,
	// 	Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,
	//
	// 	MaxRange: core.MaxMeleeRange,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(raptorStrikeRank.Cost()),
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			NonEmpty: true,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    hunter.NewTimer(),
	// 			Duration: max(raptorStrikeRank.Cooldown(), raptorStrikeRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		// Emit an "auto delayed" log line whenever the mh auto fired
	// 		// later than it would have in an uncontested rotation. Below 1ms
	// 		// is treated as rounding noise so the common case stays silent.
	// 		delay := hunter.AutoAttacks.MainHandPendingSwingDelay()
	// 		readyAt := sim.CurrentTime - delay
	// 		if sim.Log != nil && delay > time.Millisecond && readyAt > 0 {
	// 			hunter.Log(sim, "%s delayed by %s, was ready at %s", spell.ActionID, delay, readyAt)
	// 		}
	//
	// 		baseDamage := hunter.MHWeaponDamage(sim, spell.MeleeAttackPower(target)) + raptorStrikeRank.DamageEffect().Average(core.CharacterLevel)
	// 		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
	// 	},
	// })
	//
	// hunter.RegisterAura(core.Aura{
	// 	Label:    "Raptor Strike",
	// 	ActionID: core.ActionID{SpellID: raptorStrikeRank.ID}.WithTag(2),
	// 	Icd:      &hunter.RaptorStrike.CD,
	// })
}

// Returns true if the regular melee swing should be used, false otherwise.
func (hunter *Hunter) TryRaptorStrike(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if mhSwingSpell.ActionID.Tag != 1 || !hunter.RaptorStrike.CanCast(sim, hunter.CurrentTarget) {
		return mhSwingSpell
	}

	return hunter.RaptorStrike
}
