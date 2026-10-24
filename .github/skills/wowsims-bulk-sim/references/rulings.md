# Rulings

- **The batch runs the current encounter first plus each checked saved encounter with one seed, skips a duplicate
  checked encounter, sits behind `showPowerTools`, compares encounters in a table only, and the CSV rank counts the
  current gear row.** `ui/features/bulk/model/results_csv.ts`, PR #405.
