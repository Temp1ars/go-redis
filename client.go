package redis

import (
	"bytes"
	"errors"
	"net"
	"sync"
)

// RN  \r\n standart ending for redis
const RN string = "\r\n"

// Client struct client
type Client struct {
	addres string
	conn   net.Conn

	//amount = pending response
	amount int
	mu     sync.Mutex
}

// Start return client struct
func Start(addres string) *Client {
	return &Client{addres: addres}
}

// Connect connect net.Dial
func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.addres)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// CloseConnection closed net.Diadl connect
func (c *Client) CloseConnection() error {
	err := c.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// Request - dispatch request array of bulk string to server redis
func (c *Client) Request(args ...string) error {
	body := encode(args)
	_, err := c.conn.Write(body)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.amount++
	c.mu.Unlock()
	return nil
}

// PipeliningRequest - dispatch pipelining request array of bulk string to server redis
func (c *Client) PipeliningRequest(args ...[]string) error {
	body := []byte("")
	sum := 0
	for _, k := range args {
		body = AppendByte(body, encode(k)...)
		sum++
	}
	c.mu.Lock()
	c.amount += sum
	c.mu.Unlock()
	_, err := c.conn.Write(body)
	if err != nil {
		return err
	}
	return nil
}

// Response - receiving a response from the server redis
func (c *Client) Response() (data []string, errs []error) {
	buffer := make([]byte, 0, 5242880)
	c.mu.Lock()
	amount := c.amount
	c.mu.Unlock()
	//change - repeat x2
	tmp, err := read(c)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}
	buffer = AppendByte(buffer, tmp...)
	i := 0
	trigger := false
	for i < amount {
		if buffer[len(buffer)-1] != '\n' || trigger {
			tmp, err := read(c)
			if err != nil {
				errs = append(errs, err)
				return nil, errs
			}
			buffer = AppendByte(buffer, tmp...)
			trigger = false
		}

		key := buffer[0]
		if key == '*' {
			str, err, n := arrayParser(buffer)
			if n > 0 {
				errs = append(errs, err...)
				data = append(data, str)
				buffer = buffer[n:]
				i++
			} else {
				if n == 0 {
					errs = append(errs, err...)
					return nil, errs
				} else if n == -1 {
					trigger = true
				}
			}

		} else if key == '$' {
			resp, n := bulkParser(buffer)
			if n > 0 {
				data = append(data, resp)
				buffer = buffer[n:]
				i++
			} else {
				if n == 0 {
					errs = append(errs, errors.New("Atoi conv error"))
					return nil, errs
				} else if n == -1 {
					trigger = true
				}
			}
		} else {
			str := bytes.SplitAfterN(buffer, []byte(RN), 2)
			buffer = buffer[len(str[0]):]
			resp, err := ligthParser(str[0])
			if err != nil {
				errs = append(errs, err)
			}
			data = append(data, string(resp))
			i++
		}
	}
	c.mu.Lock()
	c.amount = 0
	c.mu.Unlock()
	// change
	return
}

func read(c *Client) ([]byte, error) {
	tmp := make([]byte, 4096)
	num, err := c.conn.Read(tmp)
	if err != nil {
		return nil, err
	}
	return tmp[:num], nil
}

// AppendByte -function copy byte slice[] src->dst
func AppendByte(slice []byte, data ...byte) []byte {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([]byte, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

//Set redis: SET key value
//return err, when request error, or response(redis:error oerations)
//return nil= ok.
func (c *Client) Set(key, value string) error {
	err := c.Request("SET", key, value)
	if err != nil {
		return err
	}
	_, errs := c.Response()
	if errs != nil {
		return errs[0]
	}
	return nil
}

//MSet redis: MSET key value
//sample value[0]=redis key,value[1]=redis value
func (c *Client) MSet(value []string) error {
	lenth := len(value)
	if lenth%2 != 0 {
		return errors.New(`Parity Error: "lacks key or value"`)
	}
	req := make([][]string, lenth/2)
	for i, j := 0, 0; i < lenth; i, j = i+2, j+1 {
		req[j] = append(req[j], "SET")
		req[j] = append(req[j], value[i])
		req[j] = append(req[j], value[i+1])
	}
	err := c.PipeliningRequest(req...)
	if err != nil {
		return err
	}
	return nil

}

// Get return result
func (c *Client) Get(value string) (resp string, errs []error) {
	c.mu.Lock()
	amt := c.amount
	c.mu.Unlock()
	err := c.Request("GET", value)
	if err != nil {
		errs = append(errs, err)
		return "", errs
	}
	data, errs := c.Response()
	if amt == 0 {
		resp = data[0]
		return
	}
	resp = data[amt]
	return

}

// MGet return result
func (c *Client) MGet(value []string) (resp []string, errs []error) {
	req := make([][]string, len(value))
	for i, k := range value {
		req[i] = append(req[i], "GET")
		req[i] = append(req[i], k)
	}
	c.mu.Lock()
	amt := c.amount
	c.mu.Unlock()
	err := c.PipeliningRequest(req...)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}
	resp, errs = c.Response()
	if amt == 0 {
		return
	}
	resp = resp[amt:]
	return
}
