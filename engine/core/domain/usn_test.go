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

func TestOpaqueNodeCreation(t *testing.T) {
	// GIVEN an external component instantiation
	// WHEN a USN marked as opaque is created
	usn := USN{
		UID:   "opaque-btn",
		Role:  "generic",
		IsOpaque: true,
		Source: Source{
			Platform:     PlatformWebReact,
			IsOpaque:     true,
			OpaqueSource: "@mui/material",
		},
	}

	// THEN it MUST reflect the opacity metadata
	if !usn.IsOpaque {
		t.Error("expected USN to be marked as opaque")
	}
	if !usn.Source.IsOpaque {
		t.Error("expected Source to be marked as opaque")
	}
	if usn.Source.OpaqueSource != "@mui/material" {
		t.Errorf("expected OpaqueSource '@mui/material', got %s", usn.Source.OpaqueSource)
	}
}

