# Data Licenses

This document describes the licensing status of data used by Gofonix.
It is separate from the source-code license.

> **Code vs. data.** The Gofonix source code is licensed under the
> **Apache License, Version 2.0** (see `LICENSE`). Third-party data
> referenced, bundled, or downloaded by this project is **not** automatically
> covered by Apache-2.0. Each dataset keeps its own license and attribution
> requirements, described below. When in doubt, treat data licensing
> conservatively and verify the upstream terms before redistribution.

---

## 1. CMU Pronouncing Dictionary (CMUdict)

- **Upstream:** https://github.com/cmusphinx/cmudict
- **Copyright:** Copyright (C) 1993-2015 Carnegie Mellon University. All rights reserved.
- **License:** BSD 2-Clause style license.
- **License text:** Full license text is reproduced in `THIRD_PARTY_NOTICES.md`
  and in `third_party/cmudict/LICENSE`.
- **Status in repo:** License metadata is committed. CMUdict-derived data may
  be bundled separately in the repository, for example as a reduced embedded
  dictionary used by tests or default builds.
- **Expected bundled paths:**
  - `third_party/cmudict/` — upstream license and attribution metadata
  - `internal/dict/data/` — Gofonix runtime dictionary data, when bundled
- **Modifications:** None yet. If CMUdict data is filtered, normalized,
  re-encoded, subsetted, or otherwise transformed, the change must be recorded
  in `THIRD_PARTY_NOTICES.md`, `third_party/cmudict/README.md`, and any
  relevant reproducibility notes.
- **Redistribution note:** The BSD 2-Clause style terms require that the
  copyright notice, the list of conditions, and the disclaimer be retained in
  both source and binary redistributions. Gofonix satisfies this by keeping
  the bundled CMUdict license text and the entry in `THIRD_PARTY_NOTICES.md`.
- **Binary distribution note:** If CMUdict-derived data is embedded into a
  Go binary, including through `//go:embed`, the CMUdict copyright notice,
  conditions, and disclaimer must accompany that binary distribution through
  documentation or other materials.

---

## 2. enwik8

- **Upstream URL:** https://mattmahoney.net/dc/enwik8.zip
- **Description:** The first 100,000,000 bytes of an English-language
  Wikipedia XML dump, distributed as part of the Large Text Compression
  Benchmark / Hutter Prize materials.
- **Status in repo:** **NOT committed.** This dataset is downloaded on demand
  into `testdata/downloads/` and is excluded via `.gitignore`. It must never
  be added to version control.
- **Download method:** Use `scripts/download_testdata.sh` and verify the
  uncompressed file with `scripts/verify_checksums.sh`.
- **Checksums for uncompressed `enwik8`:**
  - Size: `100000000` bytes
  - MD5: `a1fa5ffddb56f4953e226637dabbb36a`
  - SHA-1: `57b8363b814821dc9d47aa4d41f58733519076b2`
- **Licensing:** This is Wikipedia-derived content under Wikimedia licensing
  terms, including CC BY-SA and GFDL. The exact applicable license version
  should be verified from Wikimedia terms at the time of use.
- **Redistribution status:** Do not redistribute from this repository. If any
  extract, derived sample, benchmark package, or transformed version is ever
  shared, review the applicable attribution, share-alike, and documentation
  obligations first.
- **Research note:** Results that use Gofonix or other phonological resources
  as external information should be clearly labeled as such and not presented
  as directly comparable to benchmark submissions that forbid outside
  information.

---

## 3. enwik9

- **Upstream URL:** https://mattmahoney.net/dc/enwik9.zip
- **Description:** The first 1,000,000,000 bytes of an English-language
  Wikipedia XML dump, distributed as part of the Large Text Compression
  Benchmark / Hutter Prize materials.
- **Status in repo:** **NOT committed.** This dataset is downloaded on demand
  into `testdata/downloads/` and is excluded via `.gitignore`. It must never
  be added to version control.
- **Download method:** Use `scripts/download_testdata.sh` and verify the
  uncompressed file with `scripts/verify_checksums.sh`.
- **Checksums for uncompressed `enwik9`:**
  - Size: `1000000000` bytes
  - MD5: `e206c3450ac99950df65bf70ef61a12d`
  - SHA-1: `2996e86fb978f93cca8f566cc56998923e7fe581`
- **Licensing:** This is Wikipedia-derived content under Wikimedia licensing
  terms, including CC BY-SA and GFDL. The exact applicable license version
  should be verified from Wikimedia terms at the time of use.
- **Redistribution status:** Do not redistribute from this repository. If any
  extract, derived sample, benchmark package, or transformed version is ever
  shared, review the applicable attribution, share-alike, and documentation
  obligations first.
- **Research note:** Results that use Gofonix or other phonological resources
  as external information should be clearly labeled as such and not presented
  as directly comparable to benchmark submissions that forbid outside
  information.

---

## 4. Synthetic test fixtures

- **Description:** Small, hand-authored or programmatically generated inputs
  used in unit and integration tests, such as toy word lists, contrived
  phoneme sequences, and edge-case strings.
- **Origin:** Created originally for this project. They must not be copied
  from CMUdict, enwik8, enwik9, Wikipedia, books, websites, logs, user data,
  proprietary corpora, or any other third-party source unless that source is
  explicitly documented and license-reviewed.
- **Licensing:** Synthetic fixtures that are original to Gofonix are covered
  by the project's Apache-2.0 license.
- **Status in repo:** May be committed when they are small, deterministic, and
  clearly synthetic.
- **Storage:** Store synthetic fixtures under the relevant package or project
  `testdata/` directory. Do not store them under `testdata/downloads/`, which
  is reserved for downloaded external benchmark data.
- **Documentation checklist for each committed fixture:**
  - Purpose of the fixture
  - How it was generated, including script path or short description
  - Date created
  - Author or generator
  - Confirmation that it contains no copied third-party source material
- **Derived fixtures:** If a fixture is ever derived from a third-party source,
  it must be moved out of the synthetic category and documented with its own
  provenance, license, attribution, transformation steps, and checksums.

---

## 5. Generated datasets and transformations

Any generated dataset, transformed dictionary, feature stream, benchmark
fixture, or derived artifact must be documented before it is committed or
shared.

For every generated or transformed dataset, record:

- Input dataset name and upstream URL
- Upstream commit, release, version, or download date
- Input file size and checksum
- Transformation command or script path
- Transformation parameters
- Output file names
- Output file sizes and checksums
- Applicable license or redistribution terms
- Gofonix version or commit used to generate the output
- Phonological feature schema version, when relevant

Prefer committing the reproducible transformation script rather than committing
large transformed data. Do not commit generated datasets if their source
license is unclear, incompatible with redistribution, or subject to share-alike
terms that have not been reviewed.

---

## Summary

| Dataset or data class | Committed? | License / terms |
|-----------------------|:----------:|-----------------|
| CMUdict license metadata | Yes | BSD 2-Clause style license from CMU |
| CMUdict-derived dictionary data | When explicitly bundled | BSD 2-Clause style license from CMU |
| enwik8 | No | Wikimedia-derived terms, including CC BY-SA / GFDL; verify version |
| enwik9 | No | Wikimedia-derived terms, including CC BY-SA / GFDL; verify version |
| Synthetic fixtures | Yes, if original and small | Apache-2.0, when original to Gofonix |
| Generated datasets | Case by case | Depends on input data, transformation, and redistribution status |