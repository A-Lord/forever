//go:build with_db

package feralcat

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
)

func TestFeralStaticSetBonuses(t *testing.T) {
	wildheart := []int32{16716, 16715, 16714, 16720, 16706, 16718}
	feralheart := []int32{22106, 22107, 22108, 22109, 22110, 22111, 22112, 22113}
	for _, bonus := range []struct {
		set    string
		pieces []int32
		stat   stats.Stat
		want   float64
	}{
		{"Wildheart Raiment", wildheart[:2], stats.FireResistance, 8},
		{"Wildheart Raiment", wildheart[:3], stats.AttackPower, 30},
		{"Wildheart Raiment", wildheart[:3], stats.SpellDamage, 18},
		{"Wildheart Raiment", wildheart, stats.MP5, 8},
		{"Feralheart Raiment", feralheart[:2], stats.FireResistance, 8},
		{"Feralheart Raiment", feralheart[:6], stats.AttackPower, 26},
		{"Feralheart Raiment", feralheart[:6], stats.SpellDamage, 15},
		{"Feralheart Raiment", feralheart, stats.Armor, 200},
	} {
		for _, form := range []druid.DruidForm{druid.Cat, druid.DireBear} {
			t.Run(fmt.Sprintf("%s_%d_%s_form_%d", bonus.set, len(bonus.pieces), bonus.stat.StatName(), form), func(t *testing.T) {
				// Given a real character wearing the threshold's source-listed items.
				worn := map[proto.ItemSlot]int32{}
				for _, id := range bonus.pieces {
					item := core.GetItemByID(id)
					if item == nil {
						t.Fatalf("set item %d is absent", id)
					}
					worn[core.ItemTypeToSlot(item.Type)] = id
				}
				sim, cat := newCatWearing(t, "", nil, worn)
				if form == druid.DireBear {
					fillResource(sim, cat)
					if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
						t.Fatal("dire bear form did not cast")
					}
				}
				aura := cat.GetAura(fmt.Sprintf("%s %dP", bonus.set, len(bonus.pieces)))
				if aura == nil || !aura.IsActive() {
					t.Fatalf("%s %d-piece bonus is inactive", bonus.set, len(bonus.pieces))
				}
				before := cat.GetStat(bonus.stat)

				// When the set status aura expires, as it does on an item swap below the threshold.
				aura.Deactivate(sim)

				// Then the source bonus is removed once, independently of the equipped item stats.
				got := before - cat.GetStat(bonus.stat)
				if math.Abs(got-bonus.want) > 1e-9 {
					t.Errorf("removed %s = %v, want %v", bonus.stat.StatName(), got, bonus.want)
				}
				t.Logf("form %d removed %s = %v", form, bonus.stat.StatName(), got)
			})
		}
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
