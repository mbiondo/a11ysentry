package vue_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	
	"a11ysentry/adapters/vue"
	"a11ysentry/engine/core/domain"
)

func TestVueAdapter_SFC(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a Vue Single File Component
	vueContent := `
<template>
	<div class="card">
		<img :alt="dynamicAlt" src="/hero.png" />
		<img v-bind:alt="dynamicAlt2" src="/hero2.png" />
		<button @click="submit()">Submit</button>
		<p class="text" :class="{ 'error': hasError }">Message</p>
	</div>
</template>

<script setup>
import { ref } from 'vue'
const dynamicAlt = ref('Hero image')
const dynamicAlt2 = ref('Hero image 2')
const hasError = ref(true)
const submit = () => {}
</script>

<style scoped>
.card { background-color: #fff; }
.text { color: #333; }
.error { color: #f00; }
</style>
	`

	os.WriteFile(filepath.Join(tmpDir, "Card.vue"), []byte(vueContent), 0644)

	adapter := vue.NewVueAdapter()
	rootNode := &domain.FileNode{FilePath: filepath.Join(tmpDir, "Card.vue")}

	nodes, err := adapter.Ingest(context.Background(), rootNode)
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}

	foundAlt := false
	foundVBindAlt := false
	foundClick := false
	foundStyle := false
	
	for _, n := range nodes {
		t.Logf("Node: Role=%s, UID=%s, Label=%s, Traits=%v", n.Role, n.UID, n.Label, n.Traits)
		if n.Role == domain.RoleImage {
			if n.Label == "{{dynamicAlt}}" {
				foundAlt = true
			}
			if n.Label == "{{dynamicAlt2}}" {
				foundVBindAlt = true
			}
		}
		if n.Role == domain.RoleButton {
			if val, ok := n.Traits["@click"]; ok && val == "submit()" {
				foundClick = true
			}
		}
		if n.Role == "generic" && n.Traits["color"] == "#f00" || n.Traits["color"] == "#333" {
			foundStyle = true
		}
	}
	
	if !foundAlt {
		t.Errorf("Expected :alt to be parsed as dynamicAlt")
	}
	if !foundVBindAlt {
		t.Errorf("Expected v-bind:alt to be parsed as dynamicAlt2")
	}
	if !foundClick {
		t.Errorf("Expected @click to be preserved in traits")
	}
	if !foundStyle {
		// Currently, dynamic styles might not be fully resolvable via class bindings in the simple parser
		// This is acceptable as a limitation for now, but we check if regular class styles are parsed.
		// For the <p class="text"> it should definitely have color: #333.
		// Wait, if it parses <style scoped>, we expect text color to be #333.
		t.Errorf("Expected scoped <style> to be applied to elements")
	}
}
