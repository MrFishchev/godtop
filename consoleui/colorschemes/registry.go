package colorschemes

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/shibukawa/configdir"
)

func Get(configDir configdir.ConfigDir, name string) (Colorscheme, error) {
	scheme, err := getCustomColorScheme(configDir, name)
	return scheme, err
}

//getCustomColorScheme tries to read a custom json colorscheme from <configDir>/<name>.json
func getCustomColorScheme(configDir configdir.ConfigDir, name string) (Colorscheme, error) {
	var scheme Colorscheme
	filename := name + ".json"
	folder := configDir.QueryFolderContainsFile(filename)
	if folder == nil {
		paths := make([]string, 0)
		for _, dir := range configDir.QueryFolders(configdir.Existing) {
			paths = append(paths, dir.Path)
		}
		return scheme, fmt.Errorf("unable to find colorcheme %v (%v)", filename, strings.Join(paths, ","))
	}

	data, err := folder.ReadFile(filename)
	if err != nil {
		return scheme, fmt.Errorf("unable to read colorscheme %v: %v", filepath.Join(folder.Path, filename), err.Error())
	}

	err = json.Unmarshal(data, &scheme)
	if err != nil {
		return scheme, fmt.Errorf("unable to parse colorscheme: %v", err.Error())
	}

	return scheme, nil
}
