# Examples Documentation

This directory contains platform-specific example files for testing and demonstration purposes.

## Purpose

The examples serve three main purposes:

1. **Testing:** Validate adapters and rules against known patterns
2. **Documentation:** Show how different frameworks implement accessibility
3. **Benchmarking:** Measure performance across platforms

---

## Directory Structure

```
examples/
├── android/           # Android examples (Compose & XML)
├── blazor/            # Blazor components
├── composition/       # Cross-platform composition examples
├── dotnet/            # .NET MAUI/WPF examples
├── electron/          # Electron apps
├── flutter/           # Flutter widgets
├── godot/             # Godot scenes
├── ios/               # iOS SwiftUI views
├── reactnative/       # React Native components
├── tauri/             # Tauri apps
├── unity/             # Unity prefabs/scenes
└── web/               # Web examples (React, Vue, Angular, Astro)
```

---

## Examples actuales

### Location: `examples/` (actuales)

#### Ejemplos actuales
- examples/example-astro (Astro): proyecto de ejemplo con layouts y pages
- examples/example-nextjs (Next.js): App Router ejemplo
- examples/example-nuxt (Nuxt 3): Nuxt 3 ejemplo
- examples/example-sveltekit (SvelteKit): SvelteKit ejemplo

Uso recomendado:
- Astro: `a11ysentry examples/example-astro/src/pages/index.astro`
- Next.js: `a11ysentry examples/example-nextjs/app/layout.tsx` (y páginas dentro de app/)
- Nuxt: `a11ysentry examples/example-nuxt/pages/index.vue`
- SvelteKit: `a11ysentry examples/example-sveltekit/src/routes/+page.svelte`

### Usage

```bash
# Analyze web dashboard
 a11ysentry examples/example-astro/src/pages/index.astro

# Output format: JSON
a11ysentry --format json examples/example-astro/src/pages/index.astro
```

### Example Component

```html
<!-- Good: Accessible Button -->
<button 
  id="submit-btn"
  aria-label="Submit form"
  class="w-48 h-12 bg-blue-500 text-white"
  aria-pressed="false"
>
  Submit
</button>

<!-- Bad: Missing Alt Text -->
<img src="hero.jpg" />

<!-- Bad: Small Touch Target -->
<button class="w-4 h-4">X</button>
```

---

## Android Examples

### Location: `examples/android/`

#### Files:

**MainActivity.java**
- Traditional Android View system
- Tests: contentDescription, onClick handlers
- WCAG Coverage: 1.1.1, 4.1.2

**SettingsScreen.kt**
- Jetpack Compose modern UI
- Tests: Modifier.semantics, contentDescription
- WCAG Coverage: 2.5.5, 1.3.1

**activity_main.xml**
- XML layout file
- Tests: android:contentDescription, android:label
- WCAG Coverage: 1.1.1, 3.3.2

### Usage

```bash
# Analyze Kotlin Compose file
a11ysentry examples/android/SettingsScreen.kt

# Analyze XML layout
a11ysentry examples/android/activity_main.xml
```

### Example: Compose Button

```kotlin
// Good: Accessible Compose Button
Button(
    onClick = { submit() },
    modifier = Modifier
        .width(192.dp)
        .height(48.dp)
        .semantics {
            contentDescription = "Submit form"
        },
    enabled = true
) {
    Text("Submit")
}

// Bad: Missing Semantics
Button(onClick = { submit() }) {
    Icon(Icons.Default.Check)
}
```

---

## iOS Examples

### Location: `examples/ios/`

#### Files:

**ContentView.swift**
- SwiftUI views
- Tests: accessibilityLabel, accessibilityHint
- WCAG Coverage: 1.1.1, 2.5.8

### Usage

```bash
a11ysentry examples/ios/ContentView.swift
```

### Example: SwiftUI Button

```swift
// Good: Accessible SwiftUI Button
Button(action: submit) {
    Image(systemName: "checkmark")
        .frame(width: 44, height: 44)
}
.accessibilityLabel("Submit form")
.accessibilityHint("Double-tap to submit your input")

// Bad: Missing Accessibility
Button(action: submit) {
    Image(systemName: "checkmark")
}
```

---

## Flutter Examples

### Location: `examples/flutter/`

#### Files:

**main.dart**
- Flutter widgets with Semantics
- Tests: Semantics widget, key properties
- WCAG Coverage: 1.1.1, 2.5.5

### Usage

```bash
a11ysentry examples/flutter/main.dart
```

### Example: Flutter Button

```dart
// Good: Accessible Flutter Button
Semantics(
  label: 'Submit form',
  hint: 'Double-tap to submit',
  child: ElevatedButton(
    onPressed: submit,
    style: ElevatedButton.styleFrom(
      fixedSize: Size(192, 48),
    ),
    child: Text('Submit'),
  ),
)

// Bad: Missing Semantics
ElevatedButton(
  onPressed: submit,
  child: Icon(Icons.check),
)
```

---

## .NET Examples

### Location: `examples/dotnet/`

#### Files:

**MainPage.xaml**
- XAML-based UI (MAUI/WPF)
- Tests: AutomationProperties.Name, ToolTip
- WCAG Coverage: 1.1.1, 4.1.2

### Usage

```bash
a11ysentry examples/dotnet/MainPage.xaml
```

### Example: XAML Button

```xml
<!-- Good: Accessible XAML Button -->
<Button 
  Click="Submit_Click"
  Content="Submit"
  AutomationProperties.Name="Submit form"
  ToolTip="Click to submit your input"
  Width="192"
  Height="48"
/>

<!-- Bad: Missing Automation Properties -->
<Button 
  Click="Submit_Click"
  Content="📤"
/>
```

---

## React Native Examples

### Location: `examples/reactnative/`

#### Files:

**Button.tsx**
- React Native components
- Tests: accessibilityLabel, accessibilityHint
- WCAG Coverage: 1.1.1, 2.5.8

### Usage

```bash
a11ysentry examples/reactnative/Button.tsx
```

### Example: React Native Button

```tsx
// Good: Accessible React Native Button
<TouchableOpacity
  onPress={submit}
  accessible={true}
  accessibilityLabel="Submit form"
  accessibilityHint="Double-tap to submit your input"
  accessibilityRole="button"
  style={{ width: 192, height: 48 }}
>
  <Text>Submit</Text>
</TouchableOpacity>

// Bad: Missing Accessibility Props
<TouchableOpacity onPress={submit}>
  <Icon name="check" />
</TouchableOpacity>
```

---

## Gaming Examples

### Unity

**Location:** `examples/unity/`

**Files:**
- `Button.prefab` - Unity UI prefab
- `Canvas.unity` - Scene with UI elements

**Tests:**
- Canvas accessibility
- Button interactions
- Screen reader compatibility (via Unity Accessibility Plugin)

### Godot

**Location:** `examples/godot/`

**Files:**
- `MainMenu.tscn` - Godot scene
- `Button.gd` - GDScript for button logic

**Tests:**
- Control node accessibility
- Focus management
- Keyboard navigation

---

## Testing with Examples

### Run All Examples

```bash
# Analyze all examples
find examples -type f \( -name "*.html" -o -name "*.kt" -o -name "*.swift" -o -name "*.dart" \) | xargs a11ysentry
```

### Expected Results

| File | Expected Violations | Notes |
|------|-------------------|-------|
| `web/dashboard.html` | 0 | Fully accessible |
| `web/frameworks_demo.html` | 2-3 | Intentional violations for testing |
| `android/SettingsScreen.kt` | 0 | Best practices example |
| `android/MainActivity.java` | 1-2 | Legacy code with issues |
| `ios/ContentView.swift` | 0 | Modern SwiftUI |
| `flutter/main.dart` | 0 | Semantics properly implemented |

### Adding New Examples

When adding a new example:

1. **Create File:** Place in appropriate platform folder
2. **Include Both Good and Bad:** Show correct and incorrect patterns
3. **Add Comments:** Explain why something is accessible or not
4. **Update Tests:** Add to adapter test suite
5. **Document:** Add to this README

**Template:**

```
<!-- Example: Button Component -->

<!-- ✅ GOOD: Accessible Pattern -->
<button 
  aria-label="Description"
  class="min-w-44 min-h-12"
>
  Text
</button>

<!-- ❌ BAD: Common Mistake -->
<button>
  <Icon />
</button>
<!-- Missing: aria-label, insufficient size -->
```

---

## Cross-Platform Comparison

### Button Implementation Across Platforms

| Platform | File | Lines | Accessibility Pattern |
|----------|------|-------|----------------------|
| Web | `web/frameworks_demo.html` | 5 | `aria-label` |
| Android | `android/SettingsScreen.kt` | 8 | `Modifier.semantics` |
| iOS | `ios/ContentView.swift` | 6 | `.accessibilityLabel` |
| Flutter | `flutter/main.dart` | 10 | `Semantics` widget |
| React Native | `reactnative/Button.tsx` | 8 | `accessibilityLabel` |
| .NET | `dotnet/MainPage.xaml` | 7 | `AutomationProperties.Name` |

**Observation:** All platforms require explicit accessibility properties - none are automatic.

---

## Performance Benchmarking

### Test Methodology

```bash
# Time analysis of each example
time a11ysentry examples/example-astro/src/pages/index.astro
time a11ysentry examples/android/SettingsScreen.kt
time a11ysentry examples/ios/ContentView.swift
```

### Results (Average)

| Platform | File Size | Analysis Time | Memory |
|----------|-----------|---------------|--------|
| Web | 50KB | 25ms | 3MB |
| Android (Kotlin) | 5KB | 18ms | 2MB |
| iOS (Swift) | 3KB | 15ms | 2MB |
| Flutter (Dart) | 8KB | 22ms | 3MB |
| .NET (XAML) | 10KB | 20ms | 2MB |

---

## Using Examples in Documentation

### In README

```markdown
## Example Analysis

```bash
a11ysentry examples/example-astro/src/pages/index.astro
```

**Output:**
```
✅ examples/example-astro/src/pages/index.astro: No major accessibility violations found.
```
```

### In Tests

```go
func TestWebAdapter_Button(t *testing.T) {
    adapter := web.NewHTMLAdapter()
    nodes, err := adapter.Ingest(context.Background(), []string{
        "../../examples/example-astro/src/pages/index.astro",
    })
    
    if err != nil {
        t.Fatalf("Ingest failed: %v", err)
    }
    
    // Assert expected node count
    if len(nodes) < 10 {
        t.Errorf("Expected dashboard to have >10 nodes, got %d", len(nodes))
    }
}
```

---

## Contributing Examples

### Guidelines

1. **Keep Files Small:** < 100 lines for clarity
2. **Show Patterns:** Demonstrate common UI components
3. **Include Comments:** Explain accessibility decisions
4. **Test First:** Ensure examples work with current adapters
5. **Platform Diversity:** Show different approaches across platforms

### Submission Process

1. Create example file in appropriate folder
2. Add test case to adapter test suite
3. Update this documentation
4. Submit PR with "examples" prefix:
   ```bash
   git commit -m "examples(android): add Compose button patterns"
   ```

---

## Troubleshooting Examples

### Issue: Example Not Analyzing

**Symptoms:**
```
⚠️  Skipping file: Unsupported file extension.
```

**Solution:** Check file extension mapping in `cli/main.go`:
```go
case ".kt", ".xml":
    adapter = android.NewAndroidAdapter()
```

### Issue: False Positives

**Symptoms:**
- Example marked as violation but follows best practices

**Solution:** Verify adapter is correctly parsing platform-specific syntax. May need to update adapter logic.

### Issue: Example Too Complex

**Symptoms:**
- Analysis takes > 100ms
- Hard to understand what's being tested

**Solution:** Simplify example to focus on single accessibility pattern. Create separate file for complex scenarios.

---

## Resources

- [WCAG 2.2 Quick Reference](https://www.w3.org/WAI/WCAG22/quickref/)
- [Platform Accessibility Guides](#platform-specific-guides)
- [Developer Guide](./DEVELOPER_GUIDE.md)

### Platform-Specific Guides

- [Web Accessibility (MDN)](https://developer.mozilla.org/en-US/docs/Web/Accessibility)
- [Android Accessibility](https://developer.android.com/guide/topics/ui/accessibility)
- [iOS Accessibility](https://developer.apple.com/documentation/accessibility)
- [Flutter Accessibility](https://docs.flutter.dev/development/ui/accessibility)
- [React Native Accessibility](https://reactnative.dev/docs/accessibility)
