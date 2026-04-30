<template>
  <div>
    <h1>About</h1>
    <p>Built with Nuxt 3 and Vue 3.</p>

    <!-- A11Y ISSUE: button missing accessible name (WCAG_4_1_2) -->
    <button class="icon-btn">
      <svg viewBox="0 0 24 24" width="24" height="24">
        <path d="M12 4l8 16H4z" />
      </svg>
    </button>

    <!-- A11Y ISSUE: heading jump h1 -> h4 (WCAG_1_3_1) -->
    <h4>Team Members</h4>
    <ul>
      <li>Alice</li>
      <li>Bob</li>
    </ul>
  </div>
</template>
