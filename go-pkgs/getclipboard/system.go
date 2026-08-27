package getclipboard

import "golang.design/x/clipboard"

type systemSource struct{}

// System returns the real OS clipboard source.
func System() Source {
	return systemSource{}
}

func (systemSource) Init() error {
	return clipboard.Init()
}

func (systemSource) ReadImage() []byte {
	return clipboard.Read(clipboard.FmtImage)
}

func (systemSource) ReadText() []byte {
	return clipboard.Read(clipboard.FmtText)
}

func (systemSource) ExtraFormats() []Format {
	known := []clipboard.Format{
		clipboard.Register("public.svg"),
		clipboard.Register("public.html"),
		clipboard.Register("public.rtf"),
		clipboard.Register("com.adobe.pdf"),
	}
	seen := map[clipboard.Format]bool{
		clipboard.FmtText:  true,
		clipboard.FmtImage: true,
	}
	var out []Format
	add := func(f clipboard.Format) {
		if seen[f] {
			return
		}
		seen[f] = true
		fmt := f
		out = append(out, Format{
			MIME: fmt.MIME(),
			Read: func() []byte { return clipboard.Read(fmt) },
		})
	}
	for _, f := range clipboard.Formats() {
		add(f)
	}
	for _, f := range known {
		add(f)
	}
	return out
}
