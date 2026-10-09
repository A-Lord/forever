//go:build with_db

package feralcat

import (
	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
	"os"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func spellReady(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellIsReady{SpellIsReady: &proto.APLValueSpellIsReady{
		SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}},
	}}}
}

func currentEnergy() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentEnergy{CurrentEnergy: &proto.APLValueCurrentEnergy{}}}
}

// withFeralDigit sets one feral-tree digit. Index 10 is Shifting Power, 11 is Primal Bite.
func withFeralDigit(talents string, index int, digit byte) string {
	parts := []byte(talents)
	start := 0
	for i, c := range parts {
		if c == '-' {
			start = i + 1
			break
		}
	}
	parts[start+index] = digit
	return string(parts)
}

// A 180-second fight against a level-63 demon with the cat UI's default raid buffs and debuffs,
// the Level 60 gear preset (ui/specs/druid/feralcat/gear_sets/default.gear.json), and no consumables.
// Weapon is Fist of Cenarius. Howling Idol shortens Shifting Power by 1s.
func sweepRequest(talents string, rotation *proto.APLRotation, cannotShred bool, iterations int32) *proto.RaidSimRequest {
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 101, Iterations: iterations, IsTest: true},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
					TalentsString:  talents,
					ReactionTimeMs: 250,
					Equipment:      core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default").GearSet,
					Consumables:    &proto.ConsumesSpec{},
					Buffs: &proto.IndividualBuffs{
						GreaterBlessingOfKings: true,
						GreaterBlessingOfMight: true,
					},
					Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
						Options: &proto.FeralCatDruid_Options{CannotShredTarget: cannotShred},
					}},
					Rotation: rotation,
				}},
				Buffs: &proto.PartyBuffs{
					BattleShout:          proto.TristateEffect_TristateEffectRegular,
					WindfuryTotem:        true,
					ManaSpringTotem:      proto.TristateEffect_TristateEffectRegular,
					StrengthOfEarthTotem: true,
				},
			}},
			Buffs: &proto.RaidBuffs{
				ArcaneBrilliance:         true,
				PrayerOfSpirit:           true,
				GiftOfTheWild:            true,
				PrayerOfFortitude:        true,
				PrayerOfShadowProtection: true,
			},
			Debuffs: &proto.Debuffs{
				ExposeArmor:         true,
				HuntersMark:         true,
				JudgementOfWisdom:   true,
				CurseOfRecklessness: true,
				FaerieFire:          true,
				SunderArmor:         true,
				GiftOfArthas:        true,
			},
		},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{
			Name: "target", Level: 63, MobType: proto.MobType_MobTypeDemon,
		}}, Duration: 180},
	}
}

func newCatWearingTargets(t *testing.T, talents string, consumes *proto.ConsumesSpec, worn map[proto.ItemSlot]int32, targetCount int) (*core.Simulation, *FeralDruid) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	for slot, id := range worn {
		items[slot].Id = id
	}
	if consumes == nil {
		consumes = &proto.ConsumesSpec{}
	}
	targets := make([]*proto.Target, targetCount)
	for i := range targets {
		targets[i] = &proto.Target{Name: "target", Level: 60}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: talents,
				Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
					Options: &proto.FeralCatDruid_Options{},
				}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: consumes,
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: targets, Duration: 60},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*FeralDruid)
}

func buildRequest(t testing.TB, name string) *proto.RaidSimRequest {
	t.Helper()
	data, err := os.ReadFile("../../../ui/specs/druid/feralcat/builds/" + name + ".build.json")
	if err != nil {
		t.Fatal(err)
	}
	settings := &proto.IndividualSimSettings{}
	if err := protojson.Unmarshal(data, settings); err != nil {
		t.Fatal(err)
	}
	rotation := "default"
	switch name {
	case "starter":
		rotation = "starter"
	case "raid":
		rotation = "raid-rip4"
	}
	settings.Player.Rotation = core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", rotation).Rotation
	raid := core.SinglePlayerRaidProto(settings.Player, settings.PartyBuffs, settings.RaidBuffs, settings.Debuffs)
	raid.Tanks = settings.Tanks
	if settings.Encounter == nil {
		settings.Encounter = &proto.Encounter{Duration: core.LongDuration, Targets: []*proto.Target{core.FreshDefaultTargetConfig()}}
	}
	return &proto.RaidSimRequest{Raid: raid, Encounter: settings.Encounter, SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, IsTest: true}}
}

func starterRequest(t testing.TB) *proto.RaidSimRequest {
	return buildRequest(t, "starter")
}

func TestCatBuildsUseNativeRaidTargetArmor(t *testing.T) {
	for _, name := range []string{"starter", "raid"} {
		t.Run(name, func(t *testing.T) {
			request := buildRequest(t, name)
			target := request.Encounter.Targets[0]
			defaultTarget := core.FreshDefaultTargetConfig()
			if target.Level != core.CharacterLevel+3 || target.Stats[stats.Armor] != defaultTarget.Stats[stats.Armor] || target.Stats[stats.Health] != defaultTarget.Stats[stats.Health] {
				t.Fatalf("Cat build overrides native raid target level, armor, or health: %v", target)
			}
			armored := core.RunRaidSim(request)
			unarmoredRequest := goproto.Clone(request).(*proto.RaidSimRequest)
			unarmoredRequest.Encounter.Targets[0].Stats[stats.Armor] = 0
			unarmored := core.RunRaidSim(unarmoredRequest)
			if armored.Error != nil || unarmored.Error != nil {
				t.Fatalf("armor comparison failed: %v / %v", armored.Error, unarmored.Error)
			}
			if armored.RaidMetrics.Dps.Avg >= unarmored.RaidMetrics.Dps.Avg {
				t.Fatalf("native armor did not reduce Cat damage: armored %.3f, unarmored %.3f", armored.RaidMetrics.Dps.Avg, unarmored.RaidMetrics.Dps.Avg)
			}
		})
	}
}

func TestCatBuildGearMatchesFaction(t *testing.T) {
	items := make(map[int32]*proto.UIItem)
	for _, item := range database.Load().Items {
		items[item.Id] = item
	}
	for _, name := range []string{"starter", "raid"} {
		t.Run(name, func(t *testing.T) {
			player := buildRequest(t, name).Raid.Parties[0].Players[0]
			for _, equipped := range player.Equipment.Items {
				if equipped.Id == 0 {
					continue
				}
				item := items[equipped.Id]
				if item == nil {
					t.Fatalf("equipped item %d is absent from the embedded database", equipped.Id)
				}
				if player.Race == proto.Race_RaceNightElf && item.FactionRestriction == proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY {
					t.Errorf("Night Elf build equips Horde-only item %d %s", item.Id, item.Name)
				}
				if player.Race == proto.Race_RaceTauren && item.FactionRestriction == proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY {
					t.Errorf("Tauren build equips Alliance-only item %d %s", item.Id, item.Name)
				}
			}
		})
	}
}

func TestRaidGearFixtureUsesEquippableDruidItems(t *testing.T) {
	data, err := os.ReadFile("../../../ui/specs/druid/feralcat/builds/raid.build.json")
	if err != nil {
		t.Fatal(err)
	}
	settings := &proto.IndividualSimSettings{}
	if err := protojson.Unmarshal(data, settings); err != nil {
		t.Fatal(err)
	}
	gear := settings.GetPlayer().GetEquipment()
	if gear == nil {
		t.Fatal("raid build has no equipment")
	}
	if len(gear.Items) != int(proto.ItemSlot_ItemSlotRanged)+1 {
		t.Fatalf("gear has %d slots, want %d", len(gear.Items), int(proto.ItemSlot_ItemSlotRanged)+1)
	}
	for slot, spec := range gear.Items {
		if spec.Id == 0 {
			continue
		}
		item := core.GetItemByID(spec.Id)
		if item == nil {
			t.Errorf("slot %d item %d is absent from the generated database", slot, spec.Id)
			continue
		}
		if !item.UsableBy(proto.Class_ClassDruid) {
			t.Errorf("slot %d item %d is not usable by Druids", slot, spec.Id)
		}
	}
	worn := make(map[proto.ItemSlot]int32)
	for _, id := range []int32{8345, 272427} {
		found := false
		for slot, item := range gear.Items {
			found = found || item.Id == id
			if item.Id != 0 {
				worn[proto.ItemSlot(slot)] = item.Id
			}
		}
		if !found {
			t.Errorf("supported Cat item %d is absent", id)
		}
	}
	if core.GetItemByID(19287) == nil {
		t.Fatal("Darkmoon Card: Heroism is absent from the generated database")
	}
	for _, item := range gear.Items {
		if item.Id == 19287 {
			t.Error("unsupported Darkmoon Card: Heroism is equipped")
		}
	}
	sim, cat := newCatWearing(t, shiftingPowerTalents, nil, worn)
	fillResource(sim, cat)
	emptyEnergy(sim, cat)
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("raid fixture could not cast Shifting Power")
	}
	if got := cat.CurrentEnergy(); got != 45 {
		t.Errorf("raid fixture Shifting Power energy = %v, want 45", got)
	}
	if got := cat.ShiftingPower.ReadyAt(); got != 15*time.Second {
		t.Errorf("raid fixture Shifting Power ready at %v, want 15s", got)
	}
}

func TestFeralCatStarterCasts(t *testing.T) {
	request := starterRequest(t)
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceTauren} {
		t.Run(race.String(), func(t *testing.T) {
			request.Raid.Parties[0].Players[0].Race = race
			stats := core.ComputeStats(&proto.ComputeStatsRequest{Raid: request.Raid})
			for _, action := range stats.RaidStats.Parties[0].Players[0].RotationStats.PriorityList {
				for _, validation := range action.Validations {
					if validation.LogLevel != proto.LogLevel_Information {
						t.Fatalf("rotation validation: %v", validation)
					}
				}
			}
			result := core.RunRaidSim(request)
			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			player := result.RaidMetrics.Parties[0].Players[0]
			casts := make(map[int32]int32)
			for _, action := range player.Actions {
				for _, target := range action.Targets {
					casts[action.Id.GetSpellId()] += target.Casts
				}
			}
			for _, id := range []int32{9830, 9896, 9904, 1322605} {
				if casts[id] == 0 {
					t.Errorf("starter rotation never casts spell %d", id)
				}
			}
			t.Logf("casts=%v DPS=%g DTPS=%g", casts, player.Dps.Avg, player.Dtps.Avg)
		})
	}
}
