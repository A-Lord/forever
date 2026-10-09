package druid

import (
	"math"

	"github.com/wowsims/forever/sim/core"
)

func (druid *Druid) GetCatWeapon() core.Weapon {
	if druid.GetMHWeapon() == nil {
		return unarmedPaw(1)
	}
	flatWeaponDamage := druid.GetMHWeapon().Enchant.WeaponDamage + druid.GetMHWeapon().TempEnchantWeaponDamage
	unscaledWeapon := druid.WeaponFromMainHand()
	if druid.GetMHWeapon().UnroundedDPS != nil {
		return druid.tablePaw(unscaledWeapon, 1000, flatWeaponDamage)
	}
	return core.Weapon{
		BaseDamageMin:        (unscaledWeapon.BaseDamageMin-flatWeaponDamage)/unscaledWeapon.SwingSpeed + flatWeaponDamage,
		BaseDamageMax:        (unscaledWeapon.BaseDamageMax-flatWeaponDamage)/unscaledWeapon.SwingSpeed + flatWeaponDamage,
		SwingSpeed:           1.0,
		NormalizedSwingSpeed: 1.0,
		AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
		MaxRange:             core.MaxMeleeRange,
	}
}

func (druid *Druid) GetBearWeapon() core.Weapon {
	if druid.GetMHWeapon() == nil {
		return unarmedPaw(2.5)
	}
	unscaledWeapon := druid.WeaponFromMainHand()
	if druid.GetMHWeapon().UnroundedDPS != nil {
		return druid.tablePaw(unscaledWeapon, 2500, 0)
	}
	return core.Weapon{
		BaseDamageMin:        unscaledWeapon.BaseDamageMin / unscaledWeapon.SwingSpeed * 2.5,
		BaseDamageMax:        unscaledWeapon.BaseDamageMax / unscaledWeapon.SwingSpeed * 2.5,
		SwingSpeed:           2.5,
		NormalizedSwingSpeed: 2.5,
		AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
		MaxRange:             core.MaxMeleeRange,
	}
}

func (druid *Druid) tablePaw(unscaledWeapon core.Weapon, swingMs float32, postNormalizationFlat float64) core.Weapon {
	item := druid.GetMHWeapon()
	perSwing := float32(float32(*item.UnroundedDPS) * swingMs)
	perSwing = float32(perSwing * float32(0.001))
	halfVariance := float32(float32(0.4) / float32(2))
	low := float32(perSwing * float32(float32(1)-halfVariance))
	high := float32(perSwing * float32(float32(1)+halfVariance))
	swingSpeed := float64(swingMs) / 1000
	return core.Weapon{
		BaseDamageMin:        max(1, math.Floor(float64(low))) + (unscaledWeapon.BaseDamageMin-item.WeaponDamageMin-postNormalizationFlat)/unscaledWeapon.SwingSpeed*swingSpeed + postNormalizationFlat,
		BaseDamageMax:        max(1, math.Floor(float64(high)+0.5)) + (unscaledWeapon.BaseDamageMax-item.WeaponDamageMax-postNormalizationFlat)/unscaledWeapon.SwingSpeed*swingSpeed + postNormalizationFlat,
		SwingSpeed:           swingSpeed,
		NormalizedSwingSpeed: swingSpeed,
		AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
		MaxRange:             core.MaxMeleeRange,
	}
}

// TODO: Ingame research needed -- unarmed paw endpoints. The model keeps a one-damage floor.
func unarmedPaw(swingSpeed float64) core.Weapon {
	return core.Weapon{BaseDamageMin: 1, BaseDamageMax: 1, SwingSpeed: swingSpeed,
		NormalizedSwingSpeed: swingSpeed, AttackPowerPerDPS: core.DefaultAttackPowerPerDPS, MaxRange: core.MaxMeleeRange}
}
