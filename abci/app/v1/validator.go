package app

import (
	"abci-example/abci/code"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/abci/types"
)

const (
	ValidatorSetChangePrefix string = "val:"
)

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ValidatorSetChangePrefix)
}

func (app *Application) Validators() (validators []types.Validator) {
	app.logger.Infof("Validators")
	itr := app.state.db.Iterator(nil, nil)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		validator := new(types.Validator)
		err := types.ReadMessage(bytes.NewBuffer(key), validator)
		if err != nil {
			panic(err)
		}
		validators = append(validators, *validator)
	}
	return
}

// add, update, or remove a validator
func (app *Application) updateValidator(v types.ValidatorUpdate) types.ResponseDeliverTx {
	pubKeyBase64 := base64.StdEncoding.EncodeToString(v.PubKey.GetData())
	key := []byte("val:" + pubKeyBase64)
	if v.Power == 0 {
		// remove validator
		if !app.HasStateDB(key) {
			return app.ReturnDeliverTxLog(code.Unauthorized, fmt.Sprintf("Cannot remove non-existent validator %X", key), "")
		}
		app.DeleteStateDB(key)
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&v, value); err != nil {
			return app.ReturnDeliverTxLog(code.EncodingError, fmt.Sprintf("Error encoding validator: %v", err), "")
		}
		app.SetStateDB(key, value.Bytes())
	}
	app.ValUpdates[pubKeyBase64] = v
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *Application) setValidator(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetValidator, Parameter: %s", param)
	var funcParam SetValidatorParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	pubKey, err := base64.StdEncoding.DecodeString(string(funcParam.PublicKey))
	if err != nil {
		return app.ReturnDeliverTxLog(code.DecodingError, err.Error(), "")
	}
	var pubKeyObj types.PubKey
	pubKeyObj.Type = "ed25519"
	pubKeyObj.Data = pubKey
	var newValidator types.ValidatorUpdate
	newValidator.PubKey = pubKeyObj
	newValidator.Power = funcParam.Power
	return app.updateValidator(newValidator)
}
