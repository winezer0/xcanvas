// Package embeds provides embedded rules for framework and component detection.
package embeds_frame

import (
	"embed"
)

//go:embed *.yml
var FrameEmbedFS embed.FS
