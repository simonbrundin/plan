import type { Goal } from "~/types/goal";

/** Delat urval för bulkändringar i en mållista. */
export function useGoalSelection() {
  const selectedIds = ref<Set<number>>(new Set());

  function isSelected(goalId: number): boolean {
    return selectedIds.value.has(goalId);
  }

  function toggle(goalId: number) {
    const next = new Set(selectedIds.value);
    if (next.has(goalId)) next.delete(goalId);
    else next.add(goalId);
    selectedIds.value = next;
  }

  function setAll(goals: ReadonlyArray<Goal>, selected: boolean) {
    const next = new Set(selectedIds.value);
    for (const goal of goals) {
      if (selected) next.add(goal.id);
      else next.delete(goal.id);
    }
    selectedIds.value = next;
  }

  function clear() {
    selectedIds.value = new Set();
  }

  function prune(goals: ReadonlyArray<Goal>) {
    const validIds = new Set(goals.map((goal) => goal.id));
    selectedIds.value = new Set(
      [...selectedIds.value].filter((id) => validIds.has(id)),
    );
  }

  const count = computed(() => selectedIds.value.size);

  return { selectedIds, count, isSelected, toggle, setAll, clear, prune };
}
