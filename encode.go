package redis

import "strconv"

// encode - function translates arguments to array of bulk string
func encode(args []string) []byte {
	quantity := strconv.Itoa(len(args))
	str := "*" + quantity + RN
	for _, k := range args {
		length := strconv.Itoa(len(k))
		str += "$" + length + RN + k + RN
	}
	return []byte(str)
}
