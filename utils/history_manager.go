package utils

import (
	"encoding/json"
	"os"
	"path"
)

type HistoryManager struct {
	filePath string
}

// 视图函数
type HistoryEntry struct {
	OriginURL string `json:"originUrl"`
	LastURL   string `json:"lastUrl"`
	Cursor    int    `json:"cursor"`
}

func NewHistoryManager() *HistoryManager {
	historyDir := path.Join(os.Getenv("HOME"), ".nvrd")
	historyFile := path.Join(historyDir, "history.json")

	// 确保目录存在
	if _, err := os.Stat(historyDir); os.IsNotExist(err) {
		os.MkdirAll(historyDir, 0755)
	}

	return &HistoryManager{
		filePath: historyFile,
	}
}

func (hm *HistoryManager) Load() (HistoryEntry, error) {
	var history HistoryEntry

	if data, err := os.ReadFile(hm.filePath); err == nil {
		err = json.Unmarshal(data, &history)
		if err != nil {
			return HistoryEntry{}, err
		}
	}

	return history, nil
}

func (hm *HistoryManager) Save(history HistoryEntry) error {
	data, err := json.Marshal(history)
	if err != nil {
		return err
	}

	return os.WriteFile(hm.filePath, data, 0644)
}
