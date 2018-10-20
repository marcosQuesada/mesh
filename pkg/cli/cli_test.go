package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestBasicCliServer(t *testing.T) {
	srv := New(1234)

	// Register RPC Handlers
	arith := new(Arith)
	/*	srv.Register("divide", arith.Divide)
		srv.Register("multiply", arith.Multiply)*/
	srv.RegisterCommands(arith)
	go srv.Run()

	time.Sleep(time.Second * 1)
	//start client
	client, err := newFakeClient(1234)
	if err != nil {
		t.Error("Failing to connect server")
	}
	requestChan := make(chan string, 0)
	responseChan := make(chan []byte, 0)
	go func(reqChan chan string, resChan chan []byte) {
		for {
			select {
			case cmd, open := <-reqChan:
				if !open {
					return
				}

				buffer, err := client.do(cmd)
				if err != nil {
					t.Error("Error receiving divide result")
				}
				resChan <- buffer
			}
		}
	}(requestChan, responseChan)

	// INVOKE DIVIDE
	requestChan <- "divide 10 2"
	buffer := <-responseChan
	var v Quotient
	err = json.Unmarshal(buffer, &v)
	//Wait command result

	if v.Quo != 5 {
		t.Error("Unexpected quo")
	}

	if v.Rem != 0 {
		t.Error("Unexpected Rem")
	}

	//INVOKE MULTIPLY
	requestChan <- "multiply 10 2"
	prod := <-responseChan
	if string(prod) != "20\n" {
		t.Error("Unexpected Product Result", string(prod))
	}

	close(responseChan)
	close(requestChan)
}

type fakeClient struct {
	conn net.Conn
	done chan bool
}

func newFakeClient(port int) (*fakeClient, error) {
	client, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}
	done := make(chan bool, 0)

	return &fakeClient{client, done}, nil
}

func (f *fakeClient) do(cmd string) ([]byte, error) {
	f.conn.Write([]byte(fmt.Sprintf("%s\r\n", cmd)))
	b := bufio.NewReader(f.conn)

	return b.ReadBytes('\n')
}

//// Examples of implementation
type Quotient struct {
	Quo, Rem int
}

type Arith int

func (a *Arith) CliHandlers() map[string]Definition {
	return map[string]Definition{
		"divide":   Definition{a.Divide, "divide method example"},
		"multiply": Definition{a.Multiply, "multiply method example"},
	}
}

func (t *Arith) Multiply(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("Unexpected number of arguments ")
	}

	arg0, _ := strconv.ParseInt(args[0].(string), 10, 64)
	arg1, _ := strconv.ParseInt(args[1].(string), 10, 64)

	return int(arg0 * arg1), nil
}

func (t *Arith) Divide(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("Unexpected number of arguments ")
	}
	arg0, _ := strconv.ParseInt(args[0].(string), 10, 64)
	arg1, _ := strconv.ParseInt(args[1].(string), 10, 64)

	if arg1 == 0 {
		return nil, errors.New("divide by zero")
	}
	quo := &Quotient{}
	quo.Quo = int(arg0 / arg1)
	quo.Rem = int(arg0 % arg1)

	return quo, nil
}
