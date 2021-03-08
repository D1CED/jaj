package query

import (
	"github.com/Jguer/yay/v10/pkg/stringset"
	"github.com/Jguer/yay/v10/pkg/text"
)

type AURWarnings struct {
	Orphans   []string
	OutOfDate []string
	Missing   []string
	Ignore    stringset.StringSet
}

func NewWarnings() *AURWarnings {
	return &AURWarnings{Ignore: stringset.Make()}
}

func (warnings *AURWarnings) Print() {
	if len(warnings.Missing) > 0 {
		text.Warn(text.T("Missing AUR Packages:"))
		printRange(warnings.Missing)
	}

	if len(warnings.Orphans) > 0 {
		text.Warn(text.T("Orphaned AUR Packages:"))
		printRange(warnings.Orphans)
	}

	if len(warnings.OutOfDate) > 0 {
		text.Warn(text.T("Flagged Out Of Date AUR Packages:"))
		printRange(warnings.OutOfDate)
	}
}

func printRange(names []string) {
	for _, name := range names {
		text.Print("  " + text.Cyan(name))
	}
	text.Println()
}
