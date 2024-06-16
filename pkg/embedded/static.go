package embedded

import (
	"embed"
	"io/fs"
)

//go:embed static
var static embed.FS

var Static = staticFs{static}

// staticFs is a wrapper around embed.FS that replace m.js by wa.js to maintain backward
// compatibility.
type staticFs struct {
	static embed.FS
}

// Open implements fs.FS.
func (sfs staticFs) Open(name string) (fs.File, error) {
	if name == "static/m.js" {
		return sfs.static.Open("static/wa.js")
	}

	return sfs.static.Open(name)
}

// Open implements fs.FS.
func (sfs staticFs) ReadFile(name string) ([]byte, error) {
	if name == "static/m.js" {
		return sfs.static.ReadFile("static/wa.js")
	}

	return sfs.static.ReadFile(name)
}
