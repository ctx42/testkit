## v0.11.0 (Fri, 17 Jul 2026 12:59:27 UTC)
- style: fix inert //nolint directive spacing.
- chore: add SPDX headers to httpkit and prjkit.
- fix(prjkit): stop git tag regex capturing trailing refs.
- fix(prjkit): include command error on compile failure.
- docs(prjkit): fix godoc grammar and add TempDir helper mark.
- docs(pathkit): fix awkward filepath.Join godoc phrasing.
- test(pathkit): fix error subtest name and use must.Value.
- docs(jsonkit): host package godoc in package-named file and fix docs.
- test(jsonkit): use ExpectError and prepare nil arg in Given.
- fix(modkit): close go.mod handles and improve error reporting.
- test(modkit): use have, error- prefixes, fix typo and order.
- refactor(timekit): drop no-op lock and fix package godoc.
- test(timekit): apply have naming and group fixtures.
- fix(reflectkit): guard GetField/GetValue against nil inputs.
- test(reflectkit): add nil/error coverage and tighten assertions.
- refactor(randkit): correct misleading docs and tidy declarations.
- test(randkit): add Given/When/Then markers and mirror source order.
- refactor(selfkit): unify *Self locals on slf and fix godoc.
- test(selfkit): add Given block, explicit discards, uniform naming.
- docs(subkit): reference New in README and note raw-env read.
- test(subkit): unexport const helpers and fix names and structure.
- fix(testkit): honor wait throttle override and wrap read error.
- test(testkit): align with style rules.
- refactor(iokit): assert interface impls and trim method godocs.
- test(iokit): use have naming, oskit helper, fix cleanup and receiver.
- fix(netkit): correct connection detection and error handling.
- test(netkit): fix vacuous dup-port check and add missing tests.
- test(prjkit): align project tests with style conventions.
- refactor(exekit)!: fix exit-code check and split coverage API.
- test(exekit): reorganize tests, rename method tests, add ExampleNew.
- fix(oskit): stop Create truncating after a failed write.
- fix(dkrkit): clone env before append and simplify NetRm.
- refactor(dkrkit)!: reorder HasBuildArg params to key, want.
- test(dkrkit): use assert lib, fix test names and FindByID fixtures.
- fix(httpkit): guard Server against data races and panics.
- refactor(httpkit): tidy RespWriter godoc, receiver, and assertion.
- test(httpkit): cover Server error and edge branches.
- test(httpkit): add Given markers and error- subtest prefixes.

## v0.10.0 (Tue, 14 Jul 2026 14:03:51 UTC)
- refactor!: adopt xdef v0.7.0 names and align container metadata.

## v0.9.0 (Mon, 13 Jul 2026 20:15:53 UTC)
- refactor(prjkit)!: adopt xdef env constants and OCI image labels.

## v0.8.1 (Mon, 13 Jul 2026 19:52:08 UTC)
- build(deps): bump ctx42/xdef to v0.5.0.

## v0.8.0 (Fri, 10 Jul 2026 08:13:16 UTC)
- feat(jsonkit): add JSON marshalling test helpers.

## v0.7.0 (Fri, 26 Jun 2026 11:12:56 UTC)
- feat(iokit): add ReadAll and ReadAllStr helpers.
- docs(httpkit): add README and usage examples to package documentation.
- style: fix golangci-lint findings across packages.

## v0.6.0 (Thu, 25 Jun 2026 07:58:29 UTC)
- refactor(oskit)!: merge Create/Write pairs into generic functions.

## v0.5.0 (Tue, 23 Jun 2026 21:47:00 UTC)
- feat(prjkit): add Project helper for temporary Go test projects.
- docs(README): simplify main README, trim niche packages.
- fix(prjkit): set local git identity before every git commit.
- test(prjkit): set git identity in tests that call git commit directly.

## v0.4.0 (Tue, 16 Jun 2026 20:09:15 UTC)
- feat: improve documentation, add subkit, WithSeed for randkit, and runnable examples.
- feat(oskit): add filesystem and environment test helpers.
- docs: standardize examples and sync READMEs across packages.
- feat(pathkit): add path resolution test helpers.
- docs: document pathkit and list it in the module README.
- feat(modkit): add Go module location and go.mod inspection helpers.
- docs: document modkit and list it in the module README.
- feat(netkit): add networking helpers for tests.
- docs: document netkit and list it in the module README.
- ci: add GitHub Actions workflow for Go tests.
- docs: simplify README glance table by merging package and import path columns.
- doc: modkit, testkit: remove outdated TODOs and improve Wait4File docs.
- chore(idea): move run config from build/ to dev/.
- refactor(netkit): use range-over-int loop in GetFreePorts.
- refactor(oskit): remove redundant loop-variable captures in tests.
- refactor(randkit): replace counted for-loops with range-over-int.
- docs(selfkit): update README examples to use NewT.
- docs(subkit): update README examples to use NewT; fix NewPkg godoc.
- fix(iokit): correct function name in README examples.
- fix(modkit): update version assertion and testdata module paths.
- test(testkit): add tests for Wait4File.
- feat(exekit): add WithLax option.
- build(deps): bump testing to v0.54.0, add xdef v0.1.0.
- feat(dkrkit): add Docker CLI wrapper for integration tests.
- ci: pin Docker to v29.1.3 in GitHub Actions.

## v0.3.1 (Mon, 08 Jun 2026 06:52:06 UTC)
- docs: improve package code documentation.

## v0.3.0 (Sun, 07 Jun 2026 17:47:18 UTC)
- chore: update module imports and bump ctx42/testing to v0.51.0.

## v0.2.0 (Sun, 07 Jun 2026 17:20:57 UTC)
- docs: update documentation.

## v0.1.0 (Sun, 07 Jun 2026 17:10:39 UTC)
- feat: establish testkit module with seven sub-packages.

