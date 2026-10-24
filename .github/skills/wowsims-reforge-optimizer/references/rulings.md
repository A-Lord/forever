# Rulings

- **The gear optimizer is a single-gear enchants optimizer: candidates are every static enchant (`proc_kind` None,
  with stats, EP above 0) the player can use, the UI picks the slots and Go optimizes every listed one, a slot that
  can take an implemented proc enchant is skipped (combat procs for every spec, survival procs for tanks), and bulk
  never runs it.** `sim/core/reforge_optimizer/enchants.go`, PR #222.
- **Stored EP weights of percent stats (hit, crit, haste, expertise, school hit) are per percentage point; the stat
  weights dialog only displays them per rating point, and source writes a per-rating value as
  `x * Mechanics.*_RATING_PER_*_PERCENT`.** `ui/features/stat-weights/components/EpWeightsDialog/StatWeightRow.tsx`,
  PR #579.
- **The EP target ± is the 90% interval the EP popup shows (1.645 × SE of stdev/abs(refWeight)), reached by
  bulk-style top-ups at seed + n and applied to every shown metric; there is no cap by default.**
  `sim/core/statweight.go`, PR #404.
