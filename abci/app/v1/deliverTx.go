package app

import (
	"abci-example/abci/code"
	"encoding/base64"
	"fmt"

	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// DeliverTxRouter is Pointer to function
func (app *Application) DeliverTxRouter(method string, param string, nonce []byte, signature []byte, nodeID string) types.ResponseDeliverTx {
	// ---- check authorization ----
	checkTxResult := app.CheckTxRouter(method, param, nonce, signature, nodeID)
	if checkTxResult.Code != code.OK {
		if checkTxResult.Log != "" {
			return app.ReturnDeliverTxLog(checkTxResult.Code, checkTxResult.Log, "")
		}
		return app.ReturnDeliverTxLog(checkTxResult.Code, "Unauthorized", "")
	}
	result := app.callDeliverTx(method, param, nodeID)

	// Set used nonce to stateDB
	emptyValue := make([]byte, 0)
	app.SetStateDB([]byte(nonce), emptyValue)
	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
	app.deliverTxNonceState[nonceBase64] = []byte(nil)
	return result
}

func (app *Application) callDeliverTx(name string, param string, nodeID string) types.ResponseDeliverTx {
	switch name {
	case "SetKeyValue":
		return app.setKeyValue(param, nodeID)
	default:
		return types.ResponseDeliverTx{Code: code.UnknownMethod, Log: "Unknown method name"}
	}
}

// app.ReturnDeliverTxLog return types.ResponseDeliverTx
func (app *Application) ReturnDeliverTxLog(code uint32, log string, extraData string) types.ResponseDeliverTx {
	var tags []cmn.KVPair
	if code == 0 {
		var tag cmn.KVPair
		tag.Key = []byte("success")
		tag.Value = []byte("true")
		tags = append(tags, tag)
	} else {
		var tag cmn.KVPair
		tag.Key = []byte("success")
		tag.Value = []byte("false")
		tags = append(tags, tag)
	}
	return types.ResponseDeliverTx{
		Code: code,
		Log:  fmt.Sprintf(log),
		Data: []byte(extraData),
		Tags: tags,
	}
}
