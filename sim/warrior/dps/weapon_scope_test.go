package dps

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func orcWarrior(t *testing.T, gearSet string, talents string) *DpsWarrior {
	t.Helper()
	player := &proto.Player{
		Name:          "Weapon Scope",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", gearSet).GearSet,
		TalentsString: talents,
		Spec:          DefaultOptions,
		Rotation:      core.GetAplRotation("../../../ui/specs/warrior/dps/apls", "fury").Rotation,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].(*DpsWarrior)
}

// Dual Wield Specialization 5/5 gives the off-hand axe's attacks 10% hit and the main hand's none.
// Cruelty 5/5 gives each axe 5% crit. Axe Specialization is crit on every attack, not a mod on the axes.
func TestFuryTalentsReachTheirWeapons(t *testing.T) {
	warrior := orcWarrior(t, "p1_fury", DefaultFuryTalents)
	mh, oh := warrior.AutoAttacks.MHAuto(), warrior.AutoAttacks.OHAuto()

	if mh.BonusHitPercent != 0 || oh.BonusHitPercent != 10 {
		t.Errorf("main-hand and off-hand hit = %0.2f%% and %0.2f%%, want 0%% and 10%%", mh.BonusHitPercent, oh.BonusHitPercent)
	}
	if mh.BonusCritPercent != 5 || oh.BonusCritPercent != 5 {
		t.Errorf("main-hand and off-hand crit = %0.2f%% and %0.2f%%, want 5%% each", mh.BonusCritPercent, oh.BonusCritPercent)
	}
}

// A two-handed axe takes Cruelty's 5% and Weaponmaster's 5% crit, and Weaponmaster's aura is up.
func TestArmsTalentsReachTheTwoHandedAxe(t *testing.T) {
	warrior := orcWarrior(t, "p1_arms", DefaultArmsTalents)

	if crit := warrior.AutoAttacks.MHAuto().BonusCritPercent; crit != 10 {
		t.Errorf("two-handed axe crit = %0.2f%%, want 10%%", crit)
	}
	if aura := warrior.GetAura("Weaponmaster (Axe/Polearm)"); aura == nil || !aura.IsActive() {
		t.Errorf("Weaponmaster (Axe/Polearm) is not up with a two-handed axe")
	}
}

// Thunder Clap is a magic-defense spell on the ranged proc mask: a bow's hit bonus reaches a ranged
// attack made with the bow and not the clap, and an axe's crit bonus still reaches Heroic Strike. A
// warrior swings no ranged auto attack, so the test registers one.
func TestWeaponBoundModsSkipThunderClap(t *testing.T) {
	warrior := orcWarrior(t, "p1_fury", DefaultFuryTalents)
	clap := warrior.GetSpell(core.ActionID{SpellID: 11581})
	strike := warrior.GetSpell(core.ActionID{SpellID: 25286})
	rangedAuto := warrior.RegisterSpell(core.SpellConfig{ActionID: core.ActionID{SpellID: 990970},
		SpellSchool: core.SpellSchoolPhysical, ProcMask: core.ProcMaskRangedAuto, DefenseType: core.DefenseTypeRanged})
	if clap == nil || strike == nil || warrior.Ranged() == nil {
		t.Fatalf("clap %v, Heroic Strike %v, ranged weapon %v; want all three", clap, strike, warrior.Ranged())
	}
	if clap.DefenseType != core.DefenseTypeMagic || clap.ProcMask != core.ProcMaskRangedSpecial {
		t.Fatalf("clap has defense %v and proc mask %v, want magic and ranged special", clap.DefenseType, clap.ProcMask)
	}

	clapHit, rangedHit, strikeCrit := clap.BonusHitPercent, rangedAuto.BonusHitPercent, strike.BonusCritPercent
	probe := warrior.ScopedPercents(core.ProcMaskRangedAuto)
	warrior.AddStaticMod(core.SpellModConfig{Kind: core.SpellMod_BonusHit_Percent, FloatValue: 3,
		ProcMask: core.ProcMaskRanged, EquipSubclass: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING, FromRating: true})
	warrior.AddStaticMod(core.SpellModConfig{Kind: core.SpellMod_BonusCrit_Percent, FloatValue: 4,
		ProcMask: core.ProcMaskMelee, EquipSubclass: core.WeaponSubclassMask(proto.WeaponType_WeaponTypeAxe)})

	for _, c := range []struct {
		name      string
		got, want float64
	}{
		{"Thunder Clap hit", clap.BonusHitPercent - clapHit, 0},
		{"ranged attack hit", rangedAuto.BonusHitPercent - rangedHit, 3},
		{"Heroic Strike crit", strike.BonusCritPercent - strikeCrit, 4},
		{"probed ranged hit", warrior.ScopedPercents(core.ProcMaskRangedAuto).Hit - probe.Hit, 3},
		{"probed ranged hit from rating", warrior.ScopedPercents(core.ProcMaskRangedAuto).HitFromRating - probe.HitFromRating, 3},
	} {
		if c.got != c.want {
			t.Errorf("%s moved by %v, want %v", c.name, c.got, c.want)
		}
	}
}
