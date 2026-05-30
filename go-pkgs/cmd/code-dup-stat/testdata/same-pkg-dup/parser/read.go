package parser

type Config struct {
	Name string
	Port int
}

func ReadConfig(path string) (*Config, error) {
	f, err := openFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return decodeConfig(f)
}
