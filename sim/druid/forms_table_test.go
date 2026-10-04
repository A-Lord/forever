//go:build with_db

package druid

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func TestFeralTablePawEndpoints(t *testing.T) {
	const attackPower = 140
	for _, c := range []struct {
		name    string
		dps     float64
		catMin  float64
		catMax  float64
		bearMin float64
		bearMax float64
	}{
		{"unrounded source", 14.695897, 11, 18, 29, 44},
		{"explicit zero", 0, 1, 1, 1, 1},
		{"binary32 boundary", 12.4999995, 10, 15, 25, 38},
	} {
		for _, form := range []struct {
			name  string
			build func(*pawDruid) core.Weapon
			low   float64
			high  float64
			speed float64
		}{
			{"cat", (*pawDruid).GetCatWeapon, c.catMin, c.catMax, 1},
			{"bear", (*pawDruid).GetBearWeapon, c.bearMin, c.bearMax, 2.5},
		} {
			t.Run(c.name+"/"+form.name, func(t *testing.T) {
				// Given an equipped weapon with independent, unrounded source DPS.
				character := catWithMainHand(t, 866, &proto.ConsumesSpec{}, &proto.PartyBuffs{})
				character.GetMHWeapon().WeaponTableDPS = &c.dps
				cat := character.Env.Raid.Parties[0].Players[0].(*pawDruid)

				// When the form installs its paw.
				character.AutoAttacks.SetMH(form.build(cat))
				paw := character.AutoAttacks.MH()

				// Then rounding precedes the existing attack-power contribution.
				if paw.BaseDamageMin != form.low || paw.BaseDamageMax != form.high || paw.SwingSpeed != form.speed {
					t.Fatalf("installed paw = %.9f-%.9f at %.1f s, want %.0f-%.0f at %.1f s",
						paw.BaseDamageMin, paw.BaseDamageMax, paw.SwingSpeed, form.low, form.high, form.speed)
				}
				if got, want := paw.CalculateAverageWeaponDamage(attackPower), (form.low+form.high)/2+10*form.speed; got != want {
					t.Errorf("average damage at %d attack power = %.9f, want %.9f", attackPower, got, want)
				}
			})
		}
	}
}

func TestFeralTablePawKeepsWeaponBonusesWhenRebuilt(t *testing.T) {
	for _, form := range []struct {
		name  string
		build func(*pawDruid) core.Weapon
		low   float64
		high  float64
		speed float64
	}{
		{"cat", (*pawDruid).GetCatWeapon, 11, 18, 1},
		{"bear", (*pawDruid).GetBearWeapon, 29, 44, 2.5},
	} {
		t.Run(form.name, func(t *testing.T) {
			// Given a weaponstone, a flat enchant, and a main-hand DPS bonus.
			character := catWithMainHand(t, 866, &proto.ConsumesSpec{MhImbueId: 12643}, &proto.PartyBuffs{})
			dps := 14.695897
			character.GetMHWeapon().WeaponTableDPS = &dps
			character.GetMHWeapon().Enchant.WeaponDamage = 3
			character.PseudoStats.BonusMHDps = 2
			cat := character.Env.Raid.Parties[0].Players[0].(*pawDruid)

			// When repeated form changes rebuild the installed paw.
			for range 2 {
				character.AutoAttacks.SetMH(cat.GetBearWeapon())
				character.AutoAttacks.SetMH(cat.GetCatWeapon())
			}
			character.AutoAttacks.SetMH(form.build(cat))
			paw := character.AutoAttacks.MH()

			// Then each bonus is scaled from the equipped weapon exactly once.
			bonus := (11/2.4 + 2) * form.speed
			if !approxEqual(paw.BaseDamageMin, form.low+bonus) || !approxEqual(paw.BaseDamageMax, form.high+bonus) {
				t.Errorf("rebuilt paw = %.9f-%.9f, want %.9f-%.9f", paw.BaseDamageMin, paw.BaseDamageMax, form.low+bonus, form.high+bonus)
			}
		})
	}
}

func TestFeralTablePawUsesSuppliedCasterDPS(t *testing.T) {
	// Given classified caster metadata with an explicit source DPS value.
	character := catWithMainHand(t, 866, &proto.ConsumesSpec{}, &proto.PartyBuffs{})
	dps := 14.695897
	character.GetMHWeapon().WeaponTableDPS = &dps
	character.GetMHWeapon().QualityModifier = -1
	cat := character.Env.Raid.Parties[0].Players[0].(*pawDruid)

	// When the cat paw is built.
	paw := cat.GetCatWeapon()

	// Then classification does not remap the supplied DPS.
	if paw.BaseDamageMin != 11 || paw.BaseDamageMax != 18 {
		t.Errorf("caster paw = %.9f-%.9f, want 11-18", paw.BaseDamageMin, paw.BaseDamageMax)
	}
}
