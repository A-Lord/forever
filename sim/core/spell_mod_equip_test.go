package core

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func init() {
	RegisterAgentFactory(
		proto.Player_Rogue{},
		proto.Spec_SpecRogue,
		newFakeWeaponUser,
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Rogue)
			if !ok {
				panic("Invalid spec value for Rogue!")
			}
			player.Spec = playerSpec
		},
	)
}

type fakeWeaponUser struct {
	Character
}

func (fw *fakeWeaponUser) GetCharacter() *Character       { return &fw.Character }
func (fw *fakeWeaponUser) Initialize()                    {}
func (fw *fakeWeaponUser) ApplyTalents()                  {}
func (fw *fakeWeaponUser) Reset(_ *Simulation)            {}
func (fw *fakeWeaponUser) OnGCDReady(_ *Simulation)       {}
func (fw *fakeWeaponUser) OnEncounterStart(_ *Simulation) {}

func newFakeWeaponUser(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
	fw := &fakeWeaponUser{Character: *char}
	fw.EnableAutoAttacks(fw, AutoAttackOptions{
		MainHand:        fw.WeaponFromMainHand(),
		OffHand:         fw.WeaponFromOffHand(),
		Ranged:          fw.WeaponFromRanged(),
		AutoSwingMelee:  true,
		AutoSwingRanged: true,
	})
	return fw
}

const (
	equipTestSwordID int32 = 990901 + iota
	equipTestDaggerID
	equipTestShieldID
	equipTestBowID
	equipTestAxeID
)

func init() {
	damage := map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 50, WeaponDamageMax: 90}}
	addToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{
		{Id: equipTestSwordID, Name: "Test Sword", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand, WeaponSpeed: 2.6, ScalingOptions: damage},
		{Id: equipTestDaggerID, Name: "Test Dagger", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand, WeaponSpeed: 1.8, ScalingOptions: damage},
		{Id: equipTestShieldID, Name: "Test Shield", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeShield, HandType: proto.HandType_HandTypeOffHand, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}},
		{Id: equipTestBowID, Name: "Test Bow", Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow, WeaponSpeed: 3.0, ScalingOptions: damage},
		{Id: equipTestAxeID, Name: "Test Axe", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeOneHand, WeaponSpeed: 2.6, ScalingOptions: damage},
	}})
}

func weaponSlots(mainHand, offHand, ranged int32) []*proto.ItemSpec {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: mainHand}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: offHand}
	items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{Id: ranged}
	return items
}

func setupWeaponSim(t *testing.T, equipped, swapped []*proto.ItemSpec) (*Simulation, *Character) {
	t.Helper()
	return setupWeaponSimAs(t, proto.Race_RaceUndead, equipped, swapped)
}

func setupWeaponSimAs(t *testing.T, race proto.Race, equipped, swapped []*proto.ItemSpec) (*Simulation, *Character) {
	t.Helper()
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       weaponTestRaid(race, equipped, swapped),
		Encounter:  weaponTestEncounter(),
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

func weaponTestEncounter() *proto.Encounter {
	return &proto.Encounter{
		Targets:  []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}},
		Duration: 180,
	}
}

func weaponTestRaid(race proto.Race, equipped, swapped []*proto.ItemSpec) *proto.Raid {
	player := &proto.Player{
		Name:      "Weapon Tester",
		Class:     proto.Class_ClassRogue,
		Race:      race,
		Buffs:     &proto.IndividualBuffs{},
		Spec:      &proto.Player_Rogue{},
		Equipment: &proto.EquipmentSpec{Items: equipped},
		Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	if swapped != nil {
		player.EnableItemSwap = true
		player.ItemSwap = &proto.ItemSwap{Items: swapped}
	}
	return &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}}
}

func checkReaches(t *testing.T, when string, mod *SpellMod, spells map[string]*Spell, want ...string) {
	t.Helper()
	for name, spell := range spells {
		if got := slices.Contains(mod.AffectedSpells, spell); got != slices.Contains(want, name) {
			t.Errorf("%s: mod reaches the %s = %v, want %v", when, name, got, !got)
		}
	}
}

func autos(character *Character) map[string]*Spell {
	return map[string]*Spell{
		"main-hand auto": character.AutoAttacks.MHAuto(),
		"off-hand auto":  character.AutoAttacks.OHAuto(),
		"ranged auto":    character.AutoAttacks.RangedAuto(),
	}
}

func TestSubclassModFollowsTheWeaponBehindTheAttack(t *testing.T) {
	_, character := setupWeaponSim(t, weaponSlots(equipTestSwordID, equipTestDaggerID, equipTestBowID), nil)
	if !character.AutoAttacks.IsDualWielding {
		t.Fatal("sword and dagger should dual wield")
	}

	swords := character.AddDynamicMod(SpellModConfig{
		Kind:          SpellMod_BonusCrit_Percent,
		FloatValue:    5,
		EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword),
	})
	checkReaches(t, "sword and dagger", swords, autos(character), "main-hand auto")

	shooters := character.AddDynamicMod(SpellModConfig{
		Kind:          SpellMod_BonusHit_Percent,
		FloatValue:    3,
		EquipSubclass: 262156,
	})
	checkReaches(t, "bow, gun or crossbow", shooters, autos(character), "ranged auto")

	offHand := character.AddDynamicMod(SpellModConfig{
		Kind:          SpellMod_BonusHit_Percent,
		FloatValue:    2,
		EquipSubclass: 41105,
		EquipInvType:  dbcenums.INVTYPE_BIT_WEAPONOFFHAND,
	})
	checkReaches(t, "one-hander in the off hand", offHand, autos(character), "off-hand auto")

	offHandByAttr := character.AddDynamicMod(SpellModConfig{
		Kind:          SpellMod_BonusHit_Percent,
		FloatValue:    2,
		EquipSubclass: 41105,
		EquipHand:     ProcMaskMeleeOH,
	})
	checkReaches(t, "off-hand bit", offHandByAttr, autos(character), "off-hand auto")
}

func TestShieldIsNotAWeapon(t *testing.T) {
	_, character := setupWeaponSim(t, weaponSlots(equipTestSwordID, equipTestShieldID, 0), nil)

	offHand := EquipRequirement{InvTypeMask: dbcenums.INVTYPE_BIT_SHIELD | dbcenums.INVTYPE_BIT_WEAPONOFFHAND | dbcenums.INVTYPE_BIT_HOLDABLE}
	anyWeapon := EquipRequirement{SubclassMask: -1}
	for _, procMask := range []ProcMask{ProcMaskMeleeOHAuto, ProcMaskMeleeOHSpecial} {
		if offHand.Matches(&character.Unit, procMask) || anyWeapon.Matches(&character.Unit, procMask) {
			t.Errorf("a shield in the off hand matched a weapon requirement for proc mask %v", procMask)
		}
	}
	if !anyWeapon.Matches(&character.Unit, ProcMaskMeleeMHSpecial) {
		t.Error("the sword in the main hand should match any weapon")
	}
	if anyWeapon.Matches(&character.Unit, ProcMaskSpellDamage) {
		t.Error("an attack with no weapon bit should never match")
	}
}

func TestEquipRequirementNeverMatchesOtherUnits(t *testing.T) {
	sim, character := setupWeaponSim(t, weaponSlots(equipTestSwordID, equipTestDaggerID, equipTestBowID), nil)
	anyWeapon := EquipRequirement{SubclassMask: -1}
	if !anyWeapon.Matches(&character.Unit, ProcMaskMeleeMHAuto) {
		t.Fatal("the player's own sword should match")
	}

	pet := &Unit{Type: PetUnit, Env: character.Env}
	target := sim.Encounter.ActiveTargetUnits[0]
	for name, unit := range map[string]*Unit{"pet": pet, "target": target} {
		if anyWeapon.Matches(unit, ProcMaskMeleeMHAuto) {
			t.Errorf("a %s matched a weapon requirement", name)
		}
		mod := &SpellMod{EquipSubclass: -1}
		if shouldApply(&Spell{Unit: unit, ProcMask: ProcMaskMeleeMHAuto}, mod) {
			t.Errorf("a weapon mod applied to a %s's attack", name)
		}
	}
}

func TestSubclassModFollowsAnItemSwap(t *testing.T) {
	sim, character := setupWeaponSim(t,
		weaponSlots(equipTestSwordID, equipTestDaggerID, 0),
		weaponSlots(equipTestDaggerID, equipTestSwordID, 0))

	swords := character.AddDynamicMod(SpellModConfig{
		Kind:          SpellMod_BonusCrit_Percent,
		FloatValue:    5,
		EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword),
	})
	swords.Activate()
	mhAuto, ohAuto := character.AutoAttacks.MHAuto(), character.AutoAttacks.OHAuto()
	baseMH, baseOH := mhAuto.BonusCritPercent-5, ohAuto.BonusCritPercent

	check := func(when string, mh, oh float64) {
		t.Helper()
		if got := mhAuto.BonusCritPercent - baseMH; got != mh {
			t.Errorf("%s: main-hand auto has %v%% from the mod, want %v%%", when, got, mh)
		}
		if got := ohAuto.BonusCritPercent - baseOH; got != oh {
			t.Errorf("%s: off-hand auto has %v%% from the mod, want %v%%", when, got, oh)
		}
	}
	check("sword in the main hand", 5, 0)

	character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
	check("sword swapped to the off hand", 0, 5)

	character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Main, false)
	check("sword swapped back", 5, 0)

	swords.Deactivate()
	character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
	check("inactive mod across a swap", 0, 0)
	swords.Activate()
	check("mod turned on with the sword in the off hand", 0, 5)
}

func TestScopedBonusPercentMatchesTheAutos(t *testing.T) {
	_, character := setupWeaponSim(t, weaponSlots(equipTestSwordID, equipTestDaggerID, equipTestBowID), nil)
	mhAuto, ohAuto, rangedAuto := character.AutoAttacks.MHAuto(), character.AutoAttacks.OHAuto(), character.AutoAttacks.RangedAuto()
	base := map[*Spell]float64{mhAuto: mhAuto.BonusCritPercent, ohAuto: ohAuto.BonusCritPercent, rangedAuto: rangedAuto.BonusCritPercent}

	for _, config := range []SpellModConfig{
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 5, EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword)},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 3, EquipSubclass: 262156},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 2, ProcMask: ProcMaskMeleeOH},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 1},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 7, ClassMask: 1 << 40},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 7, School: SpellSchoolFire},
		{Kind: SpellMod_BonusHit_Percent, FloatValue: 11},
	} {
		character.AddStaticMod(config)
	}
	character.AddDynamicMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 13})

	for _, row := range []struct {
		name  string
		probe ProcMask
		spell *Spell
		want  float64
	}{
		{"main hand", ProcMaskMeleeMHAuto, mhAuto, 6},
		{"off hand", ProcMaskMeleeOHAuto, ohAuto, 3},
		{"ranged", ProcMaskRangedAuto, rangedAuto, 4},
	} {
		got := character.ScopedBonusPercent(SpellMod_BonusCrit_Percent, row.probe)
		if got != row.want {
			t.Errorf("%s: probe = %v, want %v", row.name, got, row.want)
		}
		if onSpell := row.spell.BonusCritPercent - base[row.spell]; got != onSpell {
			t.Errorf("%s: probe = %v, the auto carries %v", row.name, got, onSpell)
		}
	}
}
