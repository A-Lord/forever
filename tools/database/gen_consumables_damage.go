package database

import (
	"fmt"

	"github.com/wowsims/forever/sim/core/spelldata"
)

// The damage builder puts a damage over time on the one target the spell is used on, so a row whose
// damage over time lands on every enemy in an area is not what it builds.
func damageConsumableUnsupported(s *spelldata.Spell) []string {
	if periodic := s.PeriodicDamageEffect(); periodic != spelldata.NilEffect && periodic.HitsAnArea() {
		return []string{fmt.Sprintf("%d's damage over time lands on an area, which the damage builder applies to one target", s.ID)}
	}
	return nil
}
