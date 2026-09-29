package protection

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
)

// Concussion Blow deals no damage, but it is a melee ability like any other: its threat multiplier is
// 1 for a threat modifier to scale, and its casts are counted with the melee abilities.
func TestConcussionBlowIsAMeleeAbility(t *testing.T) {
	_, war, _ := shieldTestSim(t, DefaultProtectionTalents, 0)

	spell := war.GetSpell(core.ActionID{SpellID: 12809})
	if spell == nil {
		t.Fatal("the warrior has no Concussion Blow")
	}
	if spell.ThreatMultiplier != 1 {
		t.Errorf("Concussion Blow's threat multiplier is %v, want 1", spell.ThreatMultiplier)
	}
	if !spell.Flags.Matches(core.SpellFlagMeleeMetrics) || spell.ProcMask != core.ProcMaskMeleeMHSpecial {
		t.Errorf("Concussion Blow has flags %v and proc mask %v, want melee metrics and a main-hand special", spell.Flags, spell.ProcMask)
	}
}
