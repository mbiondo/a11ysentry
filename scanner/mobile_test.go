package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"a11ysentry/scanner"
	"a11ysentry/scanner/android"
	"a11ysentry/scanner/ios"
)

func TestFindProjectRootsMobile(t *testing.T) {
	tmp := t.TempDir()

	// Create a mock Android project
	androidDir := filepath.Join(tmp, "my-android-app")
	if err := os.MkdirAll(androidDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(androidDir, "build.gradle"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a mock iOS project
	iosDir := filepath.Join(tmp, "my-ios-app")
	if err := os.MkdirAll(filepath.Join(iosDir, "MyApp.xcodeproj"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a nested web project to ensure we still find it
	webDir := filepath.Join(tmp, "my-web-app")
	if err := os.MkdirAll(webDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(webDir, "package.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	roots := scanner.FindProjectRoots(tmp)

	expected := map[string]bool{
		androidDir: true,
		iosDir:     true,
		webDir:     true,
	}

	if len(roots) != len(expected) {
		t.Errorf("expected %d roots, got %d", len(expected), len(roots))
	}

	for _, root := range roots {
		if !expected[root] {
			t.Errorf("unexpected root found: %s", root)
		}
	}
}

func TestAndroidScanner(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "build.gradle"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmp, "MainActivity.kt"), []byte("import android.os.Bundle"), 0644)
	os.WriteFile(filepath.Join(tmp, "activity_main.xml"), []byte("<Layout></Layout>"), 0644)

	fw := android.New()
	if !fw.Probe(tmp) {
		t.Fatal("Android scanner should probe positive for gradle project")
	}

	ui, _, err := fw.CollectFiles(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(ui) != 2 {
		t.Errorf("expected 2 UI files, got %d: %v", len(ui), ui)
	}
}

func TestIOSScanner(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "MyApp.xcodeproj"), 0755)
	os.WriteFile(filepath.Join(tmp, "ContentView.swift"), []byte("import SwiftUI"), 0644)

	fw := ios.New()
	if !fw.Probe(tmp) {
		t.Fatal("iOS scanner should probe positive for xcodeproj project")
	}

	ui, _, err := fw.CollectFiles(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(ui) != 1 {
		t.Errorf("expected 1 UI file, got %d: %v", len(ui), ui)
	}
}
