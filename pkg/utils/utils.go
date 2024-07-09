package utils

import (
	"main/pkg/constants"
	"net/http"
	"strconv"
)

func BoolToFloat64(b bool) float64 {
	if b {
		return 1
	}

	return 0
}

func GetBlockHeightFromHeader(header http.Header) (int64, error) {
	valueStr := header.Get(constants.HeaderBlockHeight)
	if valueStr == "" {
		return 0, nil
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}
