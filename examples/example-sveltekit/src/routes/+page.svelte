<script>
  import Card from '$lib/components/Card.svelte';
  import A11yForm from '$lib/components/A11yForm.svelte';

  function handleFormSubmit(event) {
    console.log('Form submitted:', event.detail);
  }
</script>

<h1>Welcome to SvelteKit</h1>
<p>Build fast, lean web applications.</p>

<!-- A11Y ISSUE: image missing alt (WCAG_1_1_1) -->
<img src="/hero.png" />

<!-- A11Y ISSUE: input missing label (WCAG_3_3_2) -->
<input type="email" placeholder="Subscribe..." />

<section class="form-section">
  <A11yForm on:submit={handleFormSubmit} />
</section>

<div class="grid">
  <Card title="Getting Started" />
  <Card title="Documentation" />
</div>

<style>
  .form-section {
    margin: 2rem 0;
    padding: 1.5rem;
    background: #f9f9f9;
    border-radius: 8px;
    border: 1px solid #eee;
  }
</style>
