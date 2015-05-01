package test

import (
	"fmt"
	"github.com/marcosQuesada/mesh"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	go mesh.StartServer(5000) /*
		go mesh.StartServer(5001)
		go mesh.StartServer(5002)*/
	time.Sleep(1 * time.Second)
	go mesh.StartClient("1", "127.0.0.1:5000")
	go mesh.StartClient("2", "127.0.0.1:5000")
	go mesh.StartClient("3", "127.0.0.1:5000")

	fmt.Println("here it goes")
	time.Sleep(3 * time.Second)
	t.Fail()
}
