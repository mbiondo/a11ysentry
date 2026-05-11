<template>
  <main>
    <h1>Vue Example</h1>
    <img :alt="dynamicTitle" src="/logo.png" />
    <button @click="doSomething"></button> <!-- Violation: empty button -->
    
    <button class="open-modal-btn" @click="isModalOpen = true">Delete Account</button>
    <ComplexModal 
      :isOpen="isModalOpen" 
      @close="isModalOpen = false" 
      @confirm="handleConfirm" 
    />
  </main>
</template>
<script setup>
import { ref } from 'vue'
import ComplexModal from './ComplexModal.vue'

const dynamicTitle = ref('A11ySentry Vue')
const isModalOpen = ref(false)

const doSomething = () => {}

const handleConfirm = () => {
  console.log('Confirmed!')
}
</script>
<style>
.open-modal-btn {
  margin-top: 20px;
  padding: 10px 20px;
  background-color: #f44336;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
</style>
