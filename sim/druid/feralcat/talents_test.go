//go:build with_db

package feralcat

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
	"math"
	"time"

	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

// The druid trees as the talent calculator draws them: per tree, each talent's row and highest rank.
type talentTree struct {
	Name    string `json:"name"`
	Talents []struct {
		FieldName string `json:"fieldName"`
		Location  struct {
			RowIdx int `json:"rowIdx"`
		} `json:"location"`
		MaxPoints int `json:"maxPoints"`
	} `json:"talents"`
}

func druidTrees(t *testing.T) []talentTree {
	t.Helper()
	raw, err := os.ReadFile("../../../ui/sim/talents/trees/druid.json")
	if err != nil {
		t.Fatal(err)
	}
	var trees []talentTree
	if err := json.Unmarshal(raw, &trees); err != nil {
		t.Fatal(err)
	}
	return trees
}

// talentsLegal reports why a talent string cannot be built at level 60: a rank over the talent's
// highest, a row reached before 5 points per row are spent above it in the same tree, or more than
// 51 points.
func talentsLegal(trees []talentTree, talents string) error {
	parts := strings.Split(talents, "-")
	if len(parts) != len(trees) {
		return fmt.Errorf("%d trees, want %d", len(parts), len(trees))
	}
	total := 0
	for i, tree := range trees {
		digits := parts[i]
		if len(digits) > len(tree.Talents) {
			return fmt.Errorf("%s: %d talents, the tree has %d", tree.Name, len(digits), len(tree.Talents))
		}
		perRow := map[int]int{}
		for j := range digits {
			points := int(digits[j] - '0')
			talent := tree.Talents[j]
			if points > talent.MaxPoints {
				return fmt.Errorf("%s: %s %d of %d", tree.Name, talent.FieldName, points, talent.MaxPoints)
			}
			perRow[talent.Location.RowIdx] += points
			total += points
		}
		for j := range digits {
			talent := tree.Talents[j]
			if digits[j] == '0' {
				continue
			}
			above := 0
			for row := 0; row < talent.Location.RowIdx; row++ {
				above += perRow[row]
			}
			if need := 5 * talent.Location.RowIdx; above < need {
				return fmt.Errorf("%s: %s is on row %d with %d points above, needs %d", tree.Name, talent.FieldName, talent.Location.RowIdx+1, above, need)
			}
		}
	}
	if total > 51 {
		return fmt.Errorf("%d points, a level 60 druid has 51", total)
	}
	return nil
}

// Every talent string the cat and bear UIs ship can be built in game.
func TestPresetTalentsAreLegal(t *testing.T) {
	trees := druidTrees(t)
	pattern := regexp.MustCompile(`talentsString: '([0-9-]+)'`)
	for _, file := range []string{
		"../../../ui/specs/druid/feralcat/presets.ts",
		"../../../ui/specs/druid/feralbear/presets.ts",
		"../../../ui/specs/druid/feralbear/spec.ts",
	} {
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range pattern.FindAllStringSubmatch(string(source), -1) {
			if err := talentsLegal(trees, match[1]); err != nil {
				t.Errorf("%s: %s: %v", file, match[1], err)
			}
		}
	}
	for name, talents := range map[string]string{"catTalents": catTalents, "bearweaveTalents": bearweaveTalents} {
		if err := talentsLegal(trees, talents); err != nil {
			t.Errorf("%s %s: %v", name, talents, err)
		}
	}
}

func TestGenesisThreePieceIncludesDefenseAndArmor(t *testing.T) {
	sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHead: 21353, proto.ItemSlot_ItemSlotShoulder: 21354, proto.ItemSlot_ItemSlotChest: 21357,
	})
	aura := cat.GetAura("Genesis Raiment 3P")
	if aura == nil || !aura.IsActive() {
		t.Fatal("Genesis three-piece is inactive")
	}
	armor := cat.GetStat(stats.Armor)
	if got := cat.GetStat(stats.DefenseRating); got != 15 {
		t.Errorf("Genesis Defense = %g, want 15", got)
	}
	aura.Deactivate(sim)
	if got := cat.GetStat(stats.DefenseRating); got != 0 {
		t.Errorf("expired Genesis Defense = %g, want 0", got)
	}
	if got := armor - cat.GetStat(stats.Armor); got != 150 {
		t.Errorf("Genesis armor = %g, want 150", got)
	}
}

func TestGrovekeeperFerocityTwoPieceRaisesCatAttackSpeed(t *testing.T) {
	sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHands: 273889,
		proto.ItemSlot_ItemSlotLegs:  273888,
	})
	aura := cat.GetAura("Grovekeeper Ferocity 2P")
	if aura == nil || !aura.IsActive() {
		t.Fatal("Grovekeeper Ferocity two-piece bonus is inactive")
	}
	if got := cat.PseudoStats.AttackSpeedMultiplier; math.Abs(got-1.01) > 1e-9 {
		t.Errorf("Cat melee speed multiplier = %v, want 1.01", got)
	}
	aura.Deactivate(sim)
	if got := cat.PseudoStats.AttackSpeedMultiplier; got != 1 {
		t.Errorf("Cat melee speed multiplier after expiry = %v, want 1", got)
	}
}

func TestGrovekeeperFerocityFourPieceAddsAttackPowerAgainstDemons(t *testing.T) {
	sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotShoulder: 273890,
		proto.ItemSlot_ItemSlotHands:    273889,
		proto.ItemSlot_ItemSlotLegs:     273888,
		proto.ItemSlot_ItemSlotFeet:     273891,
	})
	aura := cat.GetAura("Grovekeeper Ferocity 4P")
	if aura == nil || !aura.IsActive() {
		t.Fatal("Grovekeeper Ferocity four-piece bonus is inactive")
	}
	aura.Deactivate(sim)
	cat.CurrentTarget.MobType = proto.MobType_MobTypeDemon
	aura.Activate(sim)
	table := cat.AttackTables[cat.CurrentTarget.UnitIndex]
	if got := table.MobTypeBonusStats[proto.MobType_MobTypeDemon][stats.AttackPower]; got != 36 {
		t.Errorf("attack power against Demons = %v, want 36", got)
	}
	aura.Deactivate(sim)
	if got := table.MobTypeBonusStats[proto.MobType_MobTypeDemon][stats.AttackPower]; got != 0 {
		t.Errorf("attack power against Demons after expiry = %v, want 0", got)
	}
}

func TestGrovekeeperFerocityFivePieceRegistration(t *testing.T) {
	if druid.ItemSetGrovekeeperFerocity.ID != 2115 {
		t.Errorf("Grovekeeper Ferocity set ID = %d, want 2115", druid.ItemSetGrovekeeperFerocity.ID)
	}
	if druid.ItemSetGrovekeeperFerocity.Bonuses[5] == nil {
		t.Fatal("Grovekeeper Ferocity five-piece bonus is unregistered")
	}
	effect := spelldata.MustFind(1301247).EffectN(1)
	if got := effect.TimeValue(); got != -time.Second {
		t.Errorf("Grovekeeper Ferocity cooldown modifier = %v, want -1s", got)
	}
}

func TestGenesisIncreasesCatBleedDamage(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0500000000000000-00000000000000000000-0000000000000000", nil)
	for _, bleed := range []struct {
		name       string
		base, with float64
	}{
		{"Rake", base.Rake.Dot(base.CurrentTarget).PeriodicDamageMultiplier, talented.Rake.Dot(talented.CurrentTarget).PeriodicDamageMultiplier},
		{"Rip", base.Rip.Dot(base.CurrentTarget).PeriodicDamageMultiplier, talented.Rip.Dot(talented.CurrentTarget).PeriodicDamageMultiplier},
	} {
		if got, want := bleed.with/bleed.base, 1.05; math.Abs(got-want) > 1e-9 {
			t.Errorf("%s periodic damage multiplier ratio = %v, want %v", bleed.name, got, want)
		}
	}
}

func TestNaturesMajestyRaisesCatMeleeAndSpellCrit(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0000200000000000-00000000000000000000-0000000000000000", nil)
	want := 4.0
	for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
		if got := talented.GetStat(stat) - base.GetStat(stat); math.Abs(got-want) > 1e-9 {
			t.Errorf("Nature's Majesty %s = %v, want %v", stat.StatName(), got, want)
		}
	}
}

func TestNaturesReachRaisesCatMeleeAndSpellHit(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0000020000000000-00000000000000000000-0000000000000000", nil)
	want := 4.0
	for _, stat := range []stats.Stat{stats.PhysicalHitPercent, stats.SpellHitPercent} {
		if got := talented.GetStat(stat) - base.GetStat(stat); math.Abs(got-want) > 1e-9 {
			t.Errorf("Nature's Reach %s = %v, want %v", stat.StatName(), got, want)
		}
	}
}

func TestNaturalistIncreasesCatDamage(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0000000000000000-00000000000000000000-0050000000000000", nil)
	if got, want := talented.PseudoStats.DamageDealtMultiplier/base.PseudoStats.DamageDealtMultiplier, 1.05; math.Abs(got-want) > 1e-9 {
		t.Errorf("damage dealt multiplier ratio = %v, want %v", got, want)
	}
}

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
