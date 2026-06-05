# Data Policy

This policy governs how data is handled in the Gofonix repository. It
complements `DATA_LICENSES.md` (which catalogs licenses) and
`THIRD_PARTY_NOTICES.md` (which records attributions).

The Gofonix **code** is Apache-2.0. Third-party **data** keeps its own
license. Treat all data licensing conservatively.

---

## 1. What may be committed

- Original, synthetic test fixtures created for this project (see below).
- Third-party data that is **small**, **clearly licensed for
  redistribution**, and accompanied by its license and attribution
  (e.g., CMUdict under `third_party/cmudict/`).
- Documentation, checksums, and manifests describing external datasets.
- Gofonix source code and documentation under Apache-2.0.

## 2. What must NOT be committed

- Large benchmark corpora — specifically **enwik8** and **enwik9** (and their
  `.zip` archives). These are downloaded on demand only.
- Any data whose license is unknown, ambiguous, or incompatible with
  redistribution until its status is verified and documented.
- Anything under `testdata/downloads/` or `benchmarks/results/` (gitignored).
- Generated artifacts, temporary files (`*.tmp`), or archives (`*.zip`).
- Large third-party corpora without a repository-specific approval and
  license review.

## 3. Handling CMUdict

- CMUdict is bundled under `third_party/cmudict/` and used from
  `internal/lang/en/data/`.
- Keep the verbatim license at `third_party/cmudict/LICENSE` and the entry in
  `THIRD_PARTY_NOTICES.md`. Do not remove the CMU copyright notice or
  disclaimer in any redistribution (source or binary).
- Record any change to the bundled dictionary in the **Modifications** field
  of `THIRD_PARTY_NOTICES.md` and in `third_party/cmudict/README.md`. If no
  changes have been made, the field reads `None yet`.
- When updating, document the upstream commit/release and re-verify integrity.
- If CMUdict is modified, update all relevant documentation to describe the
  change, including the source version, date, files changed, transformation
  steps, and whether the changes are reversible.

## 4. Handling enwik8 / enwik9

- **Never commit** these datasets or their archives.
- Download them only via `scripts/download_testdata.sh`, which places them in
  `testdata/downloads/` (gitignored).
- After download, always verify size and SHA-1 with
  `scripts/verify_checksums.sh` (MD5 optional). The expected values are
  recorded in `DATA_LICENSES.md` and in the verification script.
- Their licensing is Wikimedia-derived (CC BY-SA / GFDL); verify the exact
  applicable version from Wikimedia terms before any redistribution. If
  results, extracts, or derived samples are shared, review whether
  attribution, share-alike, or other obligations apply.

## 5. Creating synthetic fixtures

- Keep fixtures small, deterministic, and obviously synthetic.
- Author them originally for this project; do **not** copy from CMUdict,
  enwik8/enwik9, or any other third-party corpus.
- Do not copy passages from Wikipedia, books, websites, logs, user data, or
  proprietary corpora into synthetic fixtures.
- Store them under the relevant package's `testdata/` directory (not under
  `testdata/downloads/`).
- If a fixture is ever derived from a third-party source, document its
  provenance and license, and add it to `DATA_LICENSES.md`.
- Each synthetic fixture should include or be accompanied by documentation
  stating: purpose, how it was generated, date created, author or generator,
  and confirmation that it contains no copied third-party source material.

## 6. Documenting data transformations

- Any transformation applied to third-party data (filtering, normalization,
  re-encoding, subsetting) must be documented:
  - in `THIRD_PARTY_NOTICES.md` (Modifications field), and/or
  - in the dataset's README or in `docs/reproducibility.md`.
- Prefer **reproducible** transformations: commit the script that performs the
  transformation rather than the transformed third-party output, unless the
  output is itself small and redistributable.
- Record the input checksum, the transformation step, and the output checksum
  so the result can be regenerated and verified.
- For every transformed dataset, document: input dataset name and upstream
  URL, download date, upstream commit/release/version, input checksums and
  file sizes, transformation command or script path, transformation
  parameters, output file names/sizes/checksums, Gofonix version or commit
  used, and feature-schema version when phonological features are generated.
- Prefer scripted transformations over manual edits. When manual edits are
  unavoidable, document the exact files and rationale.

## 7. Documenting checksums

- For every externally downloaded dataset, record:
  - uncompressed **size** in bytes,
  - **SHA-1** (required), and
  - **MD5** (optional, for cross-checking; do not rely on MD5 alone for
    integrity).
- Keep these values in `DATA_LICENSES.md` and mirrored in
  `scripts/verify_checksums.sh`.
- Re-verify checksums after every download and whenever the upstream source
  is updated. A checksum mismatch must be treated as a failure — do not use
  the data until the discrepancy is resolved.

For `enwik8` and `enwik9`, the expected values for uncompressed files are:

| Dataset | Size (bytes) | SHA-1 | MD5 |
|---------|-------------:|-------|-----|
| enwik8  | 100000000    | `57b8363b814821dc9d47aa4d41f58733519076b2` | `a1fa5ffddb56f4953e226637dabbb36a` |
| enwik9  | 1000000000   | `2996e86fb978f93cca8f566cc56998923e7fe581` | `e206c3450ac99950df65bf70ef61a12d` |
