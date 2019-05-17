package main

import (
	"abci-example/test/common"
	"abci-example/utils"
	"testing"
)

func TestSetAndGetKeyValue(t *testing.T) {
	for i := 0; i < 10; i++ {
		key := utils.RandStringRunes(10)
		value := utils.RandStringRunes(20)
		common.TestSetKeyValue(t, key, value)
		common.TestGetKeyValue(t, key, value)
	}
}

