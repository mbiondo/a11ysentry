package domain

import (
	"testing"
)

func TestUSNCreation(t *testing.T) {
	// GIVEN a need to represent a UI element
	// WHEN a USN instance is created
	usn := USN{
		UID:   "test-uid",
		Role:  RoleButton,
		Label: "Click Me",
		State: USNState{
			Disabled: false,
		},
		Source: Source{
			Platform: PlatformWebReact,
		},
	}

	// THEN it MUST include core fields
	if usn.UID != "test-uid" {
		t.Errorf("expected UID 'test-uid', got %s", usn.UID)
	}
	if usn.Role != RoleButton {
		t.Errorf("expected Role 'button', got %s", usn.Role)
	}
	if usn.Label != "Click Me" {
		t.Errorf("expected Label 'Click Me', got %s", usn.Label)
	}
	if usn.Source.Platform != PlatformWebReact {
		t.Errorf("expected Platform 'WEB_REACT', got %s", usn.Source.Platform)
	}
}

