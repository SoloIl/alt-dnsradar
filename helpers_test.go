package main

import (
	"reflect"
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

			if got.LeftCount != len(normalizeSet(tt.left)) {
				t.Fatalf("compareIPSetsDetailed() left count = %d, want %d", got.LeftCount, len(normalizeSet(tt.left)))
			}

			if got.RightCount != len(normalizeSet(tt.right)) {
				t.Fatalf("compareIPSetsDetailed() right count = %d, want %d", got.RightCount, len(normalizeSet(tt.right)))
			}
		})
	}
}

func TestDNSSummaryLineForMultiEndpointMismatch(t *testing.T) {
	comparison := DNSComparison{
		LeftName:   "Google UDP",
		RightName:  "Google DoH",
		LeftCount:  4,
		RightCount: 4,
		Relation:   RelationNoOverlap,
	}

	got := dnsSummaryLine(comparison)
	want := "Google UDP and Google DoH returned different multi-endpoint sets; possible cache, CDN variance, or interception"

	if got != want {
		t.Fatalf("dnsSummaryLine(multi-endpoint) = %q, want %q", got, want)
	}
}

func TestDNSSummaryLineForSingleEndpointMismatch(t *testing.T) {
	comparison := DNSComparison{
		LeftName:   "Google UDP",
		RightName:  "Google DoH",
		LeftCount:  1,
		RightCount: 1,
		Relation:   RelationNoOverlap,
	}

	got := dnsSummaryLine(comparison)
	want := "Strong mismatch between Google UDP and Google DoH; possible DNS interception"

	if got != want {
		t.Fatalf("dnsSummaryLine(single-endpoint) = %q, want %q", got, want)
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

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantDomain  string
		wantExtras  []string
		wantFlagOut []string
	}{
		{
			name:        "domain plus flags after it",
			args:        []string{"example.com", "--lang", "ru", "--all"},
			wantDomain:  "example.com",
			wantFlagOut: []string{"--lang", "ru", "--all"},
		},
		{
			name:        "flags before domain",
			args:        []string{"--lang", "ru", "example.com"},
			wantDomain:  "example.com",
			wantFlagOut: []string{"--lang", "ru"},
		},
		{
			name:        "extra positional arg is preserved as error candidate",
			args:        []string{"dnsradar", "instagram.com"},
			wantDomain:  "dnsradar",
			wantExtras:  []string{"instagram.com"},
			wantFlagOut: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, extras, flags := splitArgs(tt.args)

			if domain != tt.wantDomain {
				t.Fatalf("splitArgs() domain = %q, want %q", domain, tt.wantDomain)
			}

			if !reflect.DeepEqual(extras, tt.wantExtras) {
				t.Fatalf("splitArgs() extras = %v, want %v", extras, tt.wantExtras)
			}

			if !reflect.DeepEqual(flags, tt.wantFlagOut) {
				t.Fatalf("splitArgs() flags = %v, want %v", flags, tt.wantFlagOut)
			}
		})
	}
}

func TestLookupWithRetry(t *testing.T) {
	t.Run("returns value from second attempt", func(t *testing.T) {
		calls := 0

		got := lookupWithRetry(func() []string {
			calls++
			if calls == 1 {
				return nil
			}

			return []string{"8.8.8.8"}
		})

		if calls != 2 {
			t.Fatalf("lookupWithRetry() calls = %d, want 2", calls)
		}

		if !reflect.DeepEqual(got, []string{"8.8.8.8"}) {
			t.Fatalf("lookupWithRetry() = %v, want %v", got, []string{"8.8.8.8"})
		}
	})

	t.Run("returns nil after all attempts fail", func(t *testing.T) {
		calls := 0

		got := lookupWithRetry(func() []string {
			calls++
			return nil
		})

		if calls != dnsDiagnosticAttempts {
			t.Fatalf("lookupWithRetry() calls = %d, want %d", calls, dnsDiagnosticAttempts)
		}

		if got != nil {
			t.Fatalf("lookupWithRetry() = %v, want nil", got)
		}
	})
}
