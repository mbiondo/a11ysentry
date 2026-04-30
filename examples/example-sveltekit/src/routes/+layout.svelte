<script>
  import { page } from '$app/stores';
</script>

<!-- A11Y ISSUE: html missing lang attribute (WCAG_3_1_1) -->
<svelte:head>
  <title>Example SvelteKit</title>
</svelte:head>

<header>
  <nav>
    <a href="/">Home</a>
    <a href="/about">About</a>
    <a href="/dashboard">Dashboard</a>
  </nav>
</header>

<main>
  <slot />
</main>

<!-- A11Y ISSUE: low contrast footer text #999 on white -->
<footer style="color: #999999; background-color: #ffffff;">
  <p>© 2024 Example SvelteKit Site</p>
</footer>
