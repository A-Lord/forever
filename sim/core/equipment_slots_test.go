package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestEquipmentPreservesEmptyFirstGroupedSlot(t *testing.T) {
	for _, tc := range []struct {
		name          string
		id            int32
		kind          proto.ItemType
		first, second proto.ItemSlot
	}{
		{"ring", 994101, proto.ItemType_ItemTypeFinger, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2},
		{"trinket", 994102, proto.ItemType_ItemTypeTrinket, proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: tc.id, Type: tc.kind, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
			items := make([]*proto.ItemSpec, NumItemSlots)
			for i := range items {
				items[i] = &proto.ItemSpec{}
			}
			items[tc.second].Id = tc.id
			gear := ProtoToEquipment(&proto.EquipmentSpec{Items: items})
			if gear[tc.first].ID != 0 || gear[tc.second].ID != tc.id {
				t.Fatalf("paired slots = [%d, %d], want [0, %d]", gear[tc.first].ID, gear[tc.second].ID, tc.id)
			}
			compact := ProtoToEquipment(&proto.EquipmentSpec{Items: []*proto.ItemSpec{{Id: tc.id}}})
			if compact[tc.first].ID != tc.id {
				t.Fatal("compact gear no longer fills its first eligible slot")
			}
		})
	}
}
