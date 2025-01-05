package settings

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
)

type Settings struct {
	app        core.App
	collection *core.Collection
}

func NewSettings(app core.App, collection *core.Collection) *Settings {
	return &Settings{
		app:        app,
		collection: collection,
	}
}

func (s *Settings) Get(key string) (any, error) {
	record, err := s.app.FindFirstRecordByData(s.collection, "key", key)
	if err != nil {
		return "", err
	}

	var value any
	if err := json.Unmarshal([]byte(record.GetString("value")), &value); err != nil {
		return "", err
	}
	return value, nil
}

func (s *Settings) GetString(key string) string {
	value, err := s.Get(key)
	if err != nil {
		return ""
	}

	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

func (s *Settings) Set(key, value any) error {
	marshaled, err := json.Marshal(value)
	if err != nil {
		return err
	}

	record, err := s.app.FindFirstRecordByData(s.collection, "key", key)
	if err != nil {
		record = core.NewRecord(s.collection)
		record.Set("key", key)
	}
	record.Set("value", string(marshaled))
	return s.app.Save(record)
}
