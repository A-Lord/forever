//go:build with_db

package feralcat

import (
	"os"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

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
	if got := cat.CurrentEnergy(); got != 60 {
		t.Errorf("raid fixture Shifting Power energy = %v, want 60", got)
	}
	if got := cat.ShiftingPower.ReadyAt(); got != 15*time.Second {
		t.Errorf("raid fixture Shifting Power ready at %v, want 15s", got)
	}
}
