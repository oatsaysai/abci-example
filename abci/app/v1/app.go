package app

import (
	"abci-example/abci/code"
	"abci-example/abci/version"
	protoTm "abci-example/protos/tendermint"
	"abci-example/utils"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var (
	stateKey = []byte("stateKey")
)

type State struct {
	db      dbm.DB
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

var _ types.Application = (*Application)(nil)

type Application struct {
	types.BaseApplication
	state                    State
	checkTxNonceState        map[string][]byte
	deliverTxNonceState      map[string][]byte
	ValUpdates               map[string]types.ValidatorUpdate
	logger                   *logrus.Entry
	Version                  string
	AppProtocolVersion       uint64
	CurrentBlock             int64
	CurrentChain             string
	HashData                 []byte
	UncommittedState         map[string][]byte
	UncommittedVersionsState map[string][]int64
}

func loadState(db dbm.DB) State {
	stateBytes := db.Get(stateKey)
	var state State
	if len(stateBytes) != 0 {
		err := json.Unmarshal(stateBytes, &state)
		if err != nil {
			panic(err)
		}
	}
	state.db = db
	return state
}

func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	state.db.Set(stateKey, stateBytes)
}

func NewApplication(logger *logrus.Entry, db dbm.DB) *Application {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("%s", identifyPanic())
			panic(r)
		}
	}()
	state := loadState(db)
	ABCIVersion := version.Version
	ABCIProtocolVersion := version.AppProtocolVersion
	logger.Infof("Start ABCI version: %s", ABCIVersion)
	return &Application{
		state:                    state,
		checkTxNonceState:        make(map[string][]byte),
		deliverTxNonceState:      make(map[string][]byte),
		logger:                   logger,
		Version:                  ABCIVersion,
		AppProtocolVersion:       ABCIProtocolVersion,
		UncommittedState:         make(map[string][]byte),
		UncommittedVersionsState: make(map[string][]int64),
		ValUpdates:               make(map[string]types.ValidatorUpdate),
	}
}

func (app *Application) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	utils.GetEnv("", "")
	var res types.ResponseInfo
	res.Version = app.Version
	res.LastBlockHeight = app.state.Height
	res.LastBlockAppHash = app.state.AppHash
	res.AppVersion = app.AppProtocolVersion
	app.CurrentBlock = app.state.Height
	return res
}

// Save the validators in the merkle tree
func (app *Application) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Error("Error updating validators", "r", r)
		}
	}
	return types.ResponseInitChain{}
}

// Track the block hash and header information
func (app *Application) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.logger.Infof("BeginBlock: %d, Chain ID: %s", req.Header.Height, req.Header.ChainID)
	app.CurrentBlock = req.Header.Height
	app.CurrentChain = req.Header.ChainID

	// reset valset changes
	app.ValUpdates = make(map[string]types.ValidatorUpdate, 0)
	return types.ResponseBeginBlock{}
}

// Update the validator set
func (app *Application) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	app.logger.Infof("EndBlock: %d", req.Height)
	valUpdates := make([]types.ValidatorUpdate, 0)
	for _, newValidator := range app.ValUpdates {
		valUpdates = append(valUpdates, newValidator)
	}
	return types.ResponseEndBlock{ValidatorUpdates: valUpdates}
}

func (app *Application) DeliverTx(tx []byte) (res types.ResponseDeliverTx) {
	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorf("Recovered in %s, %s", r, identifyPanic())
			res = app.ReturnDeliverTxLog(code.UnknownError, "Unknown error", "")
		}
	}()
	var txObj protoTm.Tx
	err := proto.Unmarshal(tx, &txObj)
	if err != nil {
		app.logger.Error(err.Error())
	}
	method := txObj.Method
	param := txObj.Params
	nonce := txObj.Nonce
	signature := txObj.Signature
	nodeID := txObj.NodeId

	// ---- Check duplicate nonce ----
	nonceDup := app.isDuplicateNonce(nonce)
	if nonceDup {
		return app.ReturnDeliverTxLog(code.DuplicateNonce, "Duplicate nonce", "")
	}
	app.logger.Infof("DeliverTx: %s, NodeID: %s", method, nodeID)
	if method != "" {
		result := app.DeliverTxRouter(method, param, nonce, signature, nodeID)
		app.logger.Infof(`DeliverTx response: {"code":%d,"log":"%s","tags":[{"key":"%s","value":"%s"}]}`, result.Code, result.Log, string(result.Tags[0].Key), string(result.Tags[0].Value))
		return result
	}
	return app.ReturnDeliverTxLog(code.MethodCanNotBeEmpty, "method can not be empty", "")
}

func (app *Application) CheckTx(tx []byte) (res types.ResponseCheckTx) {
	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorf("Recovered in %s, %s", r, identifyPanic())
			res = app.ReturnCheckTx(code.UnknownError, "Unknown error")
		}
	}()
	var txObj protoTm.Tx
	err := proto.Unmarshal(tx, &txObj)
	if err != nil {
		app.logger.Error(err.Error())
	}
	method := txObj.Method
	param := txObj.Params
	nonce := txObj.Nonce
	signature := txObj.Signature
	nodeID := txObj.NodeId
	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)

	// ---- Check duplicate nonce ----
	nonceDup := app.isDuplicateNonce(nonce)
	if nonceDup {
		res = app.ReturnCheckTx(code.DuplicateNonce, "Duplicate nonce")
		return res
	}

	// Check duplicate nonce in checkTx stateDB
	_, exist := app.checkTxNonceState[nonceBase64]
	if !exist {
		app.checkTxNonceState[nonceBase64] = []byte(nil)
	} else {
		res = app.ReturnCheckTx(code.DuplicateNonce, "Duplicate nonce")
		return res
	}
	app.logger.Infof("CheckTx: %s, NodeID: %s", method, nodeID)
	if method != "" && param != "" && nonce != nil && signature != nil && nodeID != "" {
		// Check has function in system
		if IsMethod[method] {
			result := app.CheckTxRouter(method, param, nonce, signature, nodeID)
			return result
		}
		res = app.ReturnCheckTx(code.UnknownMethod, "Unknown method name")
		return res
	}
	res = app.ReturnCheckTx(code.InvalidTransactionFormat, "Invalid transaction format")
	return res
}

func hash(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

func (app *Application) Commit() types.ResponseCommit {
	app.logger.Infof("Commit")
	app.SaveDBState()
	app.state.Height = app.state.Height + 1
	for key := range app.deliverTxNonceState {
		delete(app.checkTxNonceState, key)
	}
	app.deliverTxNonceState = make(map[string][]byte)

	// Calculate app hash
	if len(app.HashData) > 0 {
		app.HashData = append(app.state.AppHash, app.HashData...)
		app.state.AppHash = hash(app.HashData)
	}
	appHash := app.state.AppHash
	app.HashData = make([]byte, 0)

	// Save state
	saveState(app.state)
	return types.ResponseCommit{Data: appHash}
}

func (app *Application) Query(reqQuery types.RequestQuery) (res types.ResponseQuery) {
	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorf("Recovered in %s, %s", r, identifyPanic())
			res = app.ReturnQuery(nil, "Unknown error", app.state.Height)
		}
	}()
	var query protoTm.Query
	err := proto.Unmarshal(reqQuery.Data, &query)
	if err != nil {
		app.logger.Error(err.Error())
	}
	method := query.Method
	param := query.Params
	app.logger.Infof("Query: %s", method)
	height := reqQuery.Height
	if height == 0 {
		height = app.state.Height
	}
	if method != "" {
		return app.QueryRouter(method, param, height)
	}
	return app.ReturnQuery(nil, "method can't empty", app.state.Height)
}

func identifyPanic() string {
	var name, file string
	var line int
	var pc [16]uintptr
	n := runtime.Callers(3, pc[:])
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(pc)
		name = fn.Name()
		if !strings.HasPrefix(name, "runtime.") {
			break
		}
	}
	switch {
	case name != "":
		return fmt.Sprintf("%v:%v", name, line)
	case file != "":
		return fmt.Sprintf("%v:%v", file, line)
	}
	return fmt.Sprintf("pc:%x", pc)
}

func (app *Application) isDuplicateNonce(nonce []byte) bool {
	return app.HasStateDB(nonce)
}

