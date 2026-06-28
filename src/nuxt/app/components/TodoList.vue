<script setup lang="ts">
interface Todo {
  id: string;
  text: string;
  completed: boolean;
}

interface Props {
  todos: Todo[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  toggle: [id: string];
  delete: [id: string];
}>();

const toggleTodo = (id: string) => {
  emit('toggle', id);
};

const deleteTodo = (id: string) => {
  emit('delete', id);
};

const completedCount = computed(() =>
  props.todos.filter(todo => todo.completed).length
);

const totalCount = computed(() => props.todos.length);
</script>

<template>
  <div class="space-y-4">
    <div v-if="todos.length > 0" class="text-sm text-gray-600 mb-4">
      {{ completedCount }} av {{ totalCount }} uppgifter klara
    </div>

    <div v-if="todos.length === 0" class="text-center py-8 text-gray-500">
      <svg class="w-12 h-12 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <p>Inga todos än. Lägg till din första uppgift!</p>
    </div>

    <div v-else class="space-y-2">
      <TodoItem
        v-for="todo in todos"
        :key="todo.id"
        :todo="todo"
        @toggle="toggleTodo"
        @delete="deleteTodo"
      />
    </div>
  </div>
</template>