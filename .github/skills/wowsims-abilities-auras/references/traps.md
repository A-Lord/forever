# Traps in a hand-built spell

The defect classes of a spell built beside its row instead of from it. The cases reviews found for each are in
`.github/skills/wowsims-review/references/defects.md`, under the section named after the class.

- **Hand code the resolvers already build.** A duration, cooldown, cost, refund, tick count or tick the resolver fills
  is never typed beside it, and an `OnTick` equal to `DotConfig`'s default is redundant;
  `go run ./tools/spelldata -config '<call>' -package <class>` shows what the resolver fills. Cases: §Hand code the
  resolvers already build.
- **What scales.** A flat weapon bonus is added before the weapon percent, an effect's `PointsPerResource` scales it per
  combo point, and an amount priced by stacks reads the stack count its tooltip names. Cases: §What scales.
- **How many targets.** An effect whose target is an area (`EffectN(n).HitsAnArea()`) reaches every target in it, and
  "up to $s3 targets" is the total count. Cases: §How many targets.
- **What the row forbids.** A hand-built config takes its requirements from `row.CastRequirement()`, since without it
  every form, combat and equipment rule of the row is lost, and a form a class sets on an aura's gain is cleared on its
  expiry. Cases: §What the row forbids.
