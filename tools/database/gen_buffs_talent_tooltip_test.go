package database

import "testing"

// Vindication's talent states its effect only in its tooltip: rank/3 of the debuff's amount.
func TestTalentPricedByItsTooltip(t *testing.T) {
	for _, row := range resolveCommittedBuffs(t) {
		if row.Field != "vindication" {
			continue
		}
		if row.TalentRanks != 3 || row.TalentDivisor != 3 || row.TalentSpellID != 9452 {
			t.Errorf("vindication resolved to ranks %d divisor %d talent %d, want 3, 3, 9452",
				row.TalentRanks, row.TalentDivisor, row.TalentSpellID)
		}
		return
	}
	t.Fatal("no vindication row")
}

func TestTooltipDivisor(t *testing.T) {
	cases := []struct {
		text    string
		spellID int32
		want    int32
	}{
		{"reduce the target's Attack Power by ${$m1/-3*$440667m1}, and", 440667, 3},
		{"reduce the target's Attack Power by ${$m1/-3*$440667m1}, and", 11556, 0},
		{"increase your Attack Power by $m1%", 440667, 0},
	}
	for _, c := range cases {
		if got := tooltipTalentDivisor(c.text, c.spellID); got != c.want {
			t.Errorf("tooltipTalentDivisor(%q, %d) = %d, want %d", c.text, c.spellID, got, c.want)
		}
	}
}
