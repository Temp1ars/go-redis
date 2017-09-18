package redis

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	client := Start("localhost:6379")
	assert := assert.New(t)
	err := client.Connect()
	assert.NoError(err)

	t.Run("test request,response", func(t *testing.T) {
		err := client.Request("SET", "test", "ICE baby")
		if !assert.NoError(err) {
			return
		}
		err = client.PipeliningRequest([]string{"SET", "pipeling", "10"}, []string{"INCR", "pipeling"}, []string{"INCR", "pipeling"})
		if !assert.NoError(err) {
			return
		}

		err = client.Request("GET", "pipeling")
		if !assert.NoError(err) {
			return
		}
		err = client.PipeliningRequest([]string{"SET", "testErr", "NEW"}, []string{"GET", "testErr"}, []string{"GET", "test"})
		if !assert.NoError(err) {
			return
		}
		data, errs := client.Response()
		if data == nil && err != nil {
			assert.Error(errs[0])
			return
		}
		if errs != nil {
			for _, k := range errs {
				fmt.Println(k)
			}
		}
		for _, k := range data {
			fmt.Println("data:", k)
		}

		// resp, errs, n := arrayParser([]byte("*3\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Foo\r\n-Bar\r\n*2\r\n$4\r\nWoop\r\n$5\r\nWoop2\r\n"))
		// if errs != nil {
		// 	for _, k := range errs {
		// 		fmt.Println("errs:", k)
		// 	}
		// }
		// fmt.Println("Resp:", resp, "n:", n)

		// resp, errs, n = arrayParser([]byte("*3\r\n*0\r\n*-1\r\n*2\r\n$0\r\n$-1\r\n"))
		// if errs != nil {
		// 	for _, k := range errs {
		// 		fmt.Println("errs:", k)
		// 	}
		// }
		// fmt.Println("Resp:", resp, "n:", n)
	})
	t.Run("test Set,Get,MSet,MGet", func(t *testing.T) {
		err := client.Set("dyno", "Tyrannosaurus")
		if !assert.NoError(err) {
			return
		}
		resp, errs := client.Get("dyno")
		if err != nil {
			for _, k := range errs {
				fmt.Println(k)
			}
		}
		fmt.Println("dyno:", resp)
		err = client.MSet([]string{"dyno1", "Dilophosaurus", "dyno2"})
		_ = !assert.Equal(err.Error(), `Parity Error: "lacks key or value"`)
		err = client.MSet([]string{"dyno1", "Dilophosaurus", "dyno2", "Ceratopsia", "total dyno", "3"})
		if !assert.NoError(err) {
			return
		}
		data, errs := client.MGet([]string{"dyno1", "dyno2", "total dyno"})
		if errs != nil {
			for _, k := range errs {
				fmt.Println("ERROR:", k)
			}
		}
		for _, k := range data {
			fmt.Println("DATA:", k)
		}
	})
	t.Run("test error", func(t *testing.T) {
		err := client.Request("INCR", "test")
		if !assert.NoError(err) {
			return
		}
		_, errs := client.Response()
		_ = !assert.Equal(errs[0].Error(), "ERR value is not an integer or out of range")
	})
	err = client.CloseConnection()
	assert.NoError(err)
}
