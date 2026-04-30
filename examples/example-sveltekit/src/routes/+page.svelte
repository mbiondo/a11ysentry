<script>
  import Card from '$lib/components/Card.svelte';
</script>

<h1>Welcome to SvelteKit</h1>
<p>Build fast, lean web applications.</p>

<!-- A11Y ISSUE: image missing alt (WCAG_1_1_1) -->
<img src="/hero.png" />

<!-- A11Y ISSUE: input missing label (WCAG_3_3_2) -->
<input type="email" placeholder="Subscribe..." />

<div class="grid">
  <Card title="Getting Started" />
  <Card title="Documentation" />
</div>
