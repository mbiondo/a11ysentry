<script lang="ts">
  export let data;
</script>

<div style="background-color: #ffebee; color: #b71c1c; padding: 2rem; border-radius: 8px;">
  <h1>Error {data.status}</h1>
  <p>{data.message}</p>
  <a href="/">Back to safety</a>
</div>
