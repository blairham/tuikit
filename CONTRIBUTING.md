# Contributing to tuikit

Thanks for the interest. tuikit is in early `v0.0.x` development — the API is intentionally fluid while the applications that use it migrate onto it. Once the surface settles we'll tag `v0.1.0` and the API becomes more stable.

## Quick start

```bash
make check    # gofumpt + fieldalignment + golangci-lint + race tests
```

## What we want

- **Bug reports** with a minimal repro Go file. The TUI surface is small; reproductions should be tractable.
- **Cross-terminal screenshots** (macOS Terminal, iTerm2, Alacritty, kitty, Wezterm, ghostty). The `PaintBackground` knob in `theme` is load-bearing across terminals; we want to keep both modes correct.
- **Theme contributions** — Nord, Solarized Light/Dark, mono, etc. Drop a new constructor in `theme/`.
- **Example apps** in `examples/` showing a specific pattern (split-pane, form view, async data, etc.).

## What we DON'T want (yet)

- Major API surface additions while `v0.0.x` is unstable. File an issue first.
- Domain-specific features that only one app needs — keep those in your own app.
- New dependencies. The current dep tree is just Charm + image/color; we want to keep it tight.

## Workflow

1. Open an issue describing the change.
2. Fork, branch from `main`, write code + tests.
3. `make check` green locally.
4. Open a PR with a clear "why" (the "what" is in the diff).

## License

By contributing, you agree your contributions will be licensed under the Apache-2.0 License.
