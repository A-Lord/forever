//go:build with_db

package core

import "testing"

func TestWeaponMetadataEquippedFromDefaultDatabase(t *testing.T) {
	for id, want := range map[int32]float64{273456: 14.695897, 7230: 19.808466, 5279: 10.123344, 7005: 1.6533102, 15444: 11.928854051282052, 2028: 9.63904, 6256: 1.0142055, 866: 31.362663} {
		// Given
		spec := ItemSpec{ID: id}
		// When
		item := NewItem(spec)
		// Then
		if item.UnroundedDPS == nil || *item.UnroundedDPS != want {
			t.Errorf("default equipped item %d lost qualified source metadata: %+v", id, item)
		}
	}
}
