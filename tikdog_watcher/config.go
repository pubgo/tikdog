package tikdog_watcher

import (
	"github.com/imdario/mergo"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"path/filepath"
)

type config struct {
	Debug          bool
	Root           string   `toml:"root"`
	ExcludePattern []string `toml:"exclude_pattern"`
}

func initConfig(path string) (cfg *config, err error) {
	if path == "" {
		cfg, err = defaultPathConfig()
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = readConfigOrDefault(path)
		if err != nil {
			return nil, err
		}
	}
	err = mergo.Merge(cfg, defaultConfig())
	if err != nil {
		return nil, err
	}
	err = cfg.preprocess()
	return cfg, err
}

func defaultConfig() config {
	return config{
		Root: ".",
	}
}

func readConfig(path string) (*config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := new(config)
	if err = toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func readConfigOrDefault(path string) (*config, error) {
	dftCfg := defaultConfig()
	cfg, err := readConfig(path)
	if err != nil {
		return &dftCfg, err
	}

	return cfg, nil
}

func (c *config) rel(path string) string {
	s, err := filepath.Rel(c.Root, path)
	if err != nil {
		return ""
	}
	return s
}
