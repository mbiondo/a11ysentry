<script>
  import { createEventDispatcher } from 'svelte';
  
  export let actionUrl = "/api/submit";
  export let formError = "";
  
  let name = "";
  let termsAccepted = false;
  let nameInput;
  
  const dispatch = createEventDispatcher();
  
  function handleSubmit(event) {
    if (!name || !termsAccepted) {
      event.preventDefault();
      formError = "Please fill out all required fields.";
      nameInput.focus();
      return;
    }
    dispatch('submit', { name, termsAccepted });
  }
</script>

<!-- Complex accessible form using Svelte bindings and a11y attributes -->
<form
  action={actionUrl}
  method="POST"
  on:submit={handleSubmit}
  aria-labelledby="form-heading"
  aria-describedby={formError ? "form-error" : undefined}
>
  <h2 id="form-heading">Registration Form</h2>
  
  {#if formError}
    <div id="form-error" class="error-alert" role="alert" aria-live="assertive">
      {formError}
    </div>
  {/if}

  <div class="field">
    <label for="user-name">Full Name <span class="required" aria-hidden="true">*</span></label>
    <!-- testing bind:value and aria-invalid -->
    <input
      type="text"
      id="user-name"
      name="user-name"
      bind:this={nameInput}
      bind:value={name}
      aria-required="true"
      aria-invalid={name === "" && formError !== "" ? "true" : "false"}
    />
  </div>

  <div class="field checkbox">
    <input
      type="checkbox"
      id="terms"
      name="terms"
      bind:checked={termsAccepted}
      aria-required="true"
    />
    <label for="terms">I accept the terms and conditions</label>
  </div>

  <div class="actions">
    <button type="submit" class="primary-btn" aria-disabled={!termsAccepted}>
      Register
    </button>
  </div>
</form>

<style>
  .error-alert {
    background-color: #ffebee;
    color: #c62828;
    padding: 10px;
    margin-bottom: 15px;
    border-radius: 4px;
    border: 1px solid #ef9a9a;
  }
  
  .field {
    margin-bottom: 1rem;
  }
  
  .required {
    color: red;
  }
  
  .primary-btn {
    background-color: #1976d2;
    color: white;
    padding: 10px 20px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }
  
  .primary-btn[aria-disabled="true"] {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>