// Package chrome renders the visual frame around a tuikit app: top
// section (info-panel + shortcut grid + ASCII logo), bordered content
// with an injected title, breadcrumb footer, status bar, filter and
// command bars, and the help overlay.
//
// Apps assemble a [Frame] each tick and call [Chrome.Render]. The chrome
// owns the layout, the [Theme] applies the colors, and the app owns the
// content. No control inversion — apps still drive their own
// [tea.Program] update loop.
package chrome
