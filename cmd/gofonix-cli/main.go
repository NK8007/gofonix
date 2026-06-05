// gofonix-cli is the v0.3 command-line entry point for the Gofonix G2P engine.
//
// Scope (Phase 3 / Slice 3):
//   - stdlib-only (flag, encoding/json, io, os, fmt, strings).
//   - Reads input from --input or, when --input is empty, from os.Stdin.
//   - Constructs a g2p.Engine via the public pkg/g2p API ONLY.
//   - Emits either an indented JSON Result or a tab-separated text summary.
//   - Honours the v0.3 --dict-path override (ADR-0010 Decision §3); the engine
//     falls back gracefully to the embedded mini-dict whenever the full
//     CMUdict cannot be loaded.
//
// Explicitly out of scope (Slice 3 brief): cobra/viper or any third-party
// flag parser, streaming of large inputs, Polish or any non-English module,
// compression / bits-per-byte reporting. The CLI is a thin wrapper over the
// public engine and adds no domain logic.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gofonix/gofonix/pkg/g2p"
	"github.com/gofonix/gofonix/pkg/phoneme"
)

// cliVersion identifies the CLI surface independently of the engine version.
// The engine's GofonixVersion (gofonixVersion in pkg/g2p) is reported
// separately so a user can tell which library a CLI build is linked against.
const cliVersion = "v0.3.0-alpha"

// resultSchemaVersion mirrors the schemaVersion constant pinned in
// pkg/g2p/engine.go. It is duplicated here ONLY so --version can print it
// without constructing an engine; if the engine constant drifts a test will
// catch it (see TestVersionContainsSchemaVersion in main_test.go).
const resultSchemaVersion = "gofonix-result-v0.1"

// engineGofonixVersion mirrors the gofonixVersion constant pinned in
// pkg/g2p/engine.go for the same reason as resultSchemaVersion. Tests assert
// that the two stay in lockstep with the values surfaced via ResultMetadata.
const engineGofonixVersion = "v0.1.0"

// exitCode values are split out so the smoke-test harness can assert them
// without relying on raw integer literals scattered through main.
const (
	exitOK      = 0
	exitFailure = 1
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run is the testable entry point. It receives the argv tail (without the
// program name), an input reader (os.Stdin in production, an os.Pipe in
// tests), and writers for stdout/stderr. It returns the process exit code
// instead of calling os.Exit directly so tests can drive it in-process.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gofonix-cli", flag.ContinueOnError)
	// Send flag-parser errors to our stderr so smoke tests can capture them
	// deterministically; the default would be os.Stderr which would bypass
	// the test's captured buffer.
	fs.SetOutput(stderr)

	var (
		langFlag    = fs.String("language", "en", "BCP-47-ish language tag (v0.1/v0.3 supports only \"en\")")
		modeFlag    = fs.String("mode", "batch", "analysis mode: \"batch\" or \"causal\"")
		inputFlag   = fs.String("input", "", "input text; if empty, read from stdin")
		dictPath    = fs.String("dict-path", "", "optional path to the full CMUdict v0.7b (ADR-0010); empty triggers GOFONIX_DICT_PATH then $HOME/.gofonix/cmudict.dict")
		outputFlag  = fs.String("output", "json", "output format: \"json\" or \"text\"")
		versionFlag = fs.Bool("version", false, "print version information and exit")
	)

	if err := fs.Parse(args); err != nil {
		// flag.ContinueOnError prints the error and usage; just convert to
		// exit code 1 here.
		return exitFailure
	}

	if *versionFlag {
		// Multi-line, newline-terminated version banner. Each line stands on
		// its own so downstream shells can `grep` for any field.
		fmt.Fprintf(stdout, "gofonix-cli %s\n", cliVersion)
		fmt.Fprintf(stdout, "GofonixVersion: %s\n", engineGofonixVersion)
		fmt.Fprintf(stdout, "SchemaVersion: %s\n", resultSchemaVersion)
		return exitOK
	}

	mode, ok := parseMode(*modeFlag)
	if !ok {
		fmt.Fprintf(stderr, "gofonix-cli: invalid --mode %q (expected \"batch\" or \"causal\")\n", *modeFlag)
		return exitFailure
	}

	opts := g2p.Options{
		Language: *langFlag,
		Mode:     mode,
		DictPath: *dictPath,
	}

	engine, err := g2p.New(opts)
	if err != nil {
		fmt.Fprintf(stderr, "gofonix-cli: engine init failed: %v\n", err)
		return exitFailure
	}

	input, err := resolveInput(*inputFlag, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "gofonix-cli: failed to read input: %v\n", err)
		return exitFailure
	}

	res, err := engine.Process(input)
	if err != nil {
		fmt.Fprintf(stderr, "gofonix-cli: processing failed: %v\n", err)
		return exitFailure
	}

	switch strings.ToLower(strings.TrimSpace(*outputFlag)) {
	case "json":
		if err := writeJSON(stdout, res); err != nil {
			fmt.Fprintf(stderr, "gofonix-cli: json encode failed: %v\n", err)
			return exitFailure
		}
	case "text":
		if err := writeText(stdout, res); err != nil {
			fmt.Fprintf(stderr, "gofonix-cli: text encode failed: %v\n", err)
			return exitFailure
		}
	default:
		fmt.Fprintf(stderr, "gofonix-cli: invalid --output %q (expected \"json\" or \"text\")\n", *outputFlag)
		return exitFailure
	}

	return exitOK
}

// parseMode maps the user-facing --mode string to a g2p.Mode. Only "batch"
// and "causal" are exposed in the CLI (per the Slice 3 brief); "oracle" is a
// library-internal analysis mode and is intentionally not surfaced here.
//
// The match is case-insensitive after trimming surrounding whitespace; this
// is a one-line affordance for shells / Makefiles and does not alter the
// semantics of the underlying engine.
func parseMode(s string) (g2p.Mode, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "batch", "":
		return g2p.ModeBatch, true
	case "causal":
		return g2p.ModeCausal, true
	default:
		return 0, false
	}
}

// resolveInput returns the input text to process. When the --input flag is
// non-empty, its value is used verbatim. When it is empty, the entire
// contents of stdin are read into memory (the Slice 3 brief explicitly
// defers streaming to a later phase).
func resolveInput(inputFlag string, stdin io.Reader) (string, error) {
	if inputFlag != "" {
		return inputFlag, nil
	}
	if stdin == nil {
		return "", nil
	}
	buf, err := io.ReadAll(stdin)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// writeJSON serialises the entire Result as indented JSON. Note: the
// FeatureStream entries are FeatureMask values whose internal bit-array is
// unexported and therefore not surfaced by encoding/json. That is the
// intended v0.3 behaviour — the public mask accessors (Lo, Hi, IsZero) are
// the stable interface and a JSON dump is for human inspection, not for
// round-trip reconstruction of the mask bits.
func writeJSON(w io.Writer, res g2p.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	// json.Encoder appends a final newline, which is the conventional
	// terminator for tool output.
	return enc.Encode(res)
}

// writeText emits one line per token in the format `<token>\t<phonemes>\t<source>`.
// Non-word and unresolved tokens have an empty <phonemes> field but still
// appear as their own line so the count matches the token count exactly.
func writeText(w io.Writer, res g2p.Result) error {
	for _, tr := range res.Tokens {
		line := fmt.Sprintf("%s\t%s\t%s\n",
			tr.Token,
			formatPhonemes(tr.Pronunciation.Phonemes),
			formatSource(tr.Pronunciation.Source),
		)
		if _, err := io.WriteString(w, line); err != nil {
			return err
		}
	}
	return nil
}

// formatPhonemes renders a phoneme sequence as a space-separated list of
// neutral integer IDs. The CLI prints IDs rather than the debug IPA string
// (Phoneme.IPA) because the IPA field is documented as debug-only with no
// stability guarantee.
func formatPhonemes(ps []phoneme.Phoneme) string {
	if len(ps) == 0 {
		return ""
	}
	parts := make([]string, len(ps))
	for i, p := range ps {
		parts[i] = fmt.Sprintf("%d", p.ID)
	}
	return strings.Join(parts, " ")
}

// formatSource renders a g2p.Source as a short, stable string suitable for
// downstream `cut`/`awk` pipelines. The mapping is closed to exactly three
// values; an unrecognised Source (which should not be reachable) is rendered
// as "?" so the output never silently swallows a regression.
func formatSource(s g2p.Source) string {
	switch s {
	case g2p.SourceDict:
		return "dict"
	case g2p.SourceRuleFallback:
		return "fallback"
	case g2p.SourceUnknown:
		return "unknown"
	default:
		return "?"
	}
}
