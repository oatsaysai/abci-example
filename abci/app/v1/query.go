package app

import (
	"abci-example/abci/code"

	"github.com/tendermint/tendermint/abci/types"
)

// ReturnQuery return types.ResponseQuery
func (app *Application) ReturnQuery(value []byte, log string, height int64) types.ResponseQuery {
	app.logger.Infof("Query result: %s", string(value))
	var res types.ResponseQuery
	res.Value = value
	res.Log = log
	res.Height = height
	return res
}

// QueryRouter is Pointer to function
func (app *Application) QueryRouter(method string, param string, height int64) types.ResponseQuery {
	result := app.callQuery(method, param, height)
	return result
}

func (app *Application) callQuery(name string, param string, height int64) types.ResponseQuery {
	switch name {
	case "GetKeyValue":
		return app.getKeyValue(param)
	default:
		return types.ResponseQuery{Code: code.UnknownMethod, Log: "Unknown method name"}
	}
}

