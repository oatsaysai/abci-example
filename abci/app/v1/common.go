package app

import (
	"abci-example/abci/code"
	"encoding/json"

	"github.com/tendermint/tendermint/abci/types"
)

func (app *Application) setKeyValue(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetKeyValue, Parameter: %s", param)
	var funcParam SetKeyValueParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(funcParam.Key), []byte(funcParam.Value))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *Application) getKeyValue(param string) types.ResponseQuery {
	app.logger.Infof("GetKeyValue, Parameter: %s", param)
	var funcParam GetKeyValueParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetKeyValueResult
	_, value := app.GetCommittedStateDB([]byte(funcParam.Key))
	if value == nil {
		valueJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(valueJSON, "not found", app.state.Height)
	}
	result.Value = string(value)
	valueJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(valueJSON, "success", app.state.Height)
}
