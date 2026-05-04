package domain

// SemanticRole represents the accessibility role of a UI element.
type SemanticRole string

const (
	RoleButton     SemanticRole = "button"
	RoleHeading    SemanticRole = "heading"
	RoleLink       SemanticRole = "link"
	RoleInput      SemanticRole = "input"
	RoleImage      SemanticRole = "image"
	RoleLiveRegion SemanticRole = "live-region"
	RoleModal      SemanticRole = "modal"

	// Landmarks
	RoleMain    SemanticRole = "main"
	RoleNav     SemanticRole = "navigation"
	RoleAside   SemanticRole = "complementary"
	RoleHeader  SemanticRole = "banner"
	RoleFooter  SemanticRole = "contentinfo"
	RoleSection SemanticRole = "region"
	RoleForm    SemanticRole = "form"
	RoleSearch  SemanticRole = "search"

	// Groups
	RoleFieldset SemanticRole = "fieldset"
	RoleLegend   SemanticRole = "legend"
)

// Platform represents the source framework or operating system.
type Platform string

const (
	PlatformWebReact       Platform = "WEB_REACT"
	PlatformWebVue         Platform = "WEB_VUE"
	PlatformWebSvelte      Platform = "WEB_SVELTE"
	PlatformWebAngular     Platform = "WEB_ANGULAR"
	PlatformWebAstro       Platform = "WEB_ASTRO"
	PlatformAndroidCompose Platform = "ANDROID_COMPOSE"
	PlatformAndroidView    Platform = "ANDROID_VIEW"
	PlatformIOSSwiftUI     Platform = "IOS_SWIFTUI"
	PlatformFlutterDart    Platform = "FLUTTER_DART"
	PlatformDotNetXAML     Platform = "DOTNET_XAML"
	PlatformDotNetCSharp   Platform = "DOTNET_CSHARP"
	PlatformJavaFX         Platform = "JAVA_FX"
	PlatformJavaSwing      Platform = "JAVA_SWING"
	PlatformReactNative    Platform = "REACT_NATIVE"
	PlatformBlazor         Platform = "BLAZOR"
	PlatformUnity          Platform = "UNITY"
	PlatformGodot          Platform = "GODOT"
	PlatformElectron       Platform = "ELECTRON"
	PlatformTauri          Platform = "TAURI"
)

// USN (Universal Semantic Node) captures the essential semantic properties
// of a UI element in a platform-agnostic way.
type USN struct {
	UID       string
	Role      SemanticRole
	Label     string
	State     USNState
	Traits    map[string]any
	Geometry  Geometry
	Hierarchy Hierarchy
	Source    Source
}

// USNState represents the interactive state of a UI element.
type USNState struct {
	Disabled bool
	Hidden   bool
	Selected bool
	Expanded bool
	Invalid  bool
}

// Geometry represents the physical bounds of an element.
type Geometry struct {
	X, Y, W, H float64
}

// Hierarchy represents the relationships between USN nodes.
type Hierarchy struct {
	ParentID string
	Children []USN
}

// Source represents the origin information of a USN node.
type Source struct {
	Platform     Platform
	FilePath     string
	Line         int
	Column       int
	RawHTML      string
	IsComponent  bool     // true when the file is a partial component, not a full HTML document
	IgnoredRules []string // list of error codes to ignore for this specific element
}
