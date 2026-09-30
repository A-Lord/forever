package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The debuff is the shared one the raid option also builds, parsed from the row: -37 physical damage
// done by the target. Its category keeps this warlock's copy and a raid one from stacking.
func (warlock *Warlock) registerCurseOfWeakness() {
	rankedSpell := spellData.CurseOfWeakness.ByID(11708)

	warlock.CurseOfWeaknessAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataCurseOfWeaknessAura(target, true, 0)
	})
	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1
	config.ClassSpellMask = WarlockSpellCurseOfWeakness
	config.RelatedAuraArrays = warlock.CurseOfWeaknessAuras.ToMap()

	if warlock.Talents.AmplifyCurse {
		// Amplify Curse's first effect is SPELLMOD_ALL_EFFECTS +50% naming this curse: -37 becomes -55.
		ampWeakness := spellData.AmplifyCurse.ByID(18288).EffectN(1)

		// The same row, built the way the shared debuff is, with Amplify Curse's +50% parsed onto it.
		warlock.CurseOfWeaknessAmpAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			aura := target.GetOrRegisterAura(core.Aura{
				Label:    "Curse of Weakness (Amplified)",
				Tag:      buffs.CurseOfWeaknessCategory,
				ActionID: core.ActionID{SpellID: rankedSpell.ID, Tag: 1},
				Duration: rankedSpell.Duration(),
			})
			opts := append(spelldata.RaidBuffOptions(nil, false),
				spelldata.ScaledBy(ampWeakness),
				spelldata.Exclusive(buffs.CurseOfWeaknessCategory, true))
			spelldata.ParseEffects(nil, aura, rankedSpell, opts...)
			return aura
		})
		config.RelatedAuraArrays[warlock.CurseOfWeaknessAmpAuras.FindLabel()] = warlock.CurseOfWeaknessAmpAuras
	}

	// Roll to hit, and apply in place of any other Curse: amplified if Amplify Curse is up, which the
	// cast spends after its effects land.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			warlock.deactivateOtherCurses(sim, spell, target)

			// The amplified array is only built with the talent; without it, Get answers nil.
			apply, replace := warlock.CurseOfWeaknessAuras.Get(target), warlock.CurseOfWeaknessAmpAuras.Get(target)
			if replace != nil && warlock.AmplifyCurseAura.IsActive() {
				apply, replace = replace, apply
			}
			if replace != nil {
				replace.Deactivate(sim)
			}
			apply.Activate(sim)
		}
		spell.DealOutcome(sim, result)
	}

	warlock.CurseOfWeakness = warlock.RegisterSpell(config)
	warlock.Curses = append(warlock.Curses, warlock.CurseOfWeakness)
}
