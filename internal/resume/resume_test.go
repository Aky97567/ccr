package resume

import "testing"

func TestShellCommandQuotesSpaces(t *testing.T) {
	got := ShellCommand("/Users/akshay/Code/personal/cv repos/1c2p", "cc4e694a-ff12-422c-ae40-447ca532ad0f")
	want := "cd '/Users/akshay/Code/personal/cv repos/1c2p' && claude --resume 'cc4e694a-ff12-422c-ae40-447ca532ad0f'"
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestShellCommandEscapesSingleQuote(t *testing.T) {
	got := ShellCommand("/tmp/it's here", "id")
	want := `cd '/tmp/it'\''s here' && claude --resume 'id'`
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestShellQuoteEmpty(t *testing.T) {
	if shellQuote("") != "''" {
		t.Errorf("empty string should quote to '', got %q", shellQuote(""))
	}
}
