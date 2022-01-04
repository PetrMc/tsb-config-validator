package validator

import (
	"strings"

	tsize "github.com/kopoli/go-terminal-size"
)

type PrintFields struct {
	Width                    int
	Stars, Indent, Twoindent string
}

func CustomPrint() PrintFields {
	p := PrintFields{}
	s, _ := tsize.GetSize()
	p.Width = s.Width
	p.Stars = "\n" + strings.Repeat("*", s.Width) + "\n"
	p.Indent = strings.Repeat(" ", 2)
	p.Twoindent = strings.Repeat(p.Indent, 2)
	return p
}
