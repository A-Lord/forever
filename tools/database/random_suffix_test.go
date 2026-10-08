package database

import (
	"maps"
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

// A few of each item's random suffixes, by name, with the stats they add to the item: the bonus list's
// allocation times the item's random property points over 10000, rounded to the nearest point.
func TestRandomSuffixOptions(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	type stats = map[proto.Stat]float64
	spellPower := func(value float64) stats {
		return stats{proto.Stat_StatSpellDamage: value, proto.Stat_StatHealingPower: value}
	}
	with := func(a, b stats) stats {
		out := maps.Clone(a)
		maps.Copy(out, b)
		return out
	}

	for _, tc := range []struct {
		name    string
		id      int
		options map[string]stats
	}{
		{"Prismatic Band, a ring of 16 random property points", 12017, map[string]stats{
			"of the Eagle": {proto.Stat_StatStamina: 11, proto.Stat_StatIntellect: 11},
			"of Power":     {proto.Stat_StatAttackPower: 32},
			// 4106 Intellect allocation is 6.57.
			"of the Physician": {proto.Stat_StatIntellect: 7, proto.Stat_StatStamina: 12, proto.Stat_StatHealingPower: 11, proto.Stat_StatSpellDamage: 4},
			"of the Sorcerer":  with(stats{proto.Stat_StatIntellect: 12, proto.Stat_StatStamina: 7}, spellPower(6)),
		}},
		{"Solstice Staff, a two-hand weapon of 26", 15278, map[string]stats{
			"of the Tiger": {proto.Stat_StatAgility: 17, proto.Stat_StatStrength: 17},
			"of Magic":     with(stats{proto.Stat_StatIntellect: 26}, spellPower(52)),
			"of Healing":   {proto.Stat_StatIntellect: 26, proto.Stat_StatHealingPower: 98, proto.Stat_StatSpellDamage: 33},
		}},
		{"Dark Iron Boots, epic feet of 38", 20039, map[string]stats{
			"of the Eagle":  {proto.Stat_StatStamina: 14, proto.Stat_StatIntellect: 14},
			"of Defense":    {proto.Stat_StatDefenseRating: 14},
			"of the Knight": with(stats{proto.Stat_StatStrength: 7, proto.Stat_StatStamina: 13}, spellPower(7)),
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item, ok := instance.Items[tc.id]
			if !ok {
				t.Fatalf("item %d is not in the DBC inputs", tc.id)
			}
			points := float64(item.GetRandPropPoints(item.ItemLevel))

			got := map[string][]stats{}
			for _, id := range item.RandomSuffixOptions {
				suffix, ok := instance.RandomSuffix[int(id)]
				if !ok {
					t.Fatalf("option %d is not a random suffix", id)
				}
				added := stats{}
				for stat, allocation := range suffix.ToProto().Stats {
					if allocation != 0 {
						added[proto.Stat(stat)] = math.Round(allocation * points / 10000)
					}
				}
				got[suffix.Name] = append(got[suffix.Name], added)
			}

			for name, want := range tc.options {
				found := false
				for _, added := range got[name] {
					found = found || maps.Equal(added, want)
				}
				if !found {
					t.Errorf("%s: got %v, want %v", name, got[name], want)
				}
			}
		})
	}

	t.Run("of Eluding adds 1% dodge, whatever the item's points", func(t *testing.T) {
		suffix := instance.RandomSuffix[12739].ToProto()
		if suffix.Name != "of Eluding" {
			t.Fatalf("bonus list 12739 is %q, want of Eluding", suffix.Name)
		}
		if got := suffix.PseudoStats[proto.PseudoStat_PseudoStatDodgePercent]; got != 1 {
			t.Errorf("dodge: got %v, want 1", got)
		}
	})
}
