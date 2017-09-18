package redis

import (
	"bytes"
	"errors"
	"strconv"
)

type redisType byte

func ligthParser(data []byte) ([]byte, error) {
	switch data[0] {
	case '+':
		return data[1 : len(data)-2], nil
	case '-':
		return nil, errors.New(string(data[1 : len(data)-2]))
	case ':':
		return data[1 : len(data)-2], nil
	default:
		return nil, errors.New("First byte error")
	}
}

//bulkParser return Byte Bulk string? and offset buffer
//return "", -1, if len(buffer) < len bulk
//return "",0 = atoi convert error
func bulkParser(buffer []byte) (string, int) {
	if buffer[1] == '0' {
		return "Empty Bulk", 4
		//$-1\r\n == Nill Bulk
	} else if comp := bytes.Compare(buffer[1:3], []byte("-1")); comp == 0 {
		return "Nill Bulk", 5
	} else {
		//check !!! len buffer
		str := bytes.SplitAfterN(buffer, []byte(RN), 2)
		lenth := len(str[0])
		quantity, err := strconv.Atoi(string(str[0][1 : lenth-2]))
		if err != nil {
			return "", 0
		}
		if len(buffer) < lenth+quantity+2 {
			return "", -1
		}
		return string(buffer[lenth : lenth+quantity]), lenth + quantity + 2
	}
}

//arrayParser - return data string response, errs, offset
func arrayParser(buffer []byte) (data string, errs []error, offset int) {
	if buffer[1] == '0' {
		return "Empty Array", nil, 4
	} else if comp := bytes.Compare(buffer[1:3], []byte("-1")); comp == 0 {
		return "Nill Array", nil, 5
	} else {
		str := bytes.SplitAfterN(buffer, []byte(RN), 2)
		lenth := len(str[0])
		offset := lenth
		buffer = buffer[lenth:]
		quantity, err := strconv.Atoi(string(str[0][1 : lenth-2]))
		if err != nil {
			errs = append(errs, err)
			return "", errs, 0
		}

		for i := 0; i < quantity; i++ {
			key := buffer[0]
			if len(buffer) == 0 {
				return "", nil, -1
			}
			if key == '*' {
				resp, errors, n := arrayParser(buffer)
				if errors != nil {
					errs = append(errs, errors...)
				}
				offset += n
				if len(data) != 0 {
					data += "|" + resp
				} else {
					data += resp
				}
				buffer = buffer[n:]

			} else if key == '$' {
				resp, n := bulkParser(buffer)
				if n == 0 {
					errs = append(errs, errors.New("Atoi conv error"))
					return "", errs, 0
				} else if n == -1 {
					return "", nil, -1
				}
				offset += n
				if len(data) != 0 {
					data += "|" + string(resp)

				} else {
					data += resp
				}
				buffer = buffer[n:]

			} else {
				str := bytes.SplitAfterN(buffer, []byte(RN), 2)
				lenth := len(str[0])
				buffer = buffer[lenth:]
				offset += lenth
				resp, err := ligthParser(str[0])
				if err != nil {
					errs = append(errs, err)
				} else {
					if len(data) != 0 {
						data += "|" + string(resp)
					} else {
						data += string(resp)
					}
				}
			}
		}
		return data, errs, offset
	}
}
