package chains

import "errors"

// ReceiveStatusFromString returns a ReceiveStatus from a string using in CLI
// 0 for success, 1 for failed
// TODO: remove "receive" naming ans use outbound
func ReceiveStatusFromString(str string) (ReceiveStatus, error) {
	switch str {
	case "0":
		return ReceiveStatus_CREATED, nil
	case "1":
		return ReceiveStatus_SUCCESS, nil
	case "2":
		return ReceiveStatus_FAILED, nil
	default:
		return ReceiveStatus(0), errors.New("wrong status, must be 0 for success or 1 for failed")
	}
}
