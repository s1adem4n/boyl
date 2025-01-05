package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_879072730")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(12, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text2462348188",
			"max": 0,
			"min": 0,
			"name": "provider",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(13, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text1044722854",
			"max": 0,
			"min": 0,
			"name": "providerId",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_879072730")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("text2462348188")

		// remove field
		collection.Fields.RemoveById("text1044722854")

		return app.Save(collection)
	})
}
