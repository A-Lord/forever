package core_test

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestRogueDodgesFromAgility(t *testing.T) {
	unit := &core.Unit{StatDependencyManager: stats.NewStatDependencyManager()}
	unit.AddStatDependency(stats.Agility, stats.DodgePercent, core.DodgePerAgiMaxLevel[proto.Class_ClassRogue])
	unit.FinalizeStatDeps()

	sheet := unit.ApplyStatDependencies(stats.Stats{stats.Agility: 130})
	if got := sheet[stats.DodgePercent]; math.Abs(got-130*0.069) > 1e-9 {
		t.Errorf("130 agility gives a rogue %v%% dodge, want %v%%", got, 130*0.069)
	}
}

func TestEveryClassDodgesFromAgility(t *testing.T) {
	for class, perAgi := range core.DodgePerAgiMaxLevel {
		if class != proto.Class_ClassUnknown && perAgi <= 0 {
			t.Errorf("%v gains %v dodge per agility, want a positive amount", class, perAgi)
		}
	}
}

func TestClassBaseAvoidance(t *testing.T) {
	for _, tc := range []struct {
		class proto.Class
		race  proto.Race
		dodge float64
	}{
		{proto.Class_ClassWarrior, proto.Race_RaceHuman, 0},
		{proto.Class_ClassPaladin, proto.Race_RaceHuman, 0.007},
		{proto.Class_ClassPriest, proto.Race_RaceHuman, 0.03},
		{proto.Class_ClassShaman, proto.Race_RaceOrc, 0.017},
		{proto.Class_ClassMage, proto.Race_RaceHuman, 0.032},
		{proto.Class_ClassWarlock, proto.Race_RaceHuman, 0.02},
		{proto.Class_ClassDruid, proto.Race_RaceTauren, 0.009},
	} {
		t.Run(tc.class.String(), func(t *testing.T) {
			character, sheet := nakedSheet(t, tc.class, tc.race)
			if got := character.PseudoStats.BaseDodgeChance; math.Abs(got-tc.dodge) > 1e-12 {
				t.Errorf("base dodge %v, want %v", got, tc.dodge)
			}
			if got, want := sheet[stats.DodgePercent], sheet[stats.Agility]*core.DodgePerAgiMaxLevel[tc.class]; math.Abs(got-want) > 1e-9 {
				t.Errorf("dodge from %v agility %v%%, want %v%%", sheet[stats.Agility], got, want)
			}
			if character.PseudoStats.BaseParryChance != 0 || character.PseudoStats.BaseBlockChance != 0 {
				t.Errorf("class adds parry %v and block %v, want both from the attack table alone",
					character.PseudoStats.BaseParryChance, character.PseudoStats.BaseBlockChance)
			}
		})
	}
}

func TestWarriorAndPaladinParry(t *testing.T) {
	for _, class := range []proto.Class{proto.Class_ClassWarrior, proto.Class_ClassPaladin} {
		character, _ := nakedSheet(t, class, proto.Race_RaceHuman)
		if !character.PseudoStats.CanParry {
			t.Errorf("%v cannot parry", class)
		}
	}
}

func TestBlockValueFromStrength(t *testing.T) {
	for _, tc := range []struct {
		class       proto.Class
		race        proto.Race
		perStrength float64
	}{
		{proto.Class_ClassWarrior, proto.Race_RaceHuman, 0.05},
		{proto.Class_ClassPaladin, proto.Race_RaceHuman, 0.05},
		{proto.Class_ClassShaman, proto.Race_RaceOrc, 0.05},
		{proto.Class_ClassDruid, proto.Race_RaceTauren, 0},
	} {
		t.Run(tc.class.String(), func(t *testing.T) {
			_, sheet := nakedSheet(t, tc.class, tc.race)
			want := max(0, sheet[stats.Strength]*tc.perStrength-1)
			if got := sheet[stats.BlockValue]; math.Abs(got-want) > 1e-9 {
				t.Errorf("block value %v from %v strength, want %v", got, sheet[stats.Strength], want)
			}
		})
	}
}
