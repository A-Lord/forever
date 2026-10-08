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
				character.GetMHWeapon().UnroundedDPS = &c.dps
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
			character.GetMHWeapon().UnroundedDPS = &dps
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

			// Then Cat adds flat enchant damage after its normalization while Bear keeps its existing scaling.
			bonus := (11/2.4 + 2) * form.speed
			if form.name == "cat" {
				bonus = 11 + 2*form.speed
			}
			if !approxEqual(paw.BaseDamageMin, form.low+bonus) || !approxEqual(paw.BaseDamageMax, form.high+bonus) {
				t.Errorf("rebuilt paw = %.9f-%.9f, want %.9f-%.9f", paw.BaseDamageMin, paw.BaseDamageMax, form.low+bonus, form.high+bonus)
			}
		})
	}
}

func TestCatPawAddsFlatWeaponDamageAfterNormalization(t *testing.T) {
	for _, tc := range []struct {
		name       string
		swingSpeed float64
		table      bool
	}{
		{"table/2.0", 2, true},
		{"table/3.5", 3.5, true},
		{"fallback/2.0", 2, false},
		{"fallback/3.5", 3.5, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			baseCharacter := catWithMainHand(t, 866, &proto.ConsumesSpec{}, &proto.PartyBuffs{})
			baseCharacter.GetMHWeapon().SwingSpeed = tc.swingSpeed
			if !tc.table {
				baseCharacter.GetMHWeapon().UnroundedDPS = nil
			}
			base := baseCharacter.Env.Raid.Parties[0].Players[0].(*pawDruid).GetCatWeapon()

			permanentCharacter := catWithMainHand(t, 866, &proto.ConsumesSpec{}, &proto.PartyBuffs{})
			permanentItem := permanentCharacter.GetMHWeapon()
			permanentItem.SwingSpeed = tc.swingSpeed
			permanentItem.Enchant.WeaponDamage = 4
			if !tc.table {
				permanentItem.UnroundedDPS = nil
			}
			permanent := permanentCharacter.Env.Raid.Parties[0].Players[0].(*pawDruid).GetCatWeapon()

			character := catWithMainHand(t, 866, &proto.ConsumesSpec{MhImbueId: 12643}, &proto.PartyBuffs{})
			item := character.GetMHWeapon()
			item.SwingSpeed = tc.swingSpeed
			item.Enchant.WeaponDamage = 4
			if !tc.table {
				item.UnroundedDPS = nil
			}
			cat := character.Env.Raid.Parties[0].Players[0].(*pawDruid)
			for range 2 {
				character.AutoAttacks.SetMH(cat.GetCatWeapon())
			}
			paw := character.AutoAttacks.MH()

			if item.TempEnchantWeaponDamage != 8 {
				t.Fatalf("Dense Weightstone damage = %v, want 8", item.TempEnchantWeaponDamage)
			}
			if want := 4.0; !approxEqual(permanent.BaseDamageMin-base.BaseDamageMin, want) || !approxEqual(permanent.BaseDamageMax-base.BaseDamageMax, want) {
				t.Errorf("Cat permanent weapon damage = %.3f-%.3f, want %.3f on both endpoints", permanent.BaseDamageMin-base.BaseDamageMin, permanent.BaseDamageMax-base.BaseDamageMax, want)
			}
			if want := 12.0; !approxEqual(paw.BaseDamageMin-base.BaseDamageMin, want) || !approxEqual(paw.BaseDamageMax-base.BaseDamageMax, want) {
				t.Errorf("Cat flat weapon damage = %.3f-%.3f, want %.3f on both endpoints", paw.BaseDamageMin-base.BaseDamageMin, paw.BaseDamageMax-base.BaseDamageMax, want)
			}
			t.Logf("Cat %s endpoints %.3f-%.3f become %.3f-%.3f with permanent +4 and %.3f-%.3f with permanent +4 plus Dense Weightstone +8", tc.name, base.BaseDamageMin, base.BaseDamageMax, permanent.BaseDamageMin, permanent.BaseDamageMax, paw.BaseDamageMin, paw.BaseDamageMax)
		})
	}
}

func TestFeralTablePawUsesSuppliedCasterDPS(t *testing.T) {
	// Given classified caster metadata with an explicit source DPS value.
	character := catWithMainHand(t, 866, &proto.ConsumesSpec{}, &proto.PartyBuffs{})
	dps := 14.695897
	character.GetMHWeapon().UnroundedDPS = &dps
	character.GetMHWeapon().QualityModifier = -1
	cat := character.Env.Raid.Parties[0].Players[0].(*pawDruid)

	// When the cat paw is built.
	paw := cat.GetCatWeapon()

	// Then classification does not remap the supplied DPS.
	if paw.BaseDamageMin != 11 || paw.BaseDamageMax != 18 {
		t.Errorf("caster paw = %.9f-%.9f, want 11-18", paw.BaseDamageMin, paw.BaseDamageMax)
	}
}
