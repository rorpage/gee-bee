package geebee

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// LoadState reads the list of aircraft considered "currently spotted" from a
// previous run. A missing or unparseable file is treated as an empty state
// so a single corrupt/absent file never blocks a run.
func LoadState(path string) []Aircraft {
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("could not read state file %s, starting fresh: %v", path, err)
		}
		return nil
	}

	var aircraft []Aircraft
	if err := json.Unmarshal(data, &aircraft); err != nil {
		log.Printf("could not parse state file %s, starting fresh: %v", path, err)
		return nil
	}

	return aircraft
}

// SaveState persists the aircraft currently considered "spotted" so the next
// run (a fresh process, since gee-bee runs one-shot) knows not to re-notify
// for aircraft still in range. Aircraft that drop out of range are dropped
// from this list by the caller before saving, so they will notify again if
// spotted later.
func SaveState(path string, aircraft []Aircraft) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	data, err := json.MarshalIndent(aircraft, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
