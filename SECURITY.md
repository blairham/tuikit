# Security Policy

## What tuikit does with your text

tuikit is a library: it draws what your app hands it and runs inside your
app's process, with your app's privileges. It adds none of its own.

- **It does not sanitize content.** Table cells, log lines in a `tail`, view
  titles and modal text are drawn as given, escape sequences included — that
  is how styled content keeps its color. Text from an untrusted source, such
  as a container's log output, should have its terminal control sequences
  stripped by the app before it reaches tuikit.
- `table.FixSelectedRow`, `table.FixRows` and `theme.ReassertBackground`
  rewrite SGR (color) sequences and only those; every other byte passes
  through unchanged.
- The filter bar's text (`table.ParseFilter`, `tail.Model.SetFilter`) is
  compiled as a case-insensitive Go regular expression, which runs in linear
  time, and falls back to a literal match when it does not compile.
- `chrome.SaveDump` writes a new file, mode `0600`, inside the directory the
  app passes. The name is reduced to `[a-z0-9._-]`, and an existing file is
  never overwritten.

## Supported versions

tuikit is pre-stable (`v0.0.x`). Only the latest tag receives fixes.

## Verifying a release

A release is a signed git tag; there are no binaries. `go get` checks every
module version against the [Go checksum database](https://sum.golang.org),
so the code you fetch is the code that was tagged. To check the tag itself:

```sh
git fetch --tags https://github.com/blairham/tuikit
git verify-tag v0.0.17
```

GitHub shows each tag's signature as Verified on the
[tags page](https://github.com/blairham/tuikit/tags).

From v0.0.18 on, each tag also has a GitHub release built by
`.github/workflows/release.yml`: the tagged source as
`tuikit-X.Y.Z.tar.gz`, a `checksums.txt` signed with
[cosign](https://github.com/sigstore/cosign) keyless signing — tied to the
workflow that built it, not to a key someone could leak — and SLSA build
provenance for the archive. Verify the signature, then the archive against
it, then the provenance:

```sh
VERSION=v0.0.18
cosign verify-blob \
  --certificate-identity "https://github.com/blairham/tuikit/.github/workflows/release.yml@refs/tags/$VERSION" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --bundle checksums.txt.sigstore.json checksums.txt
sha256sum --check --ignore-missing checksums.txt
gh attestation verify "tuikit-${VERSION#v}.tar.gz" --repo blairham/tuikit
```

The provenance bundle is also attached to the release as
`tuikit-$VERSION.intoto.jsonl`, for checking offline with
`gh attestation verify --bundle`.

## Reporting a vulnerability

**Do not open a public issue.** Report it privately through GitHub:
[Security → Report a vulnerability](https://github.com/blairham/tuikit/security/advisories/new).

Please include the affected version or commit, what an attacker can do, and
the steps to reproduce. You should receive a response within a week.

In scope, among others:

- input that makes a tuikit function panic or hang — typed into the filter
  or command bar, or passed as content — since that takes down the app
- `SaveDump` writing outside the directory it was given, or overwriting a
  file
- the SGR rewriters emitting any sequence other than a color, or dropping
  or altering bytes that are not part of one

Out of scope: escape sequences in content an app draws without stripping
them first (see above), and anything that requires already controlling the
app's process or its configuration.
