<script setup lang="ts">
interface Todo {
  id: string;
  text: string;
  completed: boolean;
}

interface Props {
  todo: Todo;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  toggle: [id: string];
  delete: [id: string];
}>();

const toggleTodo = () => {
  emit('toggle', props.todo.id);
};

const deleteTodo = () => {
  emit('delete', props.todo.id);
};
</script>

<template>
  <div class="flex items-center gap-3 p-3 bg-white rounded-lg border border-gray-200 hover:shadow-sm transition-shadow">
    <input
      type="checkbox"
      :checked="todo.completed"
      @change="toggleTodo"
      class="h-4 w-4 text-blue-600 rounded focus:ring-blue-500"
    />
    <span
      :class="[
        'flex-1 text-sm',
        todo.completed ? 'line-through text-gray-500' : 'text-gray-900'
      ]"
    >
      {{ todo.text }}
    </span>
    <button
      @click="deleteTodo"
      class="text-red-500 hover:text-red-700 p-1"
      aria-label="Ta bort todo"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  </div>
</template>