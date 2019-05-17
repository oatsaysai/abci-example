package app

import (
	appV1 "abci-example/abci/app/v1"
	"abci-example/utils"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var _ types.Application = (*ApplicationInterface)(nil)

type ApplicationInterface struct {
	appV1        *appV1.Application
	CurrentBlock int64
}

func NewApplicationInterface() *ApplicationInterface {
	logger := logrus.WithFields(logrus.Fields{"module": "abci-app"})
	var dbType = utils.GetEnv("ABCI_DB_TYPE", "goleveldb")
	var dbDir = utils.GetEnv("ABCI_DB_DIR_PATH", "./DB")
	if err := cmn.EnsureDir(dbDir, 0700); err != nil {
		panic(fmt.Errorf("Could not create DB directory: %v", err.Error()))
	}
	name := "appDB"
	db := dbm.NewDB(name, dbm.DBBackendType(dbType), dbDir)
	return &ApplicationInterface{
		appV1: appV1.NewApplication(logger, db),
	}
}

func (app *ApplicationInterface) Info(req types.RequestInfo) types.ResponseInfo {
	return app.appV1.Info(req)
}

func (app *ApplicationInterface) SetOption(req types.RequestSetOption) types.ResponseSetOption {
	return app.appV1.SetOption(req)
}

func (app *ApplicationInterface) CheckTx(tx []byte) types.ResponseCheckTx {
	switch {
	case app.CurrentBlock >= 0:
		return app.appV1.CheckTx(tx)
	default:
		return app.appV1.CheckTx(tx)
	}
}

func (app *ApplicationInterface) DeliverTx(tx []byte) types.ResponseDeliverTx {
	switch {
	case app.CurrentBlock >= 0:
		return app.appV1.DeliverTx(tx)
	default:
		return app.appV1.DeliverTx(tx)
	}
}

func (app *ApplicationInterface) Commit() types.ResponseCommit {
	return app.appV1.Commit()
}

func (app *ApplicationInterface) Query(reqQuery types.RequestQuery) types.ResponseQuery {
	return app.appV1.Query(reqQuery)
}

func (app *ApplicationInterface) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	return app.appV1.InitChain(req)
}

func (app *ApplicationInterface) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.CurrentBlock = req.Header.Height
	return app.appV1.BeginBlock(req)
}

func (app *ApplicationInterface) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return app.appV1.EndBlock(req)
}

