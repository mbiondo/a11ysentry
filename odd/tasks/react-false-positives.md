# React false positives reported by UPCEFE

Objective: remove the five false-positive classes an external consumer (UPCEFE, Next.js 16 /
React 19 backoffice) reported against v0.2.0, plus the MCP configuration gap found while tracing them.
Branch: `fix/react-false-positives` · Delivery strategy: `single-pr` · Test runner: `go -C <module> test ./...`
(there is no root module).

## Specs

Verbatim report (thread from the UPCEFE orchestrator agent, 2026-10-14):

- **S1 — G141 must not analyze test files.** "5. G141 (50 warnings): Analiza archivos test.tsx y
  reclama falta de h1; el scanner deberia ignorar archivos de test por defecto."
- **S2 — G141 must detect an H1 that carries an id.** Derived from S1 while tracing it: the reporter
  attributed all 50 G141 warnings to test files, but the dominant cause is that
  `<h1 id="main-title">` is invisible to the rule, so pages that do have an H1 are reported as
  missing one.
- **S3 — WCAG_4_1_3 must not demand redundant `aria-live`.** "3. WCAG_4_1_3 (37 warnings): Exige
  aria-live explicito en elementos con role alert o role status, lo cual ya es implicito en
  WAI-ARIA 1.2+."
- **S4 — ARIA_1_1 must not treat every header/form as a landmark.** "4. ARIA_1_1 (28 warnings):
  Marca headers o forms de subcomponentes como banners sin nombre; en HTML5 solo el header de raiz
  actua como banner."
- **S5 — WCAG_2_4_3_HIDDEN must not leak `aria-hidden` across component boundaries.** "2.
  WCAG_2_4_3_HIDDEN (51 casos): Confunde spans dentro de labels (como asteriscos de requerido) con
  contenedores del input, reportando falsamente que el input esta dentro de aria-hidden." (The
  reported mechanism is not the real one; the symptom is. See L4.)
- **S6 — WCAG_4_1_1 must not report an ID as duplicated when the same component is ingested twice
  into one tree.** "1. WCAG_4_1_1 (Duplicate ID, 82 casos): Compara IDs globalmente entre componentes
  .tsx independientes que nunca coexisten en el DOM." (The reported mechanism is not the real one;
  the symptom is. See L5.)
- **S7 — The MCP path must honor the project's `exclude` configuration.** Derived while tracing S1:
  the CLI passes configured excludes to project discovery, the MCP server does not, so
  `a11ysentry.json` `exclude` silently has no effect through MCP.
- **S8 — Every output format must report a shared component once, not once per page.** Derived from
  L19 while re-measuring: `--format json` emitted 58 entries for 40 distinct problems, and
  `--format sarif` would post 58 Code Scanning alerts for the same 40, because both serialized the
  raw per-tree reports. Text mode already collapsed them. Fixing it also changes the JSON body,
  which becomes the cross-tree unique list and carries `pageCount`.

Scope decision (user, 2026-10-14): phases 1 and 2 — fix the five reported symptoms plus S7. Making
the rules fully ancestor-aware via a populated USN hierarchy (phase 3) is explicitly out of scope.

## Tasks

| ID | Specs | Route | Status | Evidence |
|----|-------|-------|--------|----------|
| T1 | — | inline | done | branch `fix/react-false-positives`; this document |
| T2 | S2 | inline | done | `Traits["tag"]` recorded by the adapter; `HasH1` reads it |
| T3 | S1 | inline | done | `scanner.IsTestFile` + filter in 12 collectors |
| T4 | S3 | inline | done | no warning for roles that imply `aria-live`; invalid values still reported |
| T5 | S7 | inline | done | MCP loads `a11ysentry.json` before project discovery |
| T6 | S5 | inline | done | `aria-hidden` no longer crosses inter-file boundaries |
| T7 | S6 | inline | done | one ingestion per file per analysis |
| T8 | S4 | inline | done | nameless `form`/`section` and nested `header`/`footer` are not landmarks |
| T9 | S1-S7 | delegated | done | verifier L13: 6/7 specs; 1 blocker (S7) and 2 regressions, all corrected in L14 and covered by tests |
| T10 | S8 | inline | done | `printReports` deduplicates cross-tree for text, json and sarif; RED first (`json: got 3 entries for 1 distinct problem`) |
| T11 | S8 | delegated | done | verifier over `b10098a`: no blockers, S8 met |

## Log

- **L1** (external report, 2026-10-14, verbatim): "Buenas. Te contacto desde el agente orquestador
  del repo UPCEFE. Probamos el MCP de a11ysentry v0.2.0 sobre nuestro frontend Next.js 16 y React 19
  (backoffice) y detectamos 5 tipos de falsos positivos en el analisis estatico para que las
  investigues: [...] Podes revisar el AST engine y las reglas para ajustar estas heuristicas en
  componentes React?"
- **L2** (user decision, 2026-10-14): scope = phases 1 and 2, covering the five reported symptoms
  plus S7; the USN hierarchy work (phase 3) is deferred.
- **L3** (explorer findings, 2026-10-14): of the five reports, S3 and S4 are correct as diagnosed.
  S1 is correct but incomplete. S5 and S6 have the correct symptom with the wrong mechanism.
- **L4** (explorer findings, 2026-10-14): the `aria-hidden` leak is not a sibling problem. Within a
  file, `traverse` scopes correctly (`adapters/web/adapter.go:443-459`). The leak is at FileNode
  granularity: `adapter.go:139-150` scans the file's entire flat node list and sets `childHidden` if
  any node anywhere is `aria-hidden="true"`, then `adapter.go:156` passes that flag to every imported
  child component.
- **L5** (explorer findings, 2026-10-14): duplicate IDs are not compared across pages. Each PageTree
  gets its own `Analyze` call. The cause is `scanner/nextjs/nextjs.go:143` and `:156` calling
  `CollectTree` with a fresh `make(map[string]bool)` per chain element, so a component imported by
  both a layout and its page is ingested twice within one tree.
- **L6** (explorer findings, 2026-10-14): `HasH1` is computed by `fmt.Sscanf(node.UID, "h%d")`
  (`engine/core/domain/metadata.go:61-67`), but the adapter sets `UID` from the id attribute first
  and only falls back to the tag name (`adapters/web/adapter.go:375,400-402`). Any `<h1 id="...">` is
  therefore missed.
- **L7** (explorer findings, 2026-10-14): no test/spec filter exists anywhere in `scanner/`.
  `a11ysentry.json` `exclude` is applied only by the CLI at `cmd/a11ysentry/main.go:218`;
  `mcp/server/server.go:127` calls `FindProjectRoots` with no excludes.
- **L8** (explorer findings, 2026-10-14): `mapRole` (`adapters/web/adapter.go:651-686`) is entirely
  ancestor-blind, and `isLandmark` (`engine/core/domain/utils.go:61-66`) includes `RoleForm`, so any
  nameless `<form>` counts as a landmark and feeds ARIA_1_1.
- **L9** (implementation, 2026-10-14): fixes applied. The adapter now records `Traits["tag"]` and
  `metadata.go` reads it for the heading level; `scanner.IsTestFile` filters test/spec/mock/snapshot
  files in all 12 collectors; `rules_wcag_4_1_3.go` no longer demands an explicit `aria-live` for
  roles that imply it (only a malformed value is reported); the MCP path loads `a11ysentry.json`
  before discovery; `ingestRecursive` no longer propagates `aria-hidden` across files and no longer
  re-ingests a file reached through several parents; and the adapter demotes nested `header`/`footer`
  and nameless `section`/`form` to `RoleGeneric`.
- **L10** (RED/GREEN evidence, 2026-10-14): stashed only the source changes and ran the new tests
  against the old code. All six adapter tests failed (`G141` on a page with `<h1 id=...>`, a
  `WCAG_4_1_3` warning for `role="alert"`, a nested header mapped to banner, a nameless form mapped to
  the form landmark, `aria-hidden` leaking into an imported component, and a shared component's field
  appearing twice). After restoring the fixes they pass. The scanner test also failed first — it
  caught a real over-broad regex where `latest.ts` and `Contest.tsx` matched the substring `test.`,
  which is why the patterns are now anchored.
- **L11** (own checks, 2026-10-14): `go build ./cmd/a11ysentry` exit 0; `go vet` clean on
  `engine`, `scanner`, `adapters/web`, `mcp`, `cmd/a11ysentry`; tests pass in `engine`, `scanner`,
  `adapters/web`, `adapters/angular`, `adapters/vue`, `adapters/django`, `mcp`.
- **L12** (known limitation, 2026-10-14): a `<header>` inside an imported component that is not
  wrapped in sectioning content is still mapped to banner, because confirming it would need the
  populated USN hierarchy that phase 3 defers. The deterministic cases (nested in
  `article`/`aside`/`main`/`nav`/`section` within the same file) are fixed.
- **Next step**: independent verification of the whole feature against the branch point.
- **L13** (independent verification, 2026-10-14): verifier returned **6 of 7 specs passing**. The five
  reported false positives are genuinely removed and the true-positive output for img-alt,
  heading-skip, contrast, `tabindex=-1` and two `<main>` is byte-identical to the baseline. It found
  one blocker (S7: `FindProjectRoots` short-circuits at a project root, so excludes never applied) and
  two regressions this change introduced (an `aria-labelledby`/`title`-named section or form was
  demoted; `aria-labelledby` was not recorded in Traits so the landmark rule still demanded a label).
  It also reported that `resolve.go` resolves `.tsx`/`.jsx` imports twice, duplicating every graph
  edge — a pre-existing cause of double ingestion that the ingestion guard now masks.
- **L14** (correction batch, 2026-10-14): added `scanner.FilterExcluded` and applied it in the MCP
  server and both CLI paths, closing S7; made `aria-labelledby`/`title` count as an accessible name;
  recorded `aria-labelledby` in Traits and shared one `nodeHasAccessibleName` notion between adapter
  and rules; and moved both heading checks onto a single `headingLevel` helper.
- **L15** (real-world measurement, 2026-10-14): ran the released v0.2.0 binary and then this branch
  against `C:\Users\Maxim\Documents\repositories\UPCEFE\backoffice` (Next.js 16 / React 19, no
  `a11ysentry.json`, so default excludes). Total violations **524 -> 42 (-92%)**. By rule:
  `WCAG_4_1_1` 151->2, `WCAG_2_4_3_HIDDEN` 107->2, `WCAG_4_1_3` 94->1, `ARIA_1_1` 55->0, `G141` 54->4.
  Rules that were not touched also dropped (`ADV_FOCUS_TRAP` 25->13, `MODAL_MISSING_LABEL` 25->13,
  `REDUNDANT_TITLE` 4->2, `WCAG_2_4_1` 3->1), consistent with the ingestion guard removing duplicated
  component instances.
- **Status**: implementation and correction complete; delivery (push, PR, merge) is the user's decision.
- **L16** (dead code, 2026-10-14): deleted `refactor.go`, a 394-line orphan that declared `package
  main` at the repository root where there is no Go module, so nothing ever compiled or imported it.
  It held a stale copy of the web adapter including a second `aria-hidden-inherited` setter.
- **L17** (root cause, 2026-10-14): `scanner/resolve.go` probed `.tsx`/`.jsx` twice — they are listed
  in both `UIExtensions` and the follow-up extension list — and `add()` appended without deduplicating,
  so every component edge in the graph was recorded twice. Fixed by correcting the probe and guarding
  `add()`; verified RED first (`Widget.tsx resolved 2 times`). This is the cause the ingestion guard
  had only masked.
- **L18** (major finding, 2026-10-14): the Next.js `Probe` only checked a root `app/` directory, so a
  project using `src/app/` failed detection and fell through to the generic scanner. On the reporting
  backoffice that produced 69 page trees instead of 13 and made every orphan component and layout its
  own "page" — which is why `G141` fired on `header.tsx`, `breadcrumb.tsx`, `alert-dialog.tsx` and
  `layout.tsx`. After the fix the deduplicated report went from **287 violations to 40 (-86%)** and
  `G141` reached zero.
- **L19** (measurement correction, 2026-10-14): the earlier JSON-based numbers were inflated. `printReports`
  applies `DeduplicateCrossTree` only in text mode (`main.go:573`); `--format json` serializes the raw
  per-tree reports, so violations from a layout shared by 13 pages appear once per page. Text-mode
  totals are the trustworthy ones.
- **L20** (correction, 2026-10-14): L19 also claimed the MCP path fails to collapse shared-component
  duplicates. That was wrong. `DeduplicateViolations` keys on `ErrorCode|FilePath|Line|Column`, the
  same key as `DeduplicateCrossTree`, and `analyzeDirectory`/`analyzeFiles` hand it the accumulated
  list, so the MCP output was already cross-tree unique. Measured against the same project: the MCP
  never produced the 58-entry inflation.
- **L21** (fix, 2026-10-14): `printReports` now computes the cross-tree unique list once, before the
  format switch, and renders it for text, json and sarif. SARIF derives its rules and results from
  that list (`FromReports` only reads `Violations`), and JSON serializes `[]UniqueViolation`, which
  keeps the information the old per-page repetition carried, as `pageCount`. RED observed first
  (`json: got 3 entries for 1 distinct problem`), GREEN after. On the reporting backoffice all three
  formats now agree: 40 entries, 24 errors / 16 warnings, `pageCount` summing to the previous 58.
- **L22** (independent verification of S8, 2026-10-14): a verifier over `b10098a` returned **no
  blockers**. It reproduced the RED exactly, confirmed text/json/sarif agree on the distinct count,
  that `pageCount` sums to the old per-tree total, that SARIF keeps its rules and results consistent
  with the unique list, that no in-repo consumer parses the old JSON shape, and that the MCP claim in
  the commit message is correct. Its regression probes were committed as
  `cmd/a11ysentry/s8_formats_test.go` and `mcp/server/s8_crosstree_test.go`. Advisories, none a
  defect: `PageCount` counts occurrences rather than trees, so its comment is loose; empty SARIF emits
  `results: null` (baseline behaviour); the web adapter reports only the first missing-alt image per
  file (baseline behaviour); and the S8 JSON body is a breaking change for external consumers.
- **L23** (open item, 2026-10-14): the scoped recheck of the L14 corrections was covered by the test
  suite rather than by a fresh verifier run, so T9 is recorded as done on test evidence, not on a
  second independent pass.
- **Status**: implementation, correction and measurement complete; 13 commits on the branch; S8
  verified with no blockers.
