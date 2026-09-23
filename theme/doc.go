// Package theme provides the color palette, lipgloss style helpers, and
// the [Theme] struct that consumer apps inject at construction.
//
// Every other tuikit package reads from [Theme]: chrome renders with it,
// table styles cells with it, viewfsm dispatches help text using it.
// Apps can supply [Default] for the canonical k9s-style look, or build
// their own theme to recolor the chrome wholesale.
//
// The [Theme.PaintBackground] toggle distinguishes the two terminal
// aesthetics tuikit supports:
//
//   - true  → every styled span paints Background([Theme.Bg]). Necessary
//     on terminals where the default background bleeds gray through
//     unstyled gaps (macOS Terminal in particular).
//
//   - false → only the outer screen wrapper sets the background. Lets
//     a bubbles table use [bubbles/v2/table.Styles.Selected.Foreground]
//     to override per-cell foregrounds without a background fight.
//
// Both modes are first-class. Default is true.
package theme
