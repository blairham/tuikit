package chrome

import (
	"testing"

	"github.com/blairham/tuikit/theme"
)

// The bars carry a message for the status bar, not a failure: a type with an
// Error() string method satisfies error, so fmt would print the message in
// place of the bar and errors.As would match it.
func TestBarsAreNotErrors(t *testing.T) {
	t.Parallel()
	th := theme.Default()
	for name, v := range map[string]any{
		"CommandBar": NewCommandBar(th, CommandBarOpts{}),
		"Prompt":     NewPrompt(th, PromptOpts{}),
	} {
		if _, ok := v.(error); ok {
			t.Errorf("*%s implements error; name the accessor ErrMsg", name)
		}
	}
}
