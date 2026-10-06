package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	sableWandID    int32 = 7607   // 21–40 shadow, 1.80 s.
	physicalWandID int32 = 279847 // Unstable Power Core, physical, 1.70 s.
)

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
	return newSwapWandSim(t, ranged, nil, rotation, distance)
}

// newWandSim with a swap set that holds the given ranged slot, when it is not nil.
func newSwapWandSim(t *testing.T, ranged, swapped *proto.ItemSpec, rotation *proto.APLRotation, distance float64) (*core.Simulation, *wandCaster) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotRanged] = ranged
	var swapItems []*proto.ItemSpec
	if swapped != nil {
		swapItems = make([]*proto.ItemSpec, len(items))
		for i := range swapItems {
			swapItems[i] = &proto.ItemSpec{}
		}
		swapItems[proto.ItemSlot_ItemSlotRanged] = swapped
	}

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
				EnableItemSwap:     swapped != nil,
				ItemSwap:           &proto.ItemSwap{Items: swapItems},
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

func landed(metrics *core.SpellMetrics) int32 {
	return metrics.Hits + metrics.Crits
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

	t.Run("a shot is never partially resisted", func(t *testing.T) {
		sim, caster := newSableWandSim(t)
		plain := shots(t, sim, caster, 200)

		sim, caster = newSableWandSim(t)
		sim.Encounter.ActiveTargetUnits[0].AddStatsDynamic(sim, stats.Stats{stats.ShadowResistance: 150})
		resisted := shots(t, sim, caster, 200)
		for i := range resisted {
			if resisted[i] != plain[i] {
				t.Fatalf("against 150 shadow resistance shot %d dealt %v, want the same %v as against none", i+1, resisted[i], plain[i])
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

	t.Run("a first shot fires after the wind-up and flies at 20 yards a second", func(t *testing.T) {
		sim, caster := newWandSim(t, &proto.ItemSpec{Id: sableWandID}, &proto.APLRotation{}, 20)
		target := sim.Encounter.ActiveTargetUnits[0]
		metrics := &caster.shoot.SpellMetrics[target.UnitIndex]
		if !caster.shoot.Cast(sim, target) {
			t.Fatalf("Shoot could not be cast from 20 yards")
		}

		stepPast(t, sim, 499*time.Millisecond)
		if landed(metrics) != 0 {
			t.Fatalf("Shoot fired before its 0.5 s wind-up ended")
		}
		stepPast(t, sim, 501*time.Millisecond)
		if landed(metrics) != 1 || metrics.TotalDamage != 0 {
			t.Fatalf("at 0.5 s: %d shots, %v damage; want 1 shot still in flight", landed(metrics), metrics.TotalDamage)
		}
		stepPast(t, sim, 1499*time.Millisecond)
		if metrics.TotalDamage != 0 {
			t.Fatalf("the shot landed %v before its 1 s flight ended", metrics.TotalDamage)
		}
		stepPast(t, sim, 1501*time.Millisecond)
		if metrics.TotalDamage == 0 {
			t.Fatalf("the shot had dealt nothing at 1.50 s")
		}
	})

	t.Run("a caster registered without a wand reads zero and cannot shoot", func(t *testing.T) {
		sim, caster := newWandSim(t, &proto.ItemSpec{}, &proto.APLRotation{}, 0)
		target := sim.Encounter.ActiveTargetUnits[0]
		if caster.Wand == nil {
			t.Fatalf("the caster has no Wand state")
		}
		if caster.Wand.SwingSpeed() != 0 || caster.Wand.TimeToNext(sim) != 0 || caster.Wand.TimeSinceLast(sim) != sim.CurrentTime {
			t.Fatalf("with no wand the values read %v swing, %v to the next, %v since the last", caster.Wand.SwingSpeed(), caster.Wand.TimeToNext(sim), caster.Wand.TimeSinceLast(sim))
		}
		if caster.shoot.CanCast(sim, target) {
			t.Fatalf("Shoot can be cast with no wand")
		}
		caster.shoot.Cast(sim, target)
		if caster.ChanneledDot != nil {
			t.Fatalf("Shoot is channeling with no wand")
		}
	})

	t.Run("a channel from the pull fires after the wind-up and then every wand speed", func(t *testing.T) {
		sim, caster := newWandSim(t, &proto.ItemSpec{Id: sableWandID}, channelShootRotation(), 0)
		metrics := &caster.shoot.SpellMetrics[sim.Encounter.ActiveTargetUnits[0].UnitIndex]
		sim.PrePull()

		stepPast(t, sim, 499*time.Millisecond)
		if !caster.wanding() || metrics.TotalDamage != 0 {
			t.Fatalf("at 0.499 s the pull's Shoot should be channeling with nothing dealt, wanding %t, %v dealt", caster.wanding(), metrics.TotalDamage)
		}
		stepPast(t, sim, 501*time.Millisecond)
		if landed(metrics) != 1 || metrics.TotalDamage == 0 {
			t.Fatalf("at 0.5 s the pull's Shoot should have landed its first shot, %d shots for %v", landed(metrics), metrics.TotalDamage)
		}
		stepPast(t, sim, 2299*time.Millisecond)
		if landed(metrics) != 1 {
			t.Fatalf("the second shot should wait a wand speed after the first, %d shots at 2.299 s", landed(metrics))
		}
		stepPast(t, sim, 2301*time.Millisecond)
		if metrics.Casts != 1 || landed(metrics) != 2 {
			t.Fatalf("at 2.30 s the pull's channel should have fired its second shot from one cast, %d casts, %d shots", metrics.Casts, landed(metrics))
		}
	})

	t.Run("cast speed at the start sets the shot interval", func(t *testing.T) {
		sim, caster := newWandSim(t, &proto.ItemSpec{Id: sableWandID}, channelShootRotation(), 0)
		caster.MultiplyCastSpeed(sim, 1.1)
		metrics := &caster.shoot.SpellMetrics[sim.Encounter.ActiveTargetUnits[0].UnitIndex]
		sim.PrePull()

		for _, step := range []struct {
			at    time.Duration
			shots int32
		}{{501 * time.Millisecond, 1}, {2135 * time.Millisecond, 1}, {2137 * time.Millisecond, 2}, {3773 * time.Millisecond, 3}} {
			stepPast(t, sim, step.at)
			if got := landed(metrics); got != step.shots {
				t.Fatalf("with 10%% cast speed from the pull Shoot should fire at 0.5 s and then every 1.636 s, %d shots at %v, want %d", got, step.at, step.shots)
			}
		}
	})

	t.Run("moving ends Shoot and keeps it from starting", func(t *testing.T) {
		sim, caster := newWandSim(t, &proto.ItemSpec{Id: sableWandID}, channelShootRotation(), 0)
		target := sim.Encounter.ActiveTargetUnits[0]
		sim.PrePull()

		stepPast(t, sim, time.Second)
		caster.MoveDuration(2*time.Second, sim)
		if caster.wanding() || caster.shoot.CanCast(sim, target) {
			t.Fatalf("moving at 1 s should end Shoot and keep it from starting, wanding %t", caster.wanding())
		}
	})

	t.Run("Wand Specialization's school mask leaves out a physical wand", func(t *testing.T) {
		for _, wand := range []struct {
			id         int32
			multiplier float64
		}{{sableWandID, 1.25}, {physicalWandID, 1}} {
			sim, caster := newWandSim(t, &proto.ItemSpec{Id: wand.id}, &proto.APLRotation{}, 0)
			plain := shots(t, sim, caster, 20)

			sim, caster = newWandSim(t, &proto.ItemSpec{Id: wand.id}, &proto.APLRotation{}, 0)
			caster.AddStaticMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Pct,
				ProcMask:   core.ProcMaskRangedAuto,
				School:     core.SpellSchool(126),
				FloatValue: 0.25,
			})
			specialized := shots(t, sim, caster, 20)
			for i := range specialized {
				if diff := specialized[i] - wand.multiplier*plain[i]; diff > 1e-9 || diff < -1e-9 {
					t.Fatalf("wand %d (%v) shot %d dealt %v with the talent's mod, want %v × %v", wand.id, caster.shoot.SpellSchool, i+1, specialized[i], wand.multiplier, plain[i])
				}
			}
		}
	})
}

func TestWandShootReadsEquippedWand(t *testing.T) {
	// Swaps the second ranged slot in and casts Shoot once, then checks the shot against the wand now equipped.
	checkSwappedShoot := func(t *testing.T, ranged, swapped *proto.ItemSpec, school core.SpellSchool, speed time.Duration) {
		t.Helper()
		sim, caster := newSwapWandSim(t, ranged, swapped, &proto.APLRotation{}, 0)
		caster.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
		wand := caster.WeaponFromRanged()
		target := sim.Encounter.ActiveTargetUnits[0]
		metrics := &caster.shoot.SpellMetrics[target.UnitIndex]
		stepPast(t, sim, 5*time.Second)

		if !caster.shoot.Cast(sim, target) {
			t.Fatalf("Shoot could not be cast with the swapped-in wand")
		}
		if caster.shoot.SpellSchool != school {
			t.Fatalf("Shoot is %v, want the swapped-in wand's %v", caster.shoot.SpellSchool, school)
		}
		if got := caster.Wand.SwingSpeed(); got != speed {
			t.Fatalf("the wand swing time reads %s, want the swapped-in wand's %s", got, speed)
		}
		stepPast(t, sim, sim.CurrentTime+speed+time.Millisecond)
		if landed(metrics) != 1 {
			t.Fatalf("Shoot fired %d shots within one wand speed, want 1", landed(metrics))
		}
		if metrics.TotalDamage < wand.BaseDamageMin || metrics.TotalDamage > wand.BaseDamageMax {
			t.Fatalf("the shot dealt %v, want %v to %v", metrics.TotalDamage, wand.BaseDamageMin, wand.BaseDamageMax)
		}
	}

	t.Run("a wand swapped for another shoots at the new wand's speed, damage and school", func(t *testing.T) {
		checkSwappedShoot(t, &proto.ItemSpec{Id: sableWandID}, &proto.ItemSpec{Id: physicalWandID}, core.SpellSchoolPhysical, 1700*time.Millisecond)
	})

	t.Run("a wand swapped in after registering with none shoots at its speed, damage and school", func(t *testing.T) {
		checkSwappedShoot(t, &proto.ItemSpec{}, &proto.ItemSpec{Id: sableWandID}, core.SpellSchoolShadow, 1800*time.Millisecond)
	})
}

func (c *wandCaster) wanding() bool {
	return c.ChanneledDot != nil && c.ChanneledDot.Spell == c.shoot
}

// Shoot as a channel that any other ready action interrupts.
func channelShootRotation() *proto.APLRotation {
	return &proto.APLRotation{
		Type: proto.APLRotation_TypeAPL,
		PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{Action: &proto.APLAction_ChannelSpell{ChannelSpell: &proto.APLActionChannelSpell{
			SpellId:     core.ActionID{SpellID: 5019}.ToProto(),
			InterruptIf: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "true"}}},
		}}}}},
	}
}
