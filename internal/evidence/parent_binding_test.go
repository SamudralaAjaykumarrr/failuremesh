package evidence

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

var bindingParent = []byte("abcdefghijkl")

func bindingCapture(t *testing.T, id string, start, end, length int64, text, whole string) Record {
	t.Helper()
	// Reuse only the metadata scaffold. The exact source/locations are defined
	// independently here, including deliberately untruthful reviewer inputs.
	r := excerptCapture(t, 0, 3, 1, id, whole)
	a := &r.Artifacts[0]
	a.ContentDigest = Hash(bindingParent)
	a.ByteLength = length
	l := &a.Locations[0]
	l.ArtifactDigest = a.ContentDigest
	l.Selector.Start = start
	l.Selector.End = end
	l.Excerpt = []byte(text)
	d := Hash(l.Excerpt)
	l.ExcerptDigest = &d
	return r
}

func TestB1ParentBindingReviewerAttacks(t *testing.T) {
	for _, attack := range []string{"forged_range", "forged_length"} {
		t.Run(attack, func(t *testing.T) {
			start, end, length := int64(0), int64(6), int64(12)
			expected := "excerpt_parent_slice_mismatch"
			if attack == "forged_length" {
				start = 6
				end = 12
				length = 13
				expected = "parent_length_mismatch"
			}
			first := bindingCapture(t, "alias.first", 0, 6, length, "abcdef", "deny")
			second := bindingCapture(t, "alias.second", start, end, length, "ghijkl", "deny")
			combined := composedRecords(first, second)
			for _, available := range []bool{false, true} {
				var parents [][]byte
				reason := "excerpt_parent_unavailable"
				if available {
					parents = [][]byte{bindingParent}
					reason = expected
				}
				if err := Validate(combined, parents...); err == nil || err.Error() != reason {
					t.Fatalf("Validate parent=%v: %v", available, err)
				}
				if _, _, err := Save(t.TempDir(), combined, fixtureTime, parents...); err == nil || err.Error() != reason {
					t.Fatalf("Save parent=%v: %v", available, err)
				}
				if out, err := exportRecord(combined, parents...); err == nil || out != nil || err.Error() != reason {
					t.Fatalf("Export parent=%v: %v", available, err)
				}
				root := t.TempDir()
				_, _, firstErr := Save(root, first, fixtureTime, parents...)
				if available && attack == "forged_range" {
					if firstErr != nil {
						t.Fatal("valid first half", firstErr)
					}
				} else if firstErr == nil {
					t.Fatal("unverified first capture retained")
				}
				if _, _, err := Save(root, second, fixtureTime, parents...); err == nil || err.Error() != reason {
					t.Fatalf("successive Save: %v", err)
				}
			}
		})
	}
}

func TestB1VerifiedParentRanges(t *testing.T) {
	tests := []struct {
		name               string
		start, end, length int64
		text               string
		valid              bool
	}{
		{"partial", 0, 3, 12, "abc", true},
		{"false slice with true excerpt digest", 0, 3, 12, "ghi", false},
		{"past actual parent end", 10, 13, 12, "klm", false},
		{"negative", -1, 2, 12, "abc", false},
		{"empty", 0, 0, 12, "", false},
		{"reversed", 3, 1, 12, "ab", false},
		{"false parent length", 0, 3, 13, "abc", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bindingCapture(t, "range", tt.start, tt.end, tt.length, tt.text, "deny")
			if err := Validate(r, bindingParent); (err == nil) != tt.valid {
				t.Fatalf("Validate: %v", err)
			}
			if _, _, err := Save(t.TempDir(), r, fixtureTime, bindingParent); (err == nil) != tt.valid {
				t.Fatalf("Save: %v", err)
			}
			if out, err := exportRecord(r, bindingParent); (err == nil) != tt.valid || (!tt.valid && out != nil) {
				t.Fatalf("Export: %v", err)
			}
		})
	}
	// Incorrect parent content/digest, even when the excerpt's own digest is true.
	r := bindingCapture(t, "digest", 0, 3, 12, "abc", "deny")
	if Validate(r, []byte("xxxxxxxxxxxx")) == nil {
		t.Fatal("unrelated parent accepted")
	}
	r.Artifacts[0].ContentDigest = Hash([]byte("xxxxxxxxxxxx"))
	r.Artifacts[0].Locations[0].ArtifactDigest = r.Artifacts[0].ContentDigest
	if Validate(r, bindingParent) == nil {
		t.Fatal("false parent digest accepted")
	}
	r = bindingCapture(t, "relationship", 0, 3, 12, "abc", "deny")
	r.Artifacts[0].Locations[0].ArtifactID = "different-artifact"
	if Validate(r, bindingParent) == nil {
		t.Fatal("false artifact relationship accepted")
	}
}

func TestB1VerifiedParentCoverage(t *testing.T) {
	tests := []struct {
		name        string
		start, end  int64
		text, whole string
		valid       bool
	}{
		{"gaps", 6, 9, "ghi", "deny", true},
		{"overlap", 1, 4, "bcd", "deny", true},
		{"duplicate", 0, 3, "abc", "deny", true},
		{"complement forbidden", 3, 12, "defghijkl", "deny", false},
		{"complement permitted", 3, 12, "defghijkl", "allow", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first := bindingCapture(t, "first", 0, 3, 12, "abc", tt.whole)
			second := bindingCapture(t, "second", tt.start, tt.end, 12, tt.text, tt.whole)
			combined := composedRecords(first, second)
			if err := Validate(combined, bindingParent); (err == nil) != tt.valid {
				t.Fatalf("Validate: %v", err)
			}
			if out, err := exportRecord(combined, bindingParent); (err == nil) != tt.valid || (!tt.valid && out != nil) {
				t.Fatalf("Export: %v", err)
			}
			root := t.TempDir()
			if _, _, err := Save(root, first, fixtureTime, bindingParent); err != nil {
				t.Fatal(err)
			}
			if _, _, err := Save(root, second, fixtureTime, bindingParent); (err == nil) != tt.valid {
				t.Fatalf("successive Save: %v", err)
			}
		})
	}
}

func TestB1ParentUnavailableAndTransientReplay(t *testing.T) {
	r := bindingCapture(t, "partial", 0, 3, 12, "abc", "deny")
	if Validate(r) == nil {
		t.Fatal("unverified partial accepted")
	}
	root := t.TempDir()
	rd, ad, err := Save(root, r, fixtureTime, bindingParent)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Read(root, rd); err == nil {
		t.Fatal("lost parent treated as verified")
	}
	replay, err := Read(root, rd, bindingParent)
	if err != nil {
		t.Fatal(err)
	}
	assessment, _ := json.Marshal(Assess(replay, bindingParent))
	if Hash(assessment) != ad {
		t.Fatal("replay changed")
	}
	exported, err := exportRecord(r, bindingParent)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(exported, []byte(base64.StdEncoding.EncodeToString(bindingParent))) {
		t.Fatal("transient body exported")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		raw, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(raw, bindingParent) || bytes.Contains(raw, []byte(base64.StdEncoding.EncodeToString(bindingParent))) {
			t.Fatal("transient body persisted", entry.Name())
		}
	}
	// Revalidate persisted bindings too, never just the new candidate's excerpt.
	bad := bindingCapture(t, "forged-prior", 0, 3, 12, "ghi", "deny")
	raw, _ := json.Marshal(bad.Artifacts[0])
	priorRoot := t.TempDir()
	if err := publish(priorRoot, artifactBindingName(bad.Artifacts[0].ArtifactID), raw); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Save(priorRoot, r, fixtureTime, bindingParent); err == nil || err.Error() != "excerpt_parent_slice_mismatch" {
		t.Fatalf("forged persisted binding accepted: %v", err)
	}
}

func TestB1RetainedParentBindsExcerpts(t *testing.T) {
	r := bindingCapture(t, "retained", 0, 3, 12, "abc", "allow")
	a := &r.Artifacts[0]
	a.Retained = true
	a.Body = bytes.Clone(bindingParent)
	if err := Validate(r); err != nil {
		t.Fatal(err)
	}
	a.Locations[0].Excerpt = []byte("ghi")
	d := Hash(a.Locations[0].Excerpt)
	a.Locations[0].ExcerptDigest = &d
	if Validate(r) == nil {
		t.Fatal("retained parent mismatch accepted")
	}
	a.Locations[0].Excerpt = []byte("abc")
	d = Hash(a.Locations[0].Excerpt)
	a.Locations[0].ExcerptDigest = &d
	a.ByteLength = 13
	if Validate(r) == nil {
		t.Fatal("retained parent length mismatch accepted")
	}
	a.ByteLength = 12
	a.ContentDigest = Hash([]byte("xxxxxxxxxxxx"))
	a.Locations[0].ArtifactDigest = a.ContentDigest
	if Validate(r) == nil {
		t.Fatal("retained parent digest mismatch accepted")
	}
}
