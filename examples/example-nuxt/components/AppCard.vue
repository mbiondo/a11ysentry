<template>
  <article class="card">
    <!-- A11Y ISSUE: low contrast blue-400 on white (#60a5fa = ~3.0:1) -->
    <h2 style="color: #60a5fa; background-color: #ffffff;">{{ title }}</h2>
    <p>Learn more about {{ title }} in the docs.</p>
    <button @click="$emit('click')">Read more</button>
  </article>
</template>

<script setup>
defineProps({ title: String });
defineEmits(['click']);
</script>
