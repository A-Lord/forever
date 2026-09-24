package paladin

import (
	"github.com/wowsims/forever/sim/core"
)

var forbearanceRow = spellData.TemplarsBulwarkTriggered.HighestRank()

// Forbearance
// https://www.wowhead.com/forever/spell=25771
//
// Cannot be made invulnerable by Divine Shield, Divine Protection, Blessing of Protection, or
// shielded by Templar's Bulwark.
func (paladin *Paladin) registerForbearance() {
	paladin.Forbearance = paladin.RegisterAura(core.Aura{
		Label:    "Forbearance",
		ActionID: core.ActionID{SpellID: forbearanceRow.SpellID},
		Duration: forbearanceRow.Duration - paladin.forbearanceReduction,
	})
}
