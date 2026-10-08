package stats

import (
	"testing"
)

func TestStatDependencies(t *testing.T) {
	baseStat := Stats{
		Stamina:   1,
		Intellect: 1,
		Agility:   2,
		Spirit:    1,
	}

	sdm := NewStatDependencyManager()

	sdm.MultiplyStat(Intellect, 2)
	sdm.AddStatDependency(Stamina, Intellect, 1)
	sdm.MultiplyStat(Stamina, 2)
	sdm.AddStatDependency(Agility, Stamina, 1)

	dep1 := sdm.NewDynamicMultiplyStat(Agility, 2)
	dep2 := sdm.NewDynamicMultiplyStat(Spirit, 3)
	dep3 := sdm.NewDynamicStatDependency(Agility, Spirit, 0.75)

	sdm.FinalizeStatDeps()

	result := sdm.ApplyStatDependencies(baseStat)
	expectedResult := Stats{
		Stamina:   6,
		Intellect: 14,
		Agility:   2,
		Spirit:    1,
	}
	if !result.Equals(expectedResult) {
		t.Fatalf("Stats do not match:\nActual: %s\nExpected: %s", result, expectedResult)
	}

	sdm.EnableDynamicStatDep(dep1)
	result2 := sdm.ApplyStatDependencies(baseStat)
	expectedResult2 := Stats{
		Stamina:   10,
		Intellect: 22,
		Agility:   4,
		Spirit:    1,
	}
	if !result2.Equals(expectedResult2) {
		t.Fatalf("Updated stats do not match:\nActual: %s\nExpected: %s", result2, expectedResult2)
	}

	sdm.EnableDynamicStatDep(dep2)
	result3 := sdm.ApplyStatDependencies(baseStat)
	expectedResult3 := Stats{
		Stamina:   10,
		Intellect: 22,
		Agility:   4,
		Spirit:    3,
	}
	if !result3.Equals(expectedResult3) {
		t.Fatalf("Updated stats do not match:\nActual: %s\nExpected: %s", result3, expectedResult3)
	}

	sdm.DisableDynamicStatDep(dep2)
	result4 := sdm.ApplyStatDependencies(baseStat)
	if !result4.Equals(expectedResult2) {
		t.Fatalf("Updated stats do not match:\nActual: %s\nExpected: %s", result4, expectedResult2)
	}

	sdm.EnableDynamicStatDep(dep3)
	result5 := sdm.ApplyStatDependencies(baseStat)
	expectedResult5 := Stats{
		Stamina:   10,
		Intellect: 22,
		Agility:   4,
		Spirit:    4,
	}
	if !result5.Equals(expectedResult5) {
		t.Fatalf("Updated stats do not match:\nActual: %s\nExpected: %s", result5, expectedResult5)
	}

	sdm.DisableDynamicStatDep(dep3)
	result6 := sdm.ApplyStatDependencies(baseStat)
	if !result6.Equals(expectedResult2) {
		t.Fatalf("Updated stats do not match:\nActual: %s\nExpected: %s", result6, expectedResult2)
	}
}

func TestFlooredStatDepInPlace(t *testing.T) {
	sdm := NewStatDependencyManager()
	sdm.AddFlooredStatDependency(DefenseRating, ReducedCritTakenPercent, 2.5, 0.1)
	sdm.FinalizeStatDeps()

	baseStat := Stats{DefenseRating: 9}
	expectedResult := Stats{DefenseRating: 9, ReducedCritTakenPercent: 7.5 * 0.1}

	result := sdm.ApplyStatDependencies(baseStat)
	if !result.Equals(expectedResult) {
		t.Fatalf("Stats do not match:\nActual: %s\nExpected: %s", result, expectedResult)
	}

	inPlace := baseStat
	sdm.ApplyStatDependenciesInPlace(&inPlace)
	if !inPlace.Equals(expectedResult) {
		t.Fatalf("In-place stats do not match:\nActual: %s\nExpected: %s", inPlace, expectedResult)
	}
}

func TestMultipleStatDep(t *testing.T) {
	sdm := NewStatDependencyManager()

	baseStat := Stats{
		Intellect:   100,
		SpellDamage: 100,
	}

	sdm.AddStatDependency(Intellect, SpellDamage, 0.2)
	sdm.AddStatDependency(Intellect, SpellDamage, 0.2)
	sdm.MultiplyStat(Intellect, 1.2)
	sdm.FinalizeStatDeps()
	result := sdm.ApplyStatDependencies(baseStat)

	expectedResult := Stats{
		Intellect:   100 * 1.2,
		SpellDamage: 100 + (100*1.2)*(0.2+0.2),
	}

	if !result.Equals(expectedResult) {
		t.Fatalf("Stats do not match:\nActual: %s\nExpected: %s", result, expectedResult)
	}
}

// The doubling scales the 200 armor outside the share and adds the 100 of the share after it.
func TestAddUnmultipliedStaysOutOfTheSameStatMultiplier(t *testing.T) {
	sdm := NewStatDependencyManager()
	sdm.MultiplyStat(Armor, 2)
	sdm.AddUnmultiplied(Armor, 100)
	sdm.FinalizeStatDeps()

	if got := sdm.ApplyStatDependencies(Stats{Armor: 300})[Armor]; got != 500 {
		t.Errorf("armor: %v, want (300-100)*2+100 = 500", got)
	}
}

// A share added during an iteration is gone after the reset, which returns to the share the deps were
// finalized with however often it runs.
func TestResetStatDepsRestoresTheUnmultipliedShare(t *testing.T) {
	sdm := NewStatDependencyManager()
	sdm.MultiplyStat(Armor, 2)
	sdm.AddUnmultiplied(Armor, 100)
	sdm.FinalizeStatDeps()

	for iteration := 1; iteration <= 3; iteration++ {
		sdm.AddUnmultiplied(Armor, -450)
		if got := sdm.ApplyStatDependencies(Stats{Armor: 300})[Armor]; got != 950 {
			t.Errorf("iteration %d, armor with the -450 share: %v, want (300+350)*2-350 = 950", iteration, got)
		}
		sdm.ResetStatDeps()
		if got := sdm.ApplyStatDependencies(Stats{Armor: 300})[Armor]; got != 500 {
			t.Errorf("iteration %d, armor after the reset: %v, want the finalized 500", iteration, got)
		}
	}
}

// A change in stats is scaled whole: the share is part of the stat it changes, not of the change.
func TestApplyStatDependenciesToChangeIgnoresTheUnmultipliedShare(t *testing.T) {
	sdm := NewStatDependencyManager()
	sdm.MultiplyStat(Armor, 2)
	sdm.AddUnmultiplied(Armor, 100)
	sdm.FinalizeStatDeps()

	if got := sdm.ApplyStatDependenciesToChange(Stats{Armor: 300})[Armor]; got != 600 {
		t.Errorf("a change of 300 armor: %v, want 300*2 = 600", got)
	}
}
