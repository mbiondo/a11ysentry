<template>
  <main>
    <h1>Vue Example</h1>
    <img :alt="dynamicTitle" src="/logo.png" />
    <button @click="doSomething"></button> <!-- Violation: empty button -->
  </main>
</template>
<script setup>
import { ref } from 'vue'
const dynamicTitle = ref('A11ySentry Vue')
const doSomething = () => {}
</script>
