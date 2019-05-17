package common

import (
	"abci-example/abci/app/v1"
	"abci-example/test/data"
	"abci-example/utils"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
)

func SetKeyValue(t *testing.T, nodeID, privK string, param app.SetKeyValueParam) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "SetKeyValue"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s, Key: %s", fnName, param.Key)
}

func GetKeyValue(t *testing.T, param app.GetKeyValueParam, expected string) {
	fnName := "GetKeyValue"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res app.GetKeyValueResult
	err = json.Unmarshal([]byte(resultString), &res)
	if err != nil {
		fmt.Println("error:", err)
	}
	if actual := res.Value; actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %s\nActual: %s", fnName, expected, actual)
	}
	t.Logf("PASS: %s, Key: %s", fnName, param.Key)
}

func TestSetKeyValue(t *testing.T, key, value string) {
	var param app.SetKeyValueParam
	param.Key = key
	param.Value = value
	SetKeyValue(t, data.NodeID, data.PrivateKey, param)
}

func TestGetKeyValue(t *testing.T, key, value string) {
	var param app.GetKeyValueParam
	param.Key = key
	GetKeyValue(t, param, value)
}
