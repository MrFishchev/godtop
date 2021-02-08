package config

import (
	"godtop/consoleui/colorschemes"
	"log"
	"path/filepath"
	"time"

	"github.com/shibukawa/configdir"
)

const CONFFILE = "godtop.conf"

type Config struct {
	ConfigDir            configdir.ConfigDir
	ConfigFile           string
	GraphHorizontalScale int
	Colorscheme          colorschemes.Colorscheme
	UpdateInterval       time.Duration
	StatusBar            bool
}

func NewConfig() Config {
	configDir := configdir.New("", "godtop")
	configDir.LocalPath, _ = filepath.Abs("./consoleui/config")
	config := Config{
		ConfigDir:            configDir,
		GraphHorizontalScale: 7,
		UpdateInterval:       time.Second,
		StatusBar:            false,
	}

	scheme, err := colorschemes.Get(config.ConfigDir, "default")
	if err != nil {
		log.Println("cannot find colorscheme default")
	}
	config.Colorscheme = scheme

	folder := config.ConfigDir.QueryFolderContainsFile(CONFFILE)
	if folder != nil {
		config.ConfigFile = filepath.Join(folder.Path, CONFFILE)
	}
	return config
}
