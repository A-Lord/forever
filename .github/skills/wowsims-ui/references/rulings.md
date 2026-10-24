# Rulings

**Source of truth:** the file each line names; `git log --oneline -S '<identifier>' -- <file>` finds the change that
last moved a rule.

- **TPS shows for every spec next to DPS; DTPS, TMI, CoD and the other tank metrics stay behind `showThreatMetrics`,
  which stays off for DPS specs.** `ui/features/results/model/topline_metrics.ts`, PR #233.
- **With item swapping on, a bordered kit Box holds the swap gear row and its stone, oil and imbue pickers; the box is
  hidden, not unmounted, so the swap pickers zero their picks when it is off.**
  `ui/features/item-swap/components/ItemSwapPicker/ItemSwapPicker.tsx`, PR #338.
- **The stat sweep is its own tool behind "Show advanced options", records DPS, HPS, TPS, DTPS and TMI per step,
  always normalizes (range and step in rating, a percent stat steps by its rating conversion), and stops at a cap only
  when every metric is bit-identical.** `ui/features/stat-sweep/`, PR #446.
- **The results "Group ranks" toggle (default on, Damage, Threat, Healing, Damage Taken and Casts tabs) groups ranks by
  unit, base name and lowest tag, nests spell, rank, tag, and never groups pets or unranked same-name spells.**
  `ui/features/results/model/rank_groups.ts`, PR #510.
