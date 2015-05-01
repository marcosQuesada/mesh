package mesh

import (
	"flag"
	"fmt"
)

func main() {
	serverPort := flag.Int("port", 5000, "socket server port")
	mode := flag.String("mode", "alone", "node initialization mode")
	remote := flag.String("remote", "", "remote server address to join as 127.0.0.1:8081")
	flag.Parse()

	fmt.Println("mode:", *mode)
	fmt.Println("remote:", *remote)

	go StartServer(*serverPort)

	if *remote != "" {
		go StartClient("1", *remote)
	}

	fmt.Println("here it goes")
}
