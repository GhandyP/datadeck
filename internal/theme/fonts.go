package theme

import (
	"embed"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
)

// The faces this system uses, embedded so neither surface depends on a font
// host: the app serves them from its own origin and the exported deck carries
// them inside the file.
//
//go:embed fonts/*.woff2
var fontFiles embed.FS

type face struct {
	family string
	weight string
	file   string
}

var faces = []face{
	{"Archivo", "400", "fonts/archivo-400.woff2"},
	{"Archivo", "600", "fonts/archivo-600.woff2"},
	{"Archivo Narrow", "600", "fonts/archivo-narrow-600.woff2"},
}

// FontCSS returns the @font-face rules. When inline is true the font bytes are
// embedded as data URIs, which is what the exported deck needs in order to stay
// one file that reaches no other host.
func FontCSS(inline bool) string {
	var out strings.Builder
	for _, f := range faces {
		// An embedded file cannot be missing at run time, so a failure here is a
		// build mistake rather than a condition to handle.
		data, err := fontFiles.ReadFile(f.file)
		if err != nil {
			panic(fmt.Sprintf("theme: embedded font %s is unreadable: %v", f.file, err))
		}
		source := "url(/fonts/" + strings.TrimPrefix(f.file, "fonts/") + ") format('woff2')"
		if inline {
			source = "url(data:font/woff2;base64," + base64.StdEncoding.EncodeToString(data) + ") format('woff2')"
		}
		fmt.Fprintf(&out, "@font-face{font-family:'%s';font-style:normal;font-weight:%s;font-display:swap;src:%s}\n",
			f.family, f.weight, source)
	}
	return out.String()
}

// FontNames lists the embedded font files by the name they are served under.
func FontNames() []string {
	names := make([]string, 0, len(faces))
	for _, f := range faces {
		names = append(names, strings.TrimPrefix(f.file, "fonts/"))
	}
	sort.Strings(names)
	return names
}

// Font returns one embedded font by its served name.
func Font(name string) ([]byte, bool) {
	for _, f := range faces {
		if strings.TrimPrefix(f.file, "fonts/") == name {
			data, err := fontFiles.ReadFile(f.file)
			if err != nil {
				return nil, false
			}
			return data, true
		}
	}
	return nil, false
}
