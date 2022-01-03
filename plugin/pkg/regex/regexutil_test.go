package regexutil

import "testing"

func TestRegexSplit(t *testing.T) {
	v := RegexSplit("a  abc    fff \t cc \t\t   abc", "\\s+")
	t.Log(v, len(v))
}
