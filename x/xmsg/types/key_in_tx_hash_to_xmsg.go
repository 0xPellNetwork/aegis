package types

const (
	// InTxHashToXmsgKeyPrefix is the prefix to retrieve all InTxHashToXmsg
	InTxHashToXmsgKeyPrefix = "InTxHashToXmsg/value/"
)

// InTxHashToXmsgKey returns the store key to retrieve a InTxHashToXmsg from the index fields
func InTxHashToXmsgKey(
	inTxHash string,
) []byte {
	var key []byte

	inTxHashBytes := []byte(inTxHash)
	key = append(key, inTxHashBytes...)
	key = append(key, []byte("/")...)

	return key
}
