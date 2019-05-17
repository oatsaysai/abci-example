package app

type SetValidatorParam struct {
	PublicKey string `json:"public_key"`
	Power     int64  `json:"power"`
}

type SetKeyValueParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GetKeyValueParam struct {
	Key string `json:"key"`
}

type GetKeyValueResult struct {
	Value string `json:"value"`
}

