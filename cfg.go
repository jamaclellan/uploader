package uploader

type Config struct {
	BaseURL    string   `yaml:"base_url"`
	BoltConfig *boltCfg `yaml:"bolt"`
	DirConfig  *dirCfg  `yaml:"dir"`
}

type boltCfg struct {
	Path string `yaml:"path"`
}

type dirCfg struct {
	Path string `yaml:"path"`
}
