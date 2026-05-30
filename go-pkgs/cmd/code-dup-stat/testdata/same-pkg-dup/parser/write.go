package parser

func WriteConfig(path string, cfg *Config) error {
	f, err := createFile(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return encodeConfig(f, cfg)
}
