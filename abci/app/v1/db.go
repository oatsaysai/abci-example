package app

func (app *Application) SetStateDB(key, value []byte) {
	app.HashData = append(app.HashData, key...)
	app.HashData = append(app.HashData, value...)
	app.UncommittedState[string(key)] = value
}

func (app *Application) GetStateDB(key []byte) (err error, value []byte) {
	var existInUncommittedState bool
	value, existInUncommittedState = app.UncommittedState[string(key)]
	if !existInUncommittedState {
		value = app.state.db.Get(key)
	}
	return nil, value
}

func (app *Application) GetCommittedStateDB(key []byte) (err error, value []byte) {
	value = app.state.db.Get(key)
	return nil, value
}

func (app *Application) HasStateDB(key []byte) bool {
	_, existInUncommittedState := app.UncommittedState[string(key)]
	if existInUncommittedState {
		return true
	}
	return app.state.db.Has(key)
}

func (app *Application) HasVersionedStateDB(key []byte) bool {
	versionsKeyStr := string(key) + "|versions"
	versionsKey := []byte(versionsKeyStr)

	_, existInUncommittedState := app.UncommittedVersionsState[versionsKeyStr]
	if existInUncommittedState {
		return true
	}

	return app.state.db.Has(versionsKey)
}

func (app *Application) DeleteStateDB(key []byte) {
	if !app.HasStateDB(key) {
		return
	}
	app.HashData = append(app.HashData, key...)
	app.HashData = append(app.HashData, []byte("delete")...) // Remove or replace with something else?
	app.UncommittedState[string(key)] = nil
}

func (app *Application) SaveDBState() {
	batch := app.state.db.NewBatch()
	defer batch.Close()
	for key := range app.UncommittedState {
		value := app.UncommittedState[key]
		if value != nil {
			batch.Set([]byte(key), value)
		} else {
			batch.Delete([]byte(key))
		}
	}
	batch.WriteSync()
	app.UncommittedState = make(map[string][]byte)
	app.UncommittedVersionsState = make(map[string][]int64)
}
