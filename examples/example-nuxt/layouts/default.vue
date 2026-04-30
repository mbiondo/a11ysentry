<template>
  <!-- A11Y ISSUE: html missing lang attribute (WCAG_3_1_1) -->
  <div>
    <header>
      <nav>
        <NuxtLink to="/">Home</NuxtLink>
        <NuxtLink to="/about">About</NuxtLink>
        <NuxtLink to="/contact">Contact</NuxtLink>
      </nav>
    </header>
    <main>
      <slot />
    </main>
    <!-- A11Y ISSUE: low contrast footer text #999 on white -->
    <footer style="color: #999999; background-color: #ffffff;">
      <p>© 2024 Example Nuxt Site</p>
    </footer>
  </div>
</template>
