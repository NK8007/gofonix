# Reproducibility Template

Use this template to document each experiment so that results can be
reproduced exactly. Copy this file into your experiment log (e.g.,
`experiments/<id>/reproducibility.md`) and fill in every field. Leave no
field blank — write `N/A` if genuinely not applicable.

---

## Experiment metadata

- **Experiment ID:** `<short-id>`
- **Title:** `<descriptive title>`
- **Author:** `<name>`
- **Date (UTC):** `<YYYY-MM-DD>`
- **Objective / hypothesis:** `<what is being tested>`

## Software environment

- **Gofonix version:** `<git tag or commit hash>`
- **Branch:** `<branch name>`
- **Build command:** `<e.g., go build ./...>`
- **Phonological feature schema version:** `<schema version / hash>`
- **Schema file path:** `<path>`
- **Schema checksum:** `<sha1>`
- **Notes on schema changes since previous run:** `<description or N/A>`
- **Go version:** `<go version output>`
- **OS / architecture:** `<e.g., linux/amd64>`
- **Other dependencies:** `<libraries, tools, versions>`
- **Random seed(s):** `<seed values>`

## Data versions

For **every** dataset used, fill in one block. This is the core of
reproducibility — be precise.

### Dataset: `<name, e.g., CMUdict / enwik8 / enwik9 / synthetic-fixture>`

- **Upstream URL:** `<source URL>`
- **Date downloaded (UTC):** `<YYYY-MM-DD>`
- **Upstream commit, release, or version:** `<hash or tag>`
- **Local file path:** `<path>`
- **Checksum:**
  - Size (bytes): `<size>`
  - SHA-1: `<sha1>`
  - MD5 (optional): `<md5>`
- **License or data terms:** `<license identifier or description>`
- **Redistribution status:** `<may redistribute / do not redistribute / verify before redistributing>`
- **Transformations applied:** `<none | description + script path>`
  - Input checksum → output checksum: `<...> → <...>`
- **Gofonix version used to process this data:** `<git tag or commit>`
- **Phonological feature schema version:** `<schema version / hash>`
- **Notes and caveats:** `<sampling, subsetting, encoding, etc.>`

> Repeat the **Dataset** block above for each dataset.

---

## Required entries for common datasets

The sections below are pre-filled with known values. Complete the blank
fields when running an experiment.

### CMUdict

- **Upstream URL:** https://github.com/cmusphinx/cmudict
- **Date downloaded:** `<YYYY-MM-DD>`
- **Upstream commit, release, or version:** `<hash>`
- **Local file path:** `third_party/cmudict/` or `internal/lang/en/data/`
- **SHA-1 checksum:** `<sha1 of bundled file>`
- **License:** BSD 2-Clause variant (see `third_party/cmudict/LICENSE`)
- **Redistribution status:** May redistribute with copyright notice and
  disclaimer preserved.
- **Transformations applied:** `<none | description>`

### enwik8

- **Upstream URL:** https://mattmahoney.net/dc/enwik8.zip
- **Date downloaded:** `<YYYY-MM-DD>`
- **Upstream commit, release, or version:** N/A unless the upstream page
  provides one
- **Expected uncompressed size:** `100000000` bytes
- **Expected uncompressed SHA-1:** `57b8363b814821dc9d47aa4d41f58733519076b2`
- **Expected uncompressed MD5:** `a1fa5ffddb56f4953e226637dabbb36a`
- **License or data terms:** Wikipedia-derived content under Wikimedia
  licensing terms, including CC BY-SA and GFDL; exact applicable version
  should be verified from Wikimedia terms at the time of use.
- **Redistribution status:** Verify Wikimedia terms before redistributing.
- **Transformations applied:** `<none | description>`

### enwik9

- **Upstream URL:** https://mattmahoney.net/dc/enwik9.zip
- **Date downloaded:** `<YYYY-MM-DD>`
- **Upstream commit, release, or version:** N/A unless the upstream page
  provides one
- **Expected uncompressed size:** `1000000000` bytes
- **Expected uncompressed SHA-1:** `2996e86fb978f93cca8f566cc56998923e7fe581`
- **Expected uncompressed MD5:** `e206c3450ac99950df65bf70ef61a12d`
- **License or data terms:** Wikipedia-derived content under Wikimedia
  licensing terms, including CC BY-SA and GFDL; exact applicable version
  should be verified from Wikimedia terms at the time of use.
- **Redistribution status:** Verify Wikimedia terms before redistributing.
- **Transformations applied:** `<none | description>`

---

#### Reference checksums (known datasets)

| Dataset | Size (bytes)   | SHA-1                                      | MD5                                |
|---------|---------------:|--------------------------------------------|------------------------------------|
| enwik8  | 100000000      | `57b8363b814821dc9d47aa4d41f58733519076b2` | `a1fa5ffddb56f4953e226637dabbb36a` |
| enwik9  | 1000000000     | `2996e86fb978f93cca8f566cc56998923e7fe581` | `e206c3450ac99950df65bf70ef61a12d` |

---

## Run configuration

- **Command:** `<exact command with all flags>`
- **Configuration file(s):** `<path + commit hash>`
- **Hyperparameters / settings:** `<list>`
- **Random seed:** `<value>`
- **CPU/GPU details:** `<processor, core count>`
- **Memory limit:** `<value or N/A>`
- **Number of workers:** `<value>`
- **Important environment variables:** `<KEY=value list>`

## Results

- **Output directory:** `<path; note that benchmarks/results/ is gitignored>`
- **Logs:** `<path>`
- **Metrics file:** `<path>`
- **Model or artifact files:** `<paths>`
- **Output checksums (if applicable):** `<sha1 or md5>`
- **Key metrics:** `<key numbers>`

## Verification

- [ ] Data checksums verified with `scripts/verify_checksums.sh`
- [ ] Gofonix version and feature-schema version recorded
- [ ] Commands reproducible from a clean checkout
- [ ] Random seeds fixed and recorded
- [ ] Known limitations documented
- [ ] Deviations from planned protocol documented
- [ ] Data quality caveats recorded
- [ ] License or redistribution caveats noted
