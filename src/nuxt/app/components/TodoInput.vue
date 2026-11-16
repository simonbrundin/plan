<script setup lang="ts">
const emit = defineEmits<{
  add: [text: string];
}>();

const newTodo = ref("");

const addTodo = () => {
  const text = newTodo.value.trim();
  if (text) {
    emit("add", text);
    newTodo.value = "";
  }
};

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === "Enter") {
    addTodo();
  }
};
</script>

<template>
  <div class="flex gap-2 mb-6">
    <input
      v-model="newTodo"
      @keydown="handleKeydown"
      type="text"
      placeholder="Lägg till en ny uppgift..."
      class="flex-1 px-4 py-2 border text-ink-dim border-line-strong rounded-lg focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent"
    />
    <button
      @click="addTodo"
      :disabled="!newTodo.trim()"
      class="px-6 py-2 bg-accent-strong text-ink-dim rounded-lg hover:bg-accent-deep disabled:bg-elevated disabled:cursor-not-allowed transition-colors"
    >
      Lägg till
    </button>
  </div>
</template>
