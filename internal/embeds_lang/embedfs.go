// Package lcembeds provides embedded rules for language classification.
package embeds_lang

import (
	"embed"
)

//go:embed *.yml
var LanguageEmbedFS embed.FS
