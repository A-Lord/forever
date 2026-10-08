//go:build with_db

package forever_test

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	_ "github.com/wowsims/forever/sim/mage" // registers Atiesh, Greatstaff of the Guardian.
)

// The warrior stand-in of class class wearing one item, in a party with these party buffs.
func partyItemWearer(t *testing.T, seed int64, class proto.Class, slot proto.ItemSlot, itemID int32, party *proto.PartyBuffs) (*core.Simulation, *core.Character) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[slot].Id = itemID
	player := &proto.Player{
		Name: "Wearer", Class: class, Race: proto.Race_RaceTroll,
		Equipment:   &proto.EquipmentSpec{Items: items},
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	core.WithSpec(player, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}})
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: seed},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: party}}},
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Darkspear 12802's 1314143 grants 16939, 4% physical and spell crit for 20 s, on 5% of melee hits
// with 60 s between procs. Over 300 wielders the first proc takes 20 hits on average.
func TestDarkspearGrantsItsPartyCritOnAFivePercentChance(t *testing.T) {
	hitsToProc := func(sim *core.Simulation, character *core.Character, buff *core.Aura) int {
		t.Helper()
		hits := 0
		for ; !buff.IsActive(); hits++ {
			if hits == 1000 {
				t.Fatalf("1000 melee hits granted no Darkspear crit")
			}
			dealHit(t, sim, character, meleeSwing(&character.Unit))
		}
		return hits
	}

	total := 0
	const wielders = 300
	for seed := range int64(wielders) {
		sim, character := partyItemWearer(t, seed+1, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotMainHand, 12802, &proto.PartyBuffs{})
		buff := character.GetAura("Darkspear (Player)")
		if buff == nil {
			t.Fatal("12802 registers no Darkspear crit")
		}
		total += hitsToProc(sim, character, buff)
	}
	if mean := float64(total) / wielders; mean < 14 || mean > 27 {
		t.Errorf("the first proc took %v hits on average, want about 20", mean)
	}

	sim, character := partyItemWearer(t, 1, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotMainHand, 12802, &proto.PartyBuffs{})
	buff := character.GetAura("Darkspear (Player)")
	before := character.GetStats()
	hitsToProc(sim, character, buff)
	raised := sim.CurrentTime
	for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
		if got := character.GetStat(stat) - before[stat]; got != 4 {
			t.Errorf("the buff adds %v %v, want 4", got, stat)
		}
	}
	if buff.Duration != 20*time.Second {
		t.Errorf("the buff lasts %v, want 20s", buff.Duration)
	}

	stepTo(t, sim, raised+20*time.Second+time.Millisecond)
	for range 200 {
		dealHit(t, sim, character, meleeSwing(&character.Unit))
	}
	if buff.IsActive() {
		t.Errorf("the buff came back within the 60 s between procs")
	}
	stepTo(t, sim, raised+60*time.Second)
	hitsToProc(sim, character, buff)
}

// The party's Darkspear and the wielder's own bid in one category: with both up, the wielder has 4%
// crit, not 8%.
func TestDarkspearAndThePartysDarkspearDoNotStack(t *testing.T) {
	sim, character := partyItemWearer(t, 1, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotMainHand, 12802, &proto.PartyBuffs{Darkspear: true})
	own, party := character.GetAura("Darkspear (Player)"), character.GetAura("Darkspear (External)")
	if own == nil || party == nil {
		t.Fatalf("the wielder carries %v and %v, want both copies", own, party)
	}
	own.Deactivate(sim)
	party.Deactivate(sim)
	before := character.GetStat(stats.PhysicalCritPercent)

	party.Activate(sim)
	own.Activate(sim)
	if got := character.GetStat(stats.PhysicalCritPercent) - before; got != 4 {
		t.Errorf("both copies up add %v crit, want 4", got)
	}
}

// Admiral's Hat 10030 uses 12022 on its 1 min cooldown: 10 Stamina for 10 min. Used beside the
// party's Admiral's Hat it adds nothing more.
func TestAdmiralsHatGrantsTenStaminaOnce(t *testing.T) {
	for _, tc := range []struct {
		name  string
		party *proto.PartyBuffs
		want  float64
	}{
		{"alone", &proto.PartyBuffs{}, 10},
		{"beside the party's", &proto.PartyBuffs{AdmiralsHat: true}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, character := partyItemWearer(t, 1, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotHead, 10030, tc.party)
			before := character.GetStats()
			useTrinket(t, sim, character, 10030, core.CooldownTypeDPS, time.Minute, 0, 0)
			if got := gained(character, before, stats.Stamina); got != tc.want {
				t.Errorf("the hat adds %v Stamina, want %v", got, tc.want)
			}
			if got := character.GetAura("Admiral's Hat (Player)").Duration; got != 10*time.Minute {
				t.Errorf("the hat's buff lasts %v, want 10m", got)
			}
		})
	}

	_, bare := partyItemWearer(t, 1, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotHead, 0, &proto.PartyBuffs{})
	_, buffed := partyItemWearer(t, 1, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotHead, 0, &proto.PartyBuffs{AdmiralsHat: true})
	if got := buffed.GetStat(stats.Stamina) - bare.GetStat(stats.Stamina); got != 10 {
		t.Errorf("the party's Admiral's Hat adds %v Stamina, want 10", got)
	}
}

// Atiesh, Greatstaff of the Guardian 22589's party aura 28142 gives the mage wearing it 2% spell crit,
// and the party buff's count adds the other members' staves on top.
func TestAtieshGreatstaffOfTheGuardianGivesItsWearerTwoPercentSpellCrit(t *testing.T) {
	_, bare := partyItemWearer(t, 1, proto.Class_ClassMage, proto.ItemSlot_ItemSlotMainHand, 0, &proto.PartyBuffs{})
	_, wearer := partyItemWearer(t, 1, proto.Class_ClassMage, proto.ItemSlot_ItemSlotMainHand, 22589, &proto.PartyBuffs{})
	_, counted := partyItemWearer(t, 1, proto.Class_ClassMage, proto.ItemSlot_ItemSlotMainHand, 22589, &proto.PartyBuffs{AtieshMage: 1})

	spellCrit := func(character *core.Character) float64 { return character.GetStat(stats.SpellCritPercent) }
	if got := spellCrit(wearer) - spellCrit(bare); math.Abs(got-2) > 1e-9 {
		t.Errorf("the staff adds %v spell crit, want 2", got)
	}
	if got := spellCrit(counted) - spellCrit(bare); math.Abs(got-4) > 1e-9 {
		t.Errorf("the staff and one more in the party add %v spell crit, want 4", got)
	}
	if got := wearer.GetStat(stats.PhysicalCritPercent) - bare.GetStat(stats.PhysicalCritPercent); got != 0 {
		t.Errorf("the staff adds %v physical crit, want none", got)
	}
}
