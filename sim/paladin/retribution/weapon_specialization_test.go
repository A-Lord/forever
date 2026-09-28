package retribution

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

const weaponSpecializationAPL = `{"type":"TypeAPL",
"prepullActions":[{"action":{"castSpell":{"spellId":{"spellId":20293,"rank":8}}},"doAtValue":{"const":{"val":"-1.5s"}}}],
"priorityList":[
	{"action":{"condition":{"auraIsInactive":{"auraId":{"spellId":20293,"rank":8}}},"castSpell":{"spellId":{"spellId":20293,"rank":8}}}},
	{"action":{"castSpell":{"spellId":{"spellId":10333,"rank":8}}}},
	{"action":{"castSpell":{"spellId":{"spellId":20924,"rank":5}}}}]}`

func weaponSpecializationDamage(t *testing.T, gearSpec string, talents string) map[string]float64 {
	t.Helper()
	rotation := &proto.APLRotation{}
	if err := protojson.Unmarshal([]byte(weaponSpecializationAPL), rotation); err != nil {
		t.Fatal(err)
	}

	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Name:          "Weapon Specialization",
			Race:          proto.Race_RaceHuman,
			Class:         proto.Class_ClassPaladin,
			Equipment:     core.GetGearSet("../../../ui/specs/paladin/"+gearSpec+"/gear_sets", "default").GearSet,
			TalentsString: talents,
			Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
				Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
			}},
			Rotation: rotation,
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{Iterations: 20, RandomSeed: 1},
	})
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}

	damage := map[string]float64{}
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		for _, target := range action.Targets {
			damage[core.ProtoToActionID(action.Id).String()] += target.Damage
		}
	}
	return damage
}

// One- and Two-Handed Weapon Specialization raise the melee swings and every melee-defense Holy
// spell - the Seal of Righteousness hit and Holy Strike - and leave Consecration, a magic-defense
// spell, alone.
func TestWeaponSpecializationReachesMeleeDefenseHolySpells(t *testing.T) {
	for _, tc := range []struct {
		name    string
		gear    string
		talents string
		want    float64
	}{
		{"One-Handed Weapon Specialization 3/3", "protection", "-00000000003-", 1.10},
		{"Two-Handed Weapon Specialization 3/3", "retribution", "--000000000003", 1.06},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := weaponSpecializationDamage(t, tc.gear, "--")
			spec := weaponSpecializationDamage(t, tc.gear, tc.talents)

			for _, action := range []struct {
				id   core.ActionID
				want float64
			}{
				{core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: 1}, tc.want},
				{core.ActionID{SpellID: 25713}, tc.want},
				{core.ActionID{SpellID: 10333}, tc.want},
				{core.ActionID{SpellID: 20924}, 1},
			} {
				key := action.id.String()
				if base[key] == 0 {
					t.Fatalf("%s dealt no damage", key)
				}
				if got := spec[key] / base[key]; math.Abs(got-action.want) > 1e-9 {
					t.Errorf("%s: damage ratio %.6f, want %.2f", key, got, action.want)
				}
			}
		})
	}
}
