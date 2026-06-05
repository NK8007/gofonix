package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// TestCLIVersion exercises the --version banner. It must:
//   - exit with code 0,
//   - print the CLI version string,
//   - mention the engine GofonixVersion "v0.1.0" (per the Slice 3 brief),
//   - mention the schema version "gofonix-result-v0.1".
//
// The exact line layout is left to main.go; the test only checks substrings
// so a future cosmetic change to the banner does not falsely break it.
func TestCLIVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--version"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("--version exited with code %d (stderr=%q); want 0", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "v0.1.0") {
		t.Fatalf("--version output does not contain engine version \"v0.1.0\":\n%s", out)
	}
	if !strings.Contains(out, "gofonix-result-v0.1") {
		t.Fatalf("--version output does not contain schema version \"gofonix-result-v0.1\":\n%s", out)
	}
	if !strings.Contains(out, "gofonix-cli ") {
		t.Fatalf("--version output does not contain CLI banner \"gofonix-cli \":\n%s", out)
	}
}

// TestCLIBasicJSON drives the CLI with --input "the cat" and checks that the
// emitted JSON parses cleanly, includes a populated Tokens array, and
// surfaces the ResultMetadata fields required by ADR-0010 (DictionaryID,
// FullDictAvailable). This is a smoke test: it does NOT pin phoneme values
// (those are guarded by the engine's golden tests).
func TestCLIBasicJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"--input", "the cat", "--output", "json"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("--input \"the cat\" --output json exited with code %d (stderr=%q); want 0", code, stderr.String())
	}

	// The JSON encoder appends a trailing newline; both json.Unmarshal and
	// json.Decoder accept that, so we just hand the buffer to Decoder.
	var parsed map[string]any
	if err := json.NewDecoder(&stdout).Decode(&parsed); err != nil {
		t.Fatalf("CLI emitted non-JSON output: %v", err)
	}

	// Top-level Result keys we contractually surface (ADR-0001 result-record
	// schema). We check for presence, not value equality — that is the job of
	// the engine-level golden tests.
	for _, key := range []string{"Input", "Language", "Mode", "Tokens", "FeatureStream", "SchemaVersion", "Metadata"} {
		if _, ok := parsed[key]; !ok {
			t.Fatalf("CLI JSON missing top-level key %q; parsed=%v", key, parsed)
		}
	}

	// Tokens must be a non-empty array for the input "the cat".
	tokens, ok := parsed["Tokens"].([]any)
	if !ok {
		t.Fatalf("CLI JSON \"Tokens\" is not an array: %T", parsed["Tokens"])
	}
	if len(tokens) == 0 {
		t.Fatalf("CLI JSON \"Tokens\" is empty for input \"the cat\"; want at least one token")
	}

	// Metadata must be an object and must carry the ADR-0010 fields.
	meta, ok := parsed["Metadata"].(map[string]any)
	if !ok {
		t.Fatalf("CLI JSON \"Metadata\" is not an object: %T", parsed["Metadata"])
	}
	for _, key := range []string{
		"GofonixVersion",
		"FeatureSchemaVersion",
		"NormalizerVersion",
		"TokenizerVersion",
		"OOVPolicy",
		"FallbackRulesVersion",
		"DictionaryID",
		"DictionaryChecksum",
		"FullDictAvailable",
	} {
		if _, ok := meta[key]; !ok {
			t.Fatalf("CLI JSON Metadata missing required key %q (ADR-0010, Public API Impact); meta=%v", key, meta)
		}
	}

	// In the default build there is no full dict, so FullDictAvailable must
	// be false and DictionaryID must be the mini-dict ID.
	if v, ok := meta["FullDictAvailable"].(bool); !ok || v {
		t.Fatalf("default build expected FullDictAvailable=false, got %v (type %T)", meta["FullDictAvailable"], meta["FullDictAvailable"])
	}
	if v, _ := meta["DictionaryID"].(string); v != "cmudict-mini-v0.1" {
		t.Fatalf("default build expected DictionaryID=\"cmudict-mini-v0.1\", got %q", v)
	}
}

// TestCLIInvalidMode confirms that an unknown --mode value causes exit code 1
// and prints a diagnostic on stderr (per the brief: "Nieznany --mode skutkuje
// os.Exit(1) z komunikatem błędu na stderr").
func TestCLIInvalidMode(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"--input", "hi", "--mode", "invalid"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if code != 1 {
		t.Fatalf("--mode invalid exited with code %d (stderr=%q); want 1", code, stderr.String())
	}
	if !strings.Contains(strings.ToLower(stderr.String()), "invalid") {
		t.Fatalf("--mode invalid did not surface a diagnostic on stderr: %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("--mode invalid wrote to stdout (%q); errors must go to stderr only", stdout.String())
	}
}

// TestCLIStdin verifies that an empty --input causes the CLI to read from
// stdin. We use an os.Pipe (per the brief) rather than strings.NewReader so
// the test exercises the same Reader contract a real shell pipe would.
func TestCLIStdin(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	// Write the payload from a goroutine so we don't deadlock on a pipe
	// buffer; close the writer to signal EOF to io.ReadAll on the reader
	// side.
	payload := "the dog"
	done := make(chan error, 1)
	go func() {
		_, werr := io.WriteString(w, payload)
		if cerr := w.Close(); werr == nil {
			werr = cerr
		}
		done <- werr
	}()

	var stdout, stderr bytes.Buffer
	code := run([]string{"--output", "text"}, r, &stdout, &stderr)
	if werr := <-done; werr != nil {
		t.Fatalf("pipe write failed: %v", werr)
	}
	if code != 0 {
		t.Fatalf("stdin pipeline exited with code %d (stderr=%q); want 0", code, stderr.String())
	}
	if stdout.Len() == 0 {
		t.Fatalf("stdin pipeline produced empty stdout; want at least one text line for payload %q", payload)
	}
	// In text mode each line ends with a newline and is tab-separated; check
	// that the first token surfaces the input verbatim.
	firstLine := strings.SplitN(stdout.String(), "\n", 2)[0]
	cols := strings.Split(firstLine, "\t")
	if len(cols) != 3 {
		t.Fatalf("text-mode line is not 3 tab-separated columns: %q", firstLine)
	}
	if cols[0] != "the" {
		t.Fatalf("first token from stdin pipeline is %q; want \"the\"", cols[0])
	}
}
