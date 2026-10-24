# Traps in a hand-built spell

The defect classes reviews of class ports keep finding where a spell is built beside its row instead of from it.

- **Hand code the resolvers already build.** Before typing a duration, cooldown, cost, refund, tick count or tick, read
  what the resolver fills with `go run ./tools/spelldata -config '<call>' -package <class>`. Blade Flurry 13877's four
  literals, a hand `PeriodicAction` where `DotConfig(rank, periodic)` works out the ticks, and a Rake `OnTick` equal to
  `DotConfig`'s default all went.
- **What scales.** A flat weapon bonus goes before the weapon percent (Ravage 9867: "350% damage plus 343" is 98 ×
  3.5), and an effect's `PointsPerResource` scales it per combo point (Rupture 11275's 4.73).
- **How many targets.** An effect whose target is an area (`EffectN(n).HitsAnArea()`) reaches every target in it
  (Demoralizing Roar 9898), and "up to $s3 targets" is the total, not one plus $s3 (Berserk 417141).
- **What the row forbids.** A hand-built config carries no `CastRequirement`, so it loses the not-in-combat bit
  (Prowl 9913 recast 18 times in 180 s) and every form check; a form the class sets on an aura's gain is cleared on
  its expiry.
