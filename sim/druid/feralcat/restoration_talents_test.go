//go:build with_db

package feralcat

import (
	"fmt"
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestReflectionRestoresCastingManaRegen(t *testing.T) {
	for rank, fraction := range []float64{0, 0.17, 0.33, 0.50} {
		t.Run(fmt.Sprintf("rank_%d", rank), func(t *testing.T) {
			// Given a cat with no mana buffs or equipment mana regeneration.
			talents := fmt.Sprintf("0000000000000000-00000000000000000000-00000%d0000000000", rank)
			_, cat := newCat(t, talents, nil)

			// When the mana bar computes regeneration during casting.
			got := cat.ManaRegenPerSecondWhileCasting()

			// Then the game files' rank curve preserves this share of spirit regeneration.
			want := cat.SpiritManaRegenPerSecond() * fraction
			if math.Abs(got-want) > 1e-9 {
				t.Errorf("casting mana regeneration = %v, want %v at fraction %v", got, want, fraction)
			}
			t.Logf("rank %d casting mana regeneration = %v, spirit fraction = %v", rank, got, cat.PseudoStats.SpiritRegenRateCasting)
		})
	}
}

func TestSubtletyReducesNatureThreatWithoutReducingPhysicalThreat(t *testing.T) {
	for rank, multiplier := range []float64{1, 0.9, 0.8, 0.7} {
		t.Run(fmt.Sprintf("rank_%d", rank), func(t *testing.T) {
			// Given a cat with registered Nature and physical abilities.
			talents := fmt.Sprintf("0000000000000000-00000000000000000000-000%d000000000000", rank)
			sim, cat := newCat(t, talents, nil)
			arcane := cat.RegisterSpell(core.SpellConfig{
				ActionID:         core.ActionID{SpellID: 1},
				SpellSchool:      core.SpellSchoolArcane,
				ProcMask:         core.ProcMaskSpellDamage,
				DamageMultiplier: 1,
				ThreatMultiplier: 1,
				FlatThreatBonus:  20,
			})
			for _, ability := range []struct {
				name  string
				spell *core.Spell
				want  float64
			}{
				{"nature", cat.FaerieFire.Spell, 71 * multiplier},
				{"arcane with flat threat", arcane.Spell, 85.2 * multiplier},
				{"physical", cat.Claw.Spell, 71},
			} {
				t.Run(ability.name, func(t *testing.T) {
					// When the registered spell converts a landed damage amount to threat.
					got := ability.spell.ThreatFromDamage(sim, core.OutcomeHit, 100, cat.AttackTables[cat.CurrentTarget.UnitIndex])

					// Then Subtlety affects only the row's school mask.
					if math.Abs(got-ability.want) > 1e-9 {
						t.Errorf("threat = %v, want %v", got, ability.want)
					}
					t.Logf("rank %d %s threat = %v", rank, ability.name, got)
				})
			}
		})
	}
}

func TestLivingSpiritIncreasesTotalSpirit(t *testing.T) {
	_, base := newCat(t, "", nil)
	for rank, multiplier := range []float64{1, 1.05, 1.10, 1.15} {
		t.Run(fmt.Sprintf("rank_%d", rank), func(t *testing.T) {
			// Given the same cat with a rank of Living Spirit.
			talents := fmt.Sprintf("0000000000000000-00000000000000000000-000000000000%d000", rank)
			_, cat := newCat(t, talents, nil)

			// When the character's stat dependencies resolve.
			got := cat.GetStat(stats.Spirit)

			// Then total Spirit includes the rank's percentage.
			want := math.Floor(base.GetStat(stats.Spirit) * multiplier)
			if math.Abs(got-want) > 1e-9 {
				t.Errorf("Spirit = %v, want %v at multiplier %v", got, want, multiplier)
			}
			t.Logf("rank %d Spirit = %v (untalented %v)", rank, got, base.GetStat(stats.Spirit))
		})
	}
}

func TestNaturalistIncreasesCatDamage(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0000000000000000-00000000000000000000-0050000000000000", nil)
	if got, want := talented.PseudoStats.DamageDealtMultiplier/base.PseudoStats.DamageDealtMultiplier, 1.05; math.Abs(got-want) > 1e-9 {
		t.Errorf("damage dealt multiplier ratio = %v, want %v", got, want)
	}
}
