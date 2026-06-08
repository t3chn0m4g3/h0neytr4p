package h0neytr4p

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Taken from https://stackoverflow.com/questions/55300117/how-do-i-find-all-files-that-have-a-certain-extension-in-go-regardless-of-depth
func GetAllTraps(root, ext string) ([]string, error) {
	var traps []string
	if err := filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) == ext {
			traps = append(traps, s)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(traps)
	return traps, nil
}

func ParseTraps(traps string) ([]Trap, error) {
	var trapConfigs []Trap
	trapFiles, err := GetAllTraps(traps, ".json")
	if err != nil {
		return nil, fmt.Errorf("list trap files: %w", err)
	}
	for _, trap := range trapFiles {
		trapData := Trap{}
		trapFile, err := os.ReadFile(trap)
		if err != nil {
			return nil, fmt.Errorf("read trap %s: %w", trap, err)
		}
		if err := json.Unmarshal(trapFile, &trapData); err != nil {
			return nil, fmt.Errorf("parse trap %s: %w", trap, err)
		}
		trapConfigs = append(trapConfigs, trapData)
	}
	return trapConfigs, nil
}
