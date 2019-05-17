package app

import (
	"abci-example/abci/code"
	"fmt"

	"github.com/tendermint/tendermint/abci/types"
)

// ReturnCheckTx return types.ResponseDeliverTx
func (app *Application) ReturnCheckTx(code uint32, log string) types.ResponseCheckTx {
	return types.ResponseCheckTx{
		Code: code,
		Log:  fmt.Sprintf(log),
	}
}

var IsMethod = map[string]bool{
	"SetKeyValue": true,
}

// CheckTxRouter is Pointer to function
func (app *Application) CheckTxRouter(method string, param string, nonce []byte, signature []byte, nodeID string) types.ResponseCheckTx {
	// Bypass
	return app.ReturnCheckTx(code.OK, "OK")
}
