package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/marcosQuesada/mesh/pkg/config"
	"github.com/marcosQuesada/mesh/pkg/server"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//Parse config
	cfgFile := flag.String("config", "", "Config file")
	logOut := flag.Bool("log", false, "logger out path")
	addr := flag.String("addr", "127.0.0.1:12000", "Port where Mesh is listen on")
	cluster := flag.String("cluster", "127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002", "cluster list definition separated by commas")
	cliPort := flag.Int("cli", 1234, "Port where Mesh is listen on")
	flag.Parse()

	//Define server from config
	var cfg *config.Config
	if *cfgFile != "" {
		cfg = config.ParseYML(*cfgFile)
	} else {
		cfg = config.NewConfig(*addr, *cluster, *cliPort)
	}

	//Init logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	f, err := os.OpenFile(fmt.Sprintf("log/%s.log", *addr), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Panicf("error opening file: %v", err)
	}
	defer f.Close()

	if *logOut {
		log.SetOutput(f)
	}
	s := server.New(cfg)
	c := make(chan os.Signal, 1)

	signal.Notify(
		c,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	//serve until signal
	go func() {
		<-c
		s.Close()
	}()

	//Server Run
	s.Start()
}
