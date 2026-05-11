<template>
  <Teleport to="body">
    <div
      v-if="isOpen"
      class="modal-overlay"
      @click="close"
      @keydown.esc="close"
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
      aria-describedby="modal-description"
      tabindex="-1"
    >
      <div class="modal-content" @click.stop>
        <header>
          <h2 id="modal-title">Confirm Action</h2>
          <button
            class="close-button"
            @click="close"
            aria-label="Close dialog"
          >
            X
          </button>
        </header>
        
        <div id="modal-description" class="modal-body">
          <p>Are you sure you want to proceed with this highly destructive action? This cannot be undone.</p>
        </div>

        <footer>
          <!-- Explicitly testing the a11ysentry's ability to map shorthand bindings in Vue -->
          <button @click="close" class="btn-cancel">Cancel</button>
          <button v-on:click="confirm" class="btn-confirm" aria-disabled="false">Confirm</button>
        </footer>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  isOpen: Boolean
})

const emit = defineEmits(['close', 'confirm'])

const close = () => {
  emit('close')
}

const confirm = () => {
  emit('confirm')
  close()
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-content {
  background: white;
  padding: 20px;
  border-radius: 8px;
  width: 100%;
  max-width: 500px;
}

.close-button {
  background: transparent;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
}

.btn-cancel {
  background-color: #e0e0e0;
  color: #333;
  border: none;
  padding: 8px 16px;
  cursor: pointer;
}

.btn-confirm {
  background-color: #d32f2f;
  color: white;
  border: none;
  padding: 8px 16px;
  margin-left: 10px;
  cursor: pointer;
}
</style>