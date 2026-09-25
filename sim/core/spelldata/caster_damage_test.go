package spelldata

import "testing"

func TestCasterDamageEffectIsTheHalfThatHitsTheCaster(t *testing.T) {
	defer generatedStore()()

	sapper := MustFind(13241)
	if got := sapper.CasterDamageEffect(); got.Index != 1 || got.BasePoints != 500 || got.Variance != 0.5 {
		t.Errorf("13241's caster damage is effect %d, %v ±%v; want effect 1, 500 ±0.5", got.Index, got.BasePoints, got.Variance)
	}
	if got := sapper.DamageEffect(); got.Index != 0 || got.BasePoints != 600 {
		t.Errorf("13241's damage is effect %d, %v; want effect 0, 600", got.Index, got.BasePoints)
	}

	if got := MustFind(23000).CasterDamageEffect(); got != NilEffect {
		t.Errorf("Ez-Thro Dynamite II 23000 damages its caster through effect %d, want none", got.Index)
	}
}
