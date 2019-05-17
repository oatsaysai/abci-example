package code

// Return codes for return result
const (
	OK                       uint32 = 0
	EncodingError            uint32 = 1
	DecodingError            uint32 = 2
	BadNonce                 uint32 = 3
	Unauthorized             uint32 = 4
	UnmarshalError           uint32 = 5
	MarshalError             uint32 = 6
	DuplicateNonce           uint32 = 7
	MethodCanNotBeEmpty      uint32 = 8
	UnknownMethod            uint32 = 9
	InvalidTransactionFormat uint32 = 10
	UnknownError             uint32 = 999
)
