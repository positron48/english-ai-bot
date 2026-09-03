package placementbundle

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
	"tgbot-skeleton/internal/placement"
)

//go:embed en/bank.json es/bank.json
var files embed.FS
var cache sync.Map

func Load(course string) (*placement.Bank, error) {
	if course != "es_ru" && course != "en_ru" {
		return nil, fmt.Errorf("unsupported placement course %q", course)
	}
	if b, ok := cache.Load(course); ok {
		return b.(*placement.Bank), nil
	}
	raw, err := files.ReadFile(course[:2] + "/bank.json")
	if err != nil {
		return nil, err
	}
	var b placement.Bank
	if err = json.Unmarshal(raw, &b); err != nil {
		return nil, err
	}
	if err = b.Validate(); err != nil {
		return nil, err
	}
	cache.Store(course, &b)
	return &b, nil
}
