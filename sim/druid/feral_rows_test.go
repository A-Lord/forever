package druid

import (
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func TestBerserkLeavesJumpTargetsUnparsed(t *testing.T) {
	parsed := spelldata.DryRun(spellData.Berserk.Highest())
	var jump bool
	var crit bool
	for _, skipped := range parsed.Skipped {
		if skipped.Aura == dbcenums.A_ADD_FLAT_MODIFIER && dbcenums.SpellModOp(skipped.Misc) == dbcenums.SPELLMOD_JUMP_TARGETS {
			jump = true
			if skipped.BasePoints != 3 {
				t.Errorf("jump targets base points = %v, want 3", skipped.BasePoints)
			}
		}
	}
	for _, applied := range parsed.Applied {
		if applied.Kind == "SpellMod_BonusCrit_Percent" {
			crit = true
			if applied.Value != 100 {
				t.Errorf("berserk crit mod = %v, want 100", applied.Value)
			}
		}
	}
	if !jump {
		t.Errorf("SPELLMOD_JUMP_TARGETS was parsed or missing; skipped %#v", parsed.Skipped)
	}
	if !crit {
		t.Errorf("berserk crit mod was not parsed; applied %#v", parsed.Applied)
	}
}

func TestOmenOfClarityUsesTheClassicRate(t *testing.T) {
	if got := spellData.OmenOfClarity.Highest().RPPM; got != 2 {
		t.Errorf("Omen of Clarity RPPM = %v, want the classic placeholder 2", got)
	}
}

