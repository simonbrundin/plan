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
      class="flex-1 px-4 py-2 border text-gray-600 border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
    />
    <button
      @click="addTodo"
      :disabled="!newTodo.trim()"
      class="px-6 py-2 bg-blue-600 text-gray-600 rounded-lg hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors"
    >
      Lägg till
    </button>
  </div>
</template>
