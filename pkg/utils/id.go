package utils

import "strconv"

func Uint64ToString(id uint64) string {
	return strconv.FormatUint(id, 10)
}

func StringToUint64(idStr string) uint64 {
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		panic("invalid uint64 string: " + idStr)
	}
	return id
}

func StringToUint64Slice(iList []string) []uint64 {
	result := make([]uint64, len(iList))
	for i, idStr := range iList {
		result[i] = StringToUint64(idStr)
	}
	return result
}
