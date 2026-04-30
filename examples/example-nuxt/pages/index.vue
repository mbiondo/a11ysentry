<template>
  <div>
    <h1>Welcome to Nuxt</h1>
    <p>The intuitive Vue framework.</p>

    <!-- A11Y ISSUE: image missing alt (WCAG_1_1_1) -->
    <img src="/hero.png" />

    <!-- A11Y ISSUE: input missing label (WCAG_3_3_2) -->
    <input type="search" placeholder="Search..." />

    <AppCard title="Getting Started" />
    <AppCard title="Documentation" />
  </div>
</template>

<script setup>
import AppCard from '../components/AppCard.vue';
</script>
