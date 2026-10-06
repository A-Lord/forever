package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

const sableWandID int32 = 7607 // 21–40 shadow, 1.80 s.

func init() {
	core.RegisterAgentFactory(
		proto.Player_DpsPriest{},
		proto.Spec_SpecDpsPriest,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			return &wandCaster{testCaster: testCaster{Character: *character}}
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_DpsPriest)
		},
	)
}

// Registers Shoot where the classes do, while the character is built.
type wandCaster struct {
	testCaster
	shoot *core.Spell
}

func (c *wandCaster) Initialize() {
	c.shoot = RegisterShoot(&c.Character)
}

// A priest with the given ranged slot and rotation, at the given distance, with its spells certain to
// hit and never to crit.
func newWandSim(t *testing.T, ranged *proto.ItemSpec, rotation *proto.APLRotation, distance float64) (*core.Simulation, *wandCaster) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotRanged] = ranged

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, IsTest: true},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name:               "Wander",
				Class:              proto.Class_ClassPriest,
				Race:               proto.Race_RaceHuman,
				Buffs:              &proto.IndividualBuffs{},
				Spec:               &proto.Player_DpsPriest{},
				Equipment:          &proto.EquipmentSpec{Items: items},
				Rotation:           rotation,
				DistanceFromTarget: distance,
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60, MobType: proto.MobType_MobTypeDemon}},
			Duration: 600,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	caster := sim.Raid.Parties[0].Players[0].(*wandCaster)
	caster.AddStatsDynamic(sim, stats.Stats{
		stats.SpellHitPercent:  100,
		stats.SpellCritPercent: -caster.GetStat(stats.SpellCritPercent),
	})
	return sim, caster
}

func newSableWandSim(t *testing.T) (*core.Simulation, *wandCaster) {
	t.Helper()
	return newWandSim(t, &proto.ItemSpec{Id: sableWandID}, &proto.APLRotation{}, 0)
}

// Casts Shoot, whose plain cast fires one shot, n times in a row and returns what each shot dealt.
func shots(t *testing.T, sim *core.Simulation, caster *wandCaster, n int) []float64 {
	t.Helper()
	target := sim.Encounter.ActiveTargetUnits[0]
	metrics := &caster.shoot.SpellMetrics[target.UnitIndex]
	dealt := make([]float64, n)
	for i := range dealt {
		before := metrics.TotalDamage
		start := sim.CurrentTime
		if !caster.shoot.Cast(sim, target) {
			t.Fatalf("Shoot %d could not be cast at %v", i+1, start)
		}
		stepPast(t, sim, start+1800*time.Millisecond+time.Millisecond)
		dealt[i] = metrics.TotalDamage - before
	}
	return dealt
}

func TestWandShoot(t *testing.T) {
	t.Run("a shot rolls the wand's damage in its school, with no power", func(t *testing.T) {
		sim, caster := newSableWandSim(t)
		if caster.shoot.SpellSchool != core.SpellSchoolShadow {
			t.Fatalf("Shoot with the Sable Wand is %v, want shadow", caster.shoot.SpellSchool)
		}
		plain := shots(t, sim, caster, 200)
		for i, damage := range plain {
			if damage < 21 || damage > 40 {
				t.Fatalf("shot %d dealt %v, want 21 to 40", i+1, damage)
			}
		}

		sim, caster = newSableWandSim(t)
		caster.AddStatsDynamic(sim, stats.Stats{stats.SpellDamage: 500, stats.AttackPower: 500, stats.RangedAttackPower: 500})
		powered := shots(t, sim, caster, 200)
		for i := range powered {
			if powered[i] != plain[i] {
				t.Fatalf("with 500 spell power and attack power shot %d dealt %v, want the same %v as without", i+1, powered[i], plain[i])
			}
		}
	})

	t.Run("a shot crits for 1.5 times its roll", func(t *testing.T) {
		sim, caster := newSableWandSim(t)
		if caster.shoot.DefenseType != core.DefenseTypeMagic {
			t.Fatalf("Shoot's defense type is %v, want magic", caster.shoot.DefenseType)
		}
		plain := shots(t, sim, caster, 50)

		sim, caster = newSableWandSim(t)
		caster.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: 100})
		crits := shots(t, sim, caster, 50)
		for i := range crits {
			if diff := crits[i] - 1.5*plain[i]; diff > 1e-9 || diff < -1e-9 {
				t.Fatalf("shot %d crit for %v, want 1.5 × %v", i+1, crits[i], plain[i])
			}
		}
		metrics := caster.shoot.SpellMetrics[sim.Encounter.ActiveTargetUnits[0].UnitIndex]
		if metrics.Crits != 50 || metrics.Hits != 0 {
			t.Fatalf("with 100%% spell crit Shoot landed %d crits and %d hits, want 50 and 0", metrics.Crits, metrics.Hits)
		}
	})

	t.Run("a shot fires after the wand speed and flies at 20 yards a second", func(t *testing.T) {
		sim, caster := newWandSim(t, &proto.ItemSpec{Id: sableWandID}, &proto.APLRotation{}, 20)
		target := sim.Encounter.ActiveTargetUnits[0]
		metrics := &caster.shoot.SpellMetrics[target.UnitIndex]
		if !caster.shoot.Cast(sim, target) {
			t.Fatalf("Shoot could not be cast from 20 yards")
		}

		stepPast(t, sim, 1799*time.Millisecond)
		if metrics.Hits+metrics.Crits != 0 {
			t.Fatalf("Shoot fired before 1.80 s")
		}
		stepPast(t, sim, 1801*time.Millisecond)
		if metrics.Hits+metrics.Crits != 1 || metrics.TotalDamage != 0 {
			t.Fatalf("at 1.80 s: %d shots, %v damage; want 1 shot still in flight", metrics.Hits+metrics.Crits, metrics.TotalDamage)
		}
		stepPast(t, sim, 2799*time.Millisecond)
		if metrics.TotalDamage != 0 {
			t.Fatalf("the shot landed %v before its 1 s flight ended", metrics.TotalDamage)
		}
		stepPast(t, sim, 2801*time.Millisecond)
		if metrics.TotalDamage == 0 {
			t.Fatalf("the shot had dealt nothing at 2.80 s")
		}
	})

	t.Run("no wand, no Shoot", func(t *testing.T) {
		sim, caster := newWandSim(t, &proto.ItemSpec{}, &proto.APLRotation{}, 0)
		target := sim.Encounter.ActiveTargetUnits[0]
		if caster.shoot.CanCast(sim, target) {
			t.Fatalf("Shoot can be cast with no wand")
		}
		caster.shoot.Cast(sim, target)
		if caster.Wand.IsWanding() {
			t.Fatalf("Shoot is channeling with no wand")
		}
	})

	t.Run("a channel from the pull fires its first shot after the wand speed", func(t *testing.T) {
		rotation := &proto.APLRotation{
			Type: proto.APLRotation_TypeAPL,
			PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{Action: &proto.APLAction_ChannelSpell{ChannelSpell: &proto.APLActionChannelSpell{
				SpellId:     core.ActionID{SpellID: 5019}.ToProto(),
				InterruptIf: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "true"}}},
			}}}}},
		}
		sim, caster := newWandSim(t, &proto.ItemSpec{Id: sableWandID}, rotation, 0)
		metrics := &caster.shoot.SpellMetrics[sim.Encounter.ActiveTargetUnits[0].UnitIndex]
		sim.PrePull()

		stepPast(t, sim, 1799*time.Millisecond)
		if !caster.Wand.IsWanding() || metrics.TotalDamage != 0 {
			t.Fatalf("at 1.799 s the pull's Shoot should be channeling with nothing dealt, wanding %t, %v dealt", caster.Wand.IsWanding(), metrics.TotalDamage)
		}
		stepPast(t, sim, 1801*time.Millisecond)
		if metrics.Hits+metrics.Crits != 1 || metrics.TotalDamage == 0 {
			t.Fatalf("at 1.80 s the pull's Shoot should have landed its first shot, %d shots for %v", metrics.Hits+metrics.Crits, metrics.TotalDamage)
		}
		stepPast(t, sim, 3601*time.Millisecond)
		if metrics.Casts != 1 || metrics.Hits+metrics.Crits != 2 {
			t.Fatalf("at 3.60 s the pull's channel should have fired its second shot from one cast, %d casts, %d shots", metrics.Casts, metrics.Hits+metrics.Crits)
		}
	})
}
