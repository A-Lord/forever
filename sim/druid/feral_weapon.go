package druid

import (
	"math"

	"github.com/wowsims/forever/sim/core"
)

func (druid *Druid) GetCatWeapon() core.Weapon {
	if druid.GetMHWeapon() == nil {
		return unarmedPaw(1)
	}
	unscaledWeapon := druid.WeaponFromMainHand()
	if druid.GetMHWeapon().WeaponTableDPS != nil {
		return druid.tablePaw(unscaledWeapon, 1000)
	}
	return core.Weapon{
		BaseDamageMin:        unscaledWeapon.BaseDamageMin / unscaledWeapon.SwingSpeed,
		BaseDamageMax:        unscaledWeapon.BaseDamageMax / unscaledWeapon.SwingSpeed,
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
	if druid.GetMHWeapon().WeaponTableDPS != nil {
		return druid.tablePaw(unscaledWeapon, 2500)
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

func (druid *Druid) tablePaw(unscaledWeapon core.Weapon, swingMs float32) core.Weapon {
	item := druid.GetMHWeapon()
	perSwing := float32(float32(*item.WeaponTableDPS) * swingMs)
	perSwing = float32(perSwing * float32(0.001))
	halfVariance := float32(float32(0.4) / float32(2))
	low := float32(perSwing * float32(float32(1)-halfVariance))
	high := float32(perSwing * float32(float32(1)+halfVariance))
	swingSpeed := float64(swingMs) / 1000
	return core.Weapon{
		BaseDamageMin:        max(1, math.Floor(float64(low))) + (unscaledWeapon.BaseDamageMin-item.WeaponDamageMin)/unscaledWeapon.SwingSpeed*swingSpeed,
		BaseDamageMax:        max(1, math.Floor(float64(high)+0.5)) + (unscaledWeapon.BaseDamageMax-item.WeaponDamageMax)/unscaledWeapon.SwingSpeed*swingSpeed,
		SwingSpeed:           swingSpeed,
		NormalizedSwingSpeed: swingSpeed,
		AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
		MaxRange:             core.MaxMeleeRange,
	}
}

func unarmedPaw(swingSpeed float64) core.Weapon {
	return core.Weapon{BaseDamageMin: 1, BaseDamageMax: 1, SwingSpeed: swingSpeed,
		NormalizedSwingSpeed: swingSpeed, AttackPowerPerDPS: core.DefaultAttackPowerPerDPS, MaxRange: core.MaxMeleeRange}
}
