#!/usr/bin/env python3
"""Generate the Gofonix benchmark fixtures.

This produces ORIGINAL synthetic English prose written for the Gofonix project.
Nothing is copied from the internet. The text is neutral/technical, describing a
fictional grapheme-to-phoneme pipeline, so it exercises the tokenizer and
dictionary path with realistic word/punctuation/number/whitespace mixes.

Outputs:
  testdata/fixtures/bench_ascii.txt    (~10 KB, ASCII only)
  testdata/fixtures/bench_unicode.txt  (~2 KB, curly quotes/dashes/accents)
"""
import os

HERE = os.path.dirname(os.path.abspath(__file__))
FIX = os.path.join(HERE, "..", "testdata", "fixtures")

# Original sentence templates about a fictional phonetic engine. Plain ASCII.
ASCII_SENTENCES = [
    "The engine reads each input string one byte at a time and never assumes a fixed width.",
    "A tokenizer groups letters into words, keeps whitespace intact, and isolates punctuation.",
    "When a word is found in the dictionary, the system emits a neutral phoneme sequence.",
    "Out of vocabulary words may use deterministic fallback rules before becoming unknown.",
    "Each phoneme carries a compact feature mask that records its phonological properties.",
    "The feature stream has exactly one entry per input byte, which keeps alignment simple.",
    "Numbers such as 42, 1024, and 3 are classified as numeric runs rather than words.",
    "Mixed tokens like MP3 and 3rd are treated carefully so the output stays predictable.",
    "Hyphenated forms such as well-known split into a word, a hyphen, and another word.",
    "The batch mode sees the whole text, while the causal mode may use only the prefix.",
    "Reproducibility matters, so every result records the version of the dictionary used.",
    "A checksum over the dictionary file lets two machines confirm they share the same data.",
    "The alignment maps phonemes back onto byte ranges using a uniform distribution rule.",
    "Boundary symbols exist in the inventory but never leak into the projected feature stream.",
    "Determinism is a hard requirement: the same input always yields the same output.",
    "Concurrency is safe because the engine holds no mutable state after construction.",
    "The parser skips comment lines, blank lines, and any record it cannot understand.",
    "Stress markers from the source notation are stripped before the phonemes are mapped.",
    "Variant selection is deterministic, so the canonical pronunciation is selected.",
    "Performance is measured with benchmarks that report allocations but set no thresholds.",
    "Engineers can extend the inventory later without breaking the opaque mask contract.",
    "The cat sat on the mat while the dog watched it from across the quiet room.",
    "Hello there; the well known phrase appears again, and again, across many lines.",
    "Punctuation marks, including commas, periods, and semicolons, anchor the byte spans.",
    "A robust pipeline favors clear contracts over clever shortcuts that resist testing.",
]


def build_ascii(target_bytes: int) -> str:
    out = []
    size = 0
    i = 0
    para = []
    sentences_in_para = 0
    while size < target_bytes:
        s = ASCII_SENTENCES[i % len(ASCII_SENTENCES)]
        para.append(s)
        size += len(s) + 1
        sentences_in_para += 1
        if sentences_in_para >= 4:
            block = " ".join(para) + "\n\n"
            out.append(block)
            para = []
            sentences_in_para = 0
        i += 1
    if para:
        out.append(" ".join(para) + "\n")
    text = "".join(out)
    return text


# Unicode fixture: curly quotes U+2018/U+2019/U+201C/U+201D, em-dash U+2014,
# en-dash U+2013, and accented Latin letters. Original prose.
UNICODE_PARAGRAPHS = [
    "The note said, \u201cthis pipeline is well\u2013behaved\u201d \u2014 and the result was reproducible.",
    "A na\u00efve approach would ignore the caf\u00e9 sign\u2019s accents, but our tokenizer keeps every byte.",
    "The phrase is \u2018obvious\u2019 once the spans are visible; the r\u00e9sum\u00e9 of changes covered 2018\u20132024.",
    "The fa\u00e7ade of simplicity hides care: \u201cdon\u2019t\u201d and \u201cit\u2019s\u201d use a curly apostrophe here.",
    "From na\u00efvet\u00e9 to expertise \u2014 a long road \u2013 the system kept the contract opaque and stable.",
    "The motto was \u2018measure twice\u2019, followed by the practical reminder, \u201cthen measure again.\u201d",
]


def build_unicode(target_bytes: int) -> str:
    out = []
    size = 0
    i = 0
    para = []
    n = 0
    while size < target_bytes:
        s = UNICODE_PARAGRAPHS[i % len(UNICODE_PARAGRAPHS)]
        para.append(s)
        size += len(s.encode("utf-8")) + 1
        n += 1
        if n >= 3:
            out.append(" ".join(para) + "\n\n")
            para = []
            n = 0
        i += 1
    if para:
        out.append(" ".join(para) + "\n")
    return "".join(out)


def main():
    os.makedirs(FIX, exist_ok=True)

    ascii_text = build_ascii(10 * 1024)
    ascii_path = os.path.join(FIX, "bench_ascii.txt")
    with open(ascii_path, "w", encoding="ascii") as f:
        f.write(ascii_text)
    assert ascii_text.isascii(), "ascii fixture must be ASCII-only"

    uni_text = build_unicode(2 * 1024)
    uni_path = os.path.join(FIX, "bench_unicode.txt")
    with open(uni_path, "w", encoding="utf-8") as f:
        f.write(uni_text)

    print("bench_ascii.txt   bytes:", len(ascii_text.encode("ascii")))
    print("bench_unicode.txt bytes:", len(uni_text.encode("utf-8")))


if __name__ == "__main__":
    main()
