# ADR-0008: English Phoneme Inventory and Feature Schema v0.1-en

## Status

Accepted.

## Context

ADR-0002 defines the representation machinery: neutral phonemes, opaque feature masks, and versioned schemas. This ADR defines the initial English inventory and feature schema used by current Gofonix releases.

The schema must be deterministic because golden records and compression artifacts store feature-mask values. A mask is interpretable only together with its feature schema version.

## Decision

Gofonix defines the English feature schema version:

```text
v0.1-en
```

This version is reported as `FeatureSchemaVersion` in result metadata and as `feature_schema_version` in golden records.

The schema includes:

- neutral phoneme IDs for English ARPAbet symbols;
- a boundary phoneme ID;
- 23 feature bits;
- feature-mask assignments for each phoneme.

ARPAbet remains internal. Neutral phoneme IDs are the public identity.

## Phoneme inventory

| ARPAbet | ID | Kind | Description |
| --- | ---: | --- | --- |
| AA | 1 | vowel | low back vowel |
| AE | 2 | vowel | low front vowel |
| AH | 3 | vowel | mid central vowel |
| AO | 4 | vowel | low-mid back rounded vowel |
| AW | 5 | vowel | diphthong |
| AY | 6 | vowel | diphthong |
| B | 7 | consonant | bilabial stop |
| CH | 8 | consonant | affricate |
| D | 9 | consonant | alveolar stop |
| DH | 10 | consonant | dental/alveolar fricative |
| EH | 11 | vowel | mid front vowel |
| ER | 12 | vowel | rhotic vowel |
| EY | 13 | vowel | diphthong |
| F | 14 | consonant | labiodental fricative |
| G | 15 | consonant | velar stop |
| HH | 16 | consonant | glottal fricative |
| M | 17 | consonant | bilabial nasal |
| IY | 18 | vowel | high front tense vowel |
| JH | 19 | consonant | affricate |
| K | 20 | consonant | velar stop |
| L | 21 | consonant | lateral approximant |
| IH | 22 | vowel | high front lax vowel |
| N | 23 | consonant | alveolar nasal |
| NG | 24 | consonant | velar nasal |
| OW | 25 | vowel | rounded diphthong |
| OY | 26 | vowel | rounded diphthong |
| P | 27 | consonant | bilabial stop |
| R | 28 | consonant | approximant |
| S | 29 | consonant | alveolar fricative |
| SH | 30 | consonant | palatal fricative |
| T | 31 | consonant | alveolar stop |
| TH | 32 | consonant | dental/alveolar fricative |
| UH | 33 | vowel | high back lax rounded vowel |
| UW | 34 | vowel | high back tense rounded vowel |
| V | 35 | consonant | labiodental fricative |
| W | 36 | consonant | labiovelar approximant |
| Y | 37 | consonant | palatal approximant |
| Z | 38 | consonant | alveolar fricative |
| ZH | 39 | consonant | palatal fricative |
| boundary | 0 | boundary | reserved boundary phoneme |

Stress digits in ARPAbet input are stripped before mapping to neutral phoneme IDs.

## Feature bits

| Bit | Name | Applies to | Description |
| ---: | --- | --- | --- |
| 0 | voiced | consonant, vowel | voiced phonation |
| 1 | nasal | consonant | nasal airflow |
| 2 | stop | consonant | complete oral closure |
| 3 | fricative | consonant | turbulent airflow |
| 4 | affricate | consonant | stop plus fricative release |
| 5 | approximant | consonant | approximant articulation |
| 6 | lateral | consonant | lateral articulation |
| 7 | labial | consonant | labial or labiodental place |
| 8 | alveolar | consonant | alveolar or dental place |
| 9 | palatal | consonant | palatal or postalveolar place |
| 10 | velar | consonant | velar place |
| 11 | glottal | consonant | glottal place |
| 12 | high | vowel | high vowel |
| 13 | mid | vowel | mid vowel |
| 14 | low | vowel | low vowel |
| 15 | front | vowel | front vowel |
| 16 | back | vowel | back vowel |
| 17 | rounded | vowel | lip rounding |
| 18 | tense | vowel | tense vowel |
| 19 | diphthong | vowel | diphthong |
| 20 | rhotic | vowel | r-coloring |
| 21 | syllabic | consonant, vowel | syllabic nucleus |
| 22 | boundary | boundary | abstract boundary phoneme |

Bits 23 through 127 are reserved for future schema versions.

## Feature assignments

The table below lists the low 64-bit word for each phoneme mask. The high word is `0` for all masks in `v0.1-en`.

| ID | Symbol | Bits set | lo |
| ---: | --- | --- | ---: |
| 0 | boundary | 22 | 4194304 |
| 1 | AA | 0, 14, 16, 21 | 2179073 |
| 2 | AE | 0, 14, 15, 21 | 2146305 |
| 3 | AH | 0, 13, 21 | 2105345 |
| 4 | AO | 0, 2, 14, 21 | 2113541 |
| 5 | AW | 0, 14, 16, 19, 21 | 2703361 |
| 6 | AY | 0, 14, 15, 19, 21 | 2670593 |
| 7 | B | 0, 2, 7 | 133 |
| 8 | CH | 4, 9 | 528 |
| 9 | D | 0, 3 | 9 |
| 10 | DH | 0, 3, 8 | 265 |
| 11 | EH | 0, 13, 15, 21 | 2138113 |
| 12 | ER | 0, 13, 20, 21 | 3153921 |
| 13 | EY | 0, 13, 15, 18, 19, 21 | 2924545 |
| 14 | F | 3, 7 | 136 |
| 15 | G | 0, 2 | 5 |
| 16 | HH | 3, 11 | 2056 |
| 17 | M | 0, 1, 7 | 131 |
| 18 | IY | 0, 12, 15, 18, 21 | 2396161 |
| 19 | JH | 0, 4, 9 | 529 |
| 20 | K | 2, 10 | 1028 |
| 21 | L | 0, 5, 6, 8 | 353 |
| 22 | IH | 0, 13, 21 | 2105345 |
| 23 | N | 0, 1, 8 | 259 |
| 24 | NG | 0, 1, 10 | 1027 |
| 25 | OW | 0, 13, 16, 17, 19, 21 | 2826241 |
| 26 | OY | 0, 13, 16, 17, 19, 21 | 2826241 |
| 27 | P | 2, 7 | 132 |
| 28 | R | 0, 5, 8 | 289 |
| 29 | S | 3, 8 | 264 |
| 30 | SH | 3, 9 | 520 |
| 31 | T | 2, 8 | 260 |
| 32 | TH | 3, 8 | 264 |
| 33 | UH | 0, 12, 16, 17, 21 | 2297857 |
| 34 | UW | 0, 12, 16, 17, 18, 21 | 2560001 |
| 35 | V | 0, 3, 7 | 137 |
| 36 | W | 0, 5, 7, 10 | 1185 |
| 37 | Y | 0, 5, 9 | 545 |
| 38 | Z | 0, 3, 8 | 265 |
| 39 | ZH | 0, 3, 9 | 521 |

## Boundary phoneme and zero mask

Boundary phoneme ID 0 has a non-zero mask:

```json
{"lo": 4194304, "hi": 0}
```

The zero mask is different:

```json
{"lo": 0, "hi": 0}
```

The zero mask means no phoneme governs a byte. It is used for punctuation, whitespace, numbers without pronunciation expansion, symbols, unknown tokens, and current causal scaffold output. It does not mean boundary phoneme ID 0 is present.

## Schema versioning rules

Any of the following requires a new feature schema version:

- adding or removing a feature bit;
- changing a bit position;
- changing a phoneme-to-feature assignment;
- changing phoneme IDs;
- changing the interpretation of the zero mask or boundary phoneme.

Masks produced under different schema versions must not be mixed in the same analysis without explicit conversion.

## Alternatives considered

### Include stress as a feature

Rejected for `v0.1-en`. Stress digits are stripped before mapping. Stress can be added later through a new schema version.

### Separate dental and alveolar place

Rejected for `v0.1-en`. Dental fricatives share the alveolar/dental place bit to keep the first schema compact.

### Use per-language local IDs

Rejected for the current public model. A single neutral ID namespace is simpler for persisted artifacts.

## Consequences

- Golden records can pin exact mask values.
- CLI JSON and golden JSONL can expose feature masks as `{lo,hi}`.
- Several phonemes share an identical feature mask in this compact schema while remaining distinct phoneme IDs: OW (25) and OY (26) both `2826241`; AH (3) and IH (22) both `2105345`; S (29) and TH (32) both `264`. Distinctness is preserved by phoneme ID, not by mask.
- The schema leaves substantial reserved capacity for future languages and feature refinements.

