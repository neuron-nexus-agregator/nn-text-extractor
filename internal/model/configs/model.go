package configs

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	ContentSelectors  []string `json:"content_selectors"`
	ParagraphSelector string   `json:"paragraph_selector"`
	TextDivSelector   string   `json:"text_div_selector"`
	TableSelector     string   `json:"table_selector"`
	ListItemSelector  string   `json:"list_item_selector"`
}

type SiteConfigs map[string]Config

type Configs struct {
	SiteConfigs SiteConfigs
}

func (cfg *Configs) Init(configFilePath string) error {
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл конфигураций %s: %w", configFilePath, err)
	}
	defer configFile.Close()
	var configs SiteConfigs
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&configs)
	if err != nil {
		return fmt.Errorf("не удалось декодировать файл конфигураций %s: %w", configFilePath, err)
	}
	cfg.SiteConfigs = configs
	return nil
}
