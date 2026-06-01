package migrator

import "testing"

// TestUseRefEmptyArgRegex confirms the regex matches generics with nested
// brackets (the original failure mode that motivated this fix).
func TestUseRefEmptyArgRegex(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{
			`const r = useRef<string>()`,
			`const r = useRef<string | undefined>(undefined)`,
		},
		{
			`const r = useRef<ReturnType<typeof setTimeout>>()`,
			`const r = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)`,
		},
		{
			`const r = useRef<HTMLDivElement | null>(null)`, // already typed: leave alone
			`const r = useRef<HTMLDivElement | null>(null)`,
		},
		{
			`const r = useRef<Map<string, number>>()`,
			`const r = useRef<Map<string, number> | undefined>(undefined)`,
		},
		{
			`const r = useRef<X | undefined>(undefined)`, // already has undefined: leave alone
			`const r = useRef<X | undefined>(undefined)`,
		},
	}

	for _, tc := range cases {
		got := useRefEmptyArgRe.ReplaceAllStringFunc(tc.input, func(match string) string {
			m := useRefEmptyArgRe.FindStringSubmatch(match)
			if len(m) < 2 {
				return match
			}
			t := m[1]
			if containsUndefined(t) {
				return match
			}
			return "useRef<" + t + " | undefined>(undefined)"
		})
		if got != tc.want {
			t.Errorf("\ninput: %s\n got: %s\nwant: %s", tc.input, got, tc.want)
		}
	}
}

func containsUndefined(s string) bool {
	for i := 0; i+len("undefined") <= len(s); i++ {
		if s[i:i+len("undefined")] == "undefined" {
			return true
		}
	}
	return false
}
