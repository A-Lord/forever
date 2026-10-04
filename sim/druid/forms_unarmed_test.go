//go:build with_db

package druid

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestFeralPortPawLaw(t *testing.T) {
	const attackPower = 140

	t.Run("unarmed cat", func(t *testing.T) {
		character := catWithMainHand(t, 0, &proto.ConsumesSpec{}, &proto.PartyBuffs{})

		paw := character.AutoAttacks.MH()

		if paw.BaseDamageMin != 1 || paw.BaseDamageMax != 1 {
			t.Fatalf("the installed cat paw deals %.3f-%.3f before attack power, want 1.000-1.000", paw.BaseDamageMin, paw.BaseDamageMax)
		}
		if paw.SwingSpeed != 1 {
			t.Fatalf("the installed cat paw swings every %.1f seconds, want 1.0", paw.SwingSpeed)
		}
		if got := paw.CalculateAverageWeaponDamage(attackPower); got != 11 {
			t.Fatalf("the installed cat paw averages %.3f damage at %d attack power, want 11.000", got, attackPower)
		}
	})

	t.Run("unarmed bear", func(t *testing.T) {
		character := catWithMainHand(t, 0, &proto.ConsumesSpec{}, &proto.PartyBuffs{})
		cat := character.Env.Raid.Parties[0].Players[0].(*pawDruid)

		character.AutoAttacks.SetMH(cat.GetBearWeapon())
		paw := character.AutoAttacks.MH()

		if paw.BaseDamageMin != 1 || paw.BaseDamageMax != 1 {
			t.Fatalf("the installed bear paw deals %.3f-%.3f before attack power, want 1.000-1.000", paw.BaseDamageMin, paw.BaseDamageMax)
		}
		if paw.SwingSpeed != 2.5 {
			t.Fatalf("the installed bear paw swings every %.1f seconds, want 2.5", paw.SwingSpeed)
		}
		if got := paw.CalculateAverageWeaponDamage(attackPower); got != 26 {
			t.Fatalf("the installed bear paw averages %.3f damage at %d attack power, want 26.000", got, attackPower)
		}
	})
}
