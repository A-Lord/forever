package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The debuff is this warlock's own copy of the one the raid option builds, parsed from the row: -75 to
// all resistances and 10% more damage taken. Its category keeps it from stacking with another warlock's
// copy or a raid one.
func (warlock *Warlock) registerCurseOfElements() {
	rankedSpell := spellData.CurseOfTheElements.ByID(1311680)

	warlock.CurseOfElementsAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataCurseOfElementsAuraFor(target, &warlock.Unit)
	})
	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1
	config.ClassSpellMask = WarlockSpellCurseOfElements
	config.RelatedAuraArrays = warlock.CurseOfElementsAuras.ToMap()

	if warlock.Talents.AmplifyCurse {
		// Amplify Curse's second effect is a flat SPELLMOD_EFFECT1 of -20 naming this curse. Scaling the
		// parse by it prices only the first effect attached, the resistance: -75 becomes -95.
		ampElements := spellData.AmplifyCurse.ByID(18288).EffectN(2)

		// The same row, built the way the shared debuff is, with Amplify Curse's -20 parsed onto it.
		warlock.CurseOfElementsAmpAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			aura := target.GetOrRegisterAura(core.Aura{
				Label:    "Curse of the Elements (Amplified)-" + warlock.Label,
				Tag:      buffs.CurseOfElementsCategory,
				ActionID: core.ActionID{SpellID: rankedSpell.ID, Tag: 1},
				Duration: rankedSpell.Duration(),
			})
			opts := append(spelldata.RaidBuffOptions(nil, false),
				spelldata.ScaledBy(ampElements),
				spelldata.Exclusive(buffs.CurseOfElementsCategory, false))
			spelldata.ParseEffects(nil, aura, rankedSpell, opts...)
			return aura
		})
		config.RelatedAuraArrays[warlock.CurseOfElementsAmpAuras.FindLabel()] = warlock.CurseOfElementsAmpAuras
	}

	// Roll to hit, and apply in place of any other Curse: amplified if Amplify Curse is up, which the
	// cast spends after its effects land.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			warlock.deactivateOtherCurses(sim, spell, target)

			// The amplified array is only built with the talent; without it, Get answers nil.
			apply, replace := warlock.CurseOfElementsAuras.Get(target), warlock.CurseOfElementsAmpAuras.Get(target)
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

	warlock.CurseOfElements = warlock.RegisterSpell(config)
	warlock.Curses = append(warlock.Curses, warlock.CurseOfElements)
}
