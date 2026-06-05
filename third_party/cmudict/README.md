# CMU Pronouncing Dictionary (CMUdict)

## Source

CMUdict comes from the CMU Sphinx project:
**https://github.com/cmusphinx/cmudict**

It is the CMU Pronouncing Dictionary, a machine-readable pronunciation
dictionary for North American English.

## License

BSD 2-Clause style license.
**Copyright (C) 1993-2015 Carnegie Mellon University. All rights reserved.**

The redistribution conditions (retaining the copyright notice, the list of
conditions, and the disclaimer) apply to both source and binary forms.

> Note: CMUdict includes the non-standard clause "The contents of this file
> are deemed to be source code." This is informational, not a license
> restriction — it means the dictionary text is treated as source code for the
> purposes of condition 1 (retain notice in source redistributions).

The CMUdict README also requests (as a courtesy, not a license condition):
*"we request that you acknowledge its origin in your descriptions."*
This is recommended but not legally required.

## Where to find the full license text

- Verbatim copy in this directory: [`LICENSE`](./LICENSE)
- Repository-level third-party attributions:
  [`../../THIRD_PARTY_NOTICES.md`](../../THIRD_PARTY_NOTICES.md)

## Modifications

**None yet.** The bundled data is used as-is from upstream. If the dictionary
is ever modified (filtering, normalization, re-encoding, subsetting), record
the change here and in `THIRD_PARTY_NOTICES.md` (the *Modifications* field),
and document the transformation per `docs/data_policy.md`.

## Bundled paths

This dictionary is used from:

- `third_party/cmudict/` — verbatim upstream copy and license
- `internal/lang/en/data/` — copy used by the application at runtime

## How to update the version

1. Identify the target upstream commit or release at
   https://github.com/cmusphinx/cmudict.
2. Replace the bundled data files with the new upstream version.
3. Refresh `LICENSE` here **only if** the upstream license text changed
   (copy it verbatim — no edits).
4. Update the upstream commit/release reference and the *Modifications* field
   in `THIRD_PARTY_NOTICES.md`.
5. If any transformation is applied, document it per `docs/data_policy.md`
   and record input/output checksums.
6. Run the test suite to confirm the new version integrates correctly.
