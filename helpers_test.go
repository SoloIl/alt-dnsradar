package main

import (
	"testing"
	"time"
)

func TestCleanURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "bare domain",
			in:   "instagram.com",
			want: "instagram.com",
		},
		{
			name: "https url with path",
			in:   "https://www.instagram.com/explore",
			want: "www.instagram.com",
		},
		{
			name: "url with port",
			in:   "http://example.com:8080/path",
			want: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanURL(tt.in)
			if err != nil {
				t.Fatalf("cleanURL(%q) returned error: %v", tt.in, err)
			}

			if got != tt.want {
				t.Fatalf("cleanURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	in := []string{"1.1.1.1", "8.8.8.8", "1.1.1.1", "9.9.9.9", "8.8.8.8"}
	want := []string{"1.1.1.1", "8.8.8.8", "9.9.9.9"}

	got := removeDuplicates(in)

	if len(got) != len(want) {
		t.Fatalf("removeDuplicates() len = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("removeDuplicates() = %v, want %v", got, want)
		}
	}
}

func TestNormalizeSet(t *testing.T) {
	in := []string{"8.8.8.8", "1.1.1.1", "8.8.8.8", "9.9.9.9"}
	want := []string{"1.1.1.1", "8.8.8.8", "9.9.9.9"}

	got := normalizeSet(in)

	if len(got) != len(want) {
		t.Fatalf("normalizeSet() len = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("normalizeSet() = %v, want %v", got, want)
		}
	}
}

func TestSameSet(t *testing.T) {
	if !sameSet([]string{"8.8.8.8", "1.1.1.1"}, []string{"1.1.1.1", "8.8.8.8"}) {
		t.Fatal("sameSet() = false, want true for equal sets")
	}

	if sameSet([]string{"1.1.1.1"}, []string{"8.8.8.8"}) {
		t.Fatal("sameSet() = true, want false for different sets")
	}
}

func TestHasOverlap(t *testing.T) {
	if !hasOverlap([]string{"1.1.1.1", "8.8.8.8"}, []string{"8.8.8.8", "9.9.9.9"}) {
		t.Fatal("hasOverlap() = false, want true")
	}

	if hasOverlap([]string{"1.1.1.1"}, []string{"9.9.9.9"}) {
		t.Fatal("hasOverlap() = true, want false")
	}
}

func TestCompareIPSetsDetailed(t *testing.T) {
	tests := []struct {
		name  string
		left  []string
		right []string
		want  SetRelation
	}{
		{
			name:  "unavailable when one side empty",
			left:  nil,
			right: []string{"1.1.1.1"},
			want:  RelationUnavailable,
		},
		{
			name:  "exact match",
			left:  []string{"8.8.8.8", "1.1.1.1"},
			right: []string{"1.1.1.1", "8.8.8.8"},
			want:  RelationExactMatch,
		},
		{
			name:  "partial overlap",
			left:  []string{"1.1.1.1", "8.8.8.8"},
			right: []string{"8.8.8.8", "9.9.9.9"},
			want:  RelationPartialOverlap,
		},
		{
			name:  "no overlap",
			left:  []string{"1.1.1.1"},
			right: []string{"9.9.9.9"},
			want:  RelationNoOverlap,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareIPSetsDetailed("left", tt.left, "right", tt.right)
			if got.Relation != tt.want {
				t.Fatalf("compareIPSetsDetailed() relation = %v, want %v", got.Relation, tt.want)
			}
		})
	}
}

func TestMedianDuration(t *testing.T) {
	odd := []time.Duration{
		48 * time.Millisecond,
		46 * time.Millisecond,
		120 * time.Millisecond,
	}

	if got := medianDuration(odd); got != 48*time.Millisecond {
		t.Fatalf("medianDuration(odd) = %v, want %v", got, 48*time.Millisecond)
	}

	even := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
	}

	if got := medianDuration(even); got != 25*time.Millisecond {
		t.Fatalf("medianDuration(even) = %v, want %v", got, 25*time.Millisecond)
	}
}

func TestParseASN(t *testing.T) {
	if got := parseASN("AS32934 Facebook, Inc."); got != "AS32934" {
		t.Fatalf("parseASN() = %q, want %q", got, "AS32934")
	}

	if got := parseASN("Facebook, Inc."); got != "-" {
		t.Fatalf("parseASN() = %q, want %q", got, "-")
	}
}

func TestDetectCDN(t *testing.T) {
	tests := []struct {
		org  string
		want string
	}{
		{org: "Facebook, Inc.", want: "Meta"},
		{org: "Google LLC", want: "Google"},
		{org: "Cloudflare, Inc.", want: "Cloudflare"},
		{org: "Unknown Provider", want: "-"},
	}

	for _, tt := range tests {
		if got := detectCDN(tt.org); got != tt.want {
			t.Fatalf("detectCDN(%q) = %q, want %q", tt.org, got, tt.want)
		}
	}
}

func TestValueOrDash(t *testing.T) {
	if got := valueOrDash("Amsterdam"); got != "Amsterdam" {
		t.Fatalf("valueOrDash(non-empty) = %q, want %q", got, "Amsterdam")
	}

	if got := valueOrDash("   "); got != "-" {
		t.Fatalf("valueOrDash(blank) = %q, want %q", got, "-")
	}
}
