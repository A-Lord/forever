package druid

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	_ "github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

const (
	monksStaffID       int32 = 866
	denseWeightstoneID int32 = 16622
)

func init() {
	core.RegisterAgentFactory(
		proto.Player_FeralCatDruid{},
		proto.Spec_SpecFeralCatDruid,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			return newFakeCatDruid(character)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_FeralCatDruid)
			if !ok {
				panic("Invalid spec value for Feral Druid!")
			}
			player.Spec = playerSpec
		},
	)
}

type fakeCatDruid struct {
	*Druid
}

func newFakeCatDruid(character *core.Character) *fakeCatDruid {
	cat := &fakeCatDruid{Druid: New(character, Cat, SelfBuffs{}, "")}
	cat.EnableAutoAttacks(cat, core.AutoAttackOptions{
		MainHand:       cat.GetCatWeapon(),
		AutoSwingMelee: true,
	})
	return cat
}

func (cat *fakeCatDruid) Initialize()                         {}
func (cat *fakeCatDruid) ApplyTalents()                       {}
func (cat *fakeCatDruid) Reset(_ *core.Simulation)            {}
func (cat *fakeCatDruid) OnGCDReady(_ *core.Simulation)       {}
func (cat *fakeCatDruid) OnEncounterStart(_ *core.Simulation) {}

func setupCatDruid(mhImbueId int32) *fakeCatDruid {
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: &proto.Raid{
			Parties: []*proto.Party{
				{
					Players: []*proto.Player{
						{
							Name:        "Druid",
							Class:       proto.Class_ClassDruid,
							Race:        proto.Race_RaceTauren,
							Buffs:       &proto.IndividualBuffs{},
							Consumables: &proto.ConsumesSpec{MhImbueId: mhImbueId},
							Spec:        &proto.Player_FeralCatDruid{},
							Equipment:   &proto.EquipmentSpec{Items: []*proto.ItemSpec{{Id: monksStaffID}}},
							Rotation:    &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
						},
					},
					Buffs: &proto.PartyBuffs{},
				},
			},
			Buffs: &proto.RaidBuffs{},
		},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 63}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim.Raid.Parties[0].Players[0].(*fakeCatDruid)
}

func checkFormWeapon(t *testing.T, label string, got core.Weapon, wantMin, wantMax, wantSpeed float64) {
	t.Helper()
	if !core.WithinToleranceFloat64(wantMin, got.BaseDamageMin, 1e-9) ||
		!core.WithinToleranceFloat64(wantMax, got.BaseDamageMax, 1e-9) ||
		got.SwingSpeed != wantSpeed {
		t.Fatalf("%s: expected %0.4f - %0.4f at %0.1f speed, got %0.4f - %0.4f at %0.1f speed",
			label, wantMin, wantMax, wantSpeed, got.BaseDamageMin, got.BaseDamageMax, got.SwingSpeed)
	}
}

func TestFormWeaponsCarryTheMainHandImbueFlatDamage(t *testing.T) {
	if core.GetItemByID(monksStaffID) == nil {
		t.Skip("no item database loaded; run with -tags with_db")
	}

	cat := setupCatDruid(denseWeightstoneID)

	staff := cat.WeaponFromMainHand()
	if staff.SwingSpeed != 2.4 {
		t.Fatalf("Monk's Staff should swing at 2.4, got %0.2f", staff.SwingSpeed)
	}
	catMin := (staff.BaseDamageMin + 8) / staff.SwingSpeed
	catMax := (staff.BaseDamageMax + 8) / staff.SwingSpeed

	checkFormWeapon(t, "cat weapon", cat.GetCatWeapon(), catMin, catMax, 1.0)
	checkFormWeapon(t, "bear weapon", cat.GetBearWeapon(), catMin*2.5, catMax*2.5, 2.5)

	cat.AutoAttacks.SetMH(cat.GetBearWeapon())
	checkFormWeapon(t, "main hand after shifting to bear", *cat.AutoAttacks.MH(), catMin*2.5, catMax*2.5, 2.5)
	cat.AutoAttacks.SetMH(cat.GetCatWeapon())
	checkFormWeapon(t, "main hand after shifting back to cat", *cat.AutoAttacks.MH(), catMin, catMax, 1.0)

	plain := setupCatDruid(0)
	checkFormWeapon(t, "cat weapon without an imbue", plain.GetCatWeapon(),
		staff.BaseDamageMin/staff.SwingSpeed, staff.BaseDamageMax/staff.SwingSpeed, 1.0)
}
