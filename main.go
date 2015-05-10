package main

import (
	"flag"
	"github.com/marcosQuesada/mesh/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	//Parse config
	addr := flag.String("addr", "127.0.0.1:11000", "Port where Mesh is listen on")
	raftAddr := flag.String("raft_addr", "127.0.0.1:12000", "cluster node listening address")
	raftCluster := flag.String("raft_cluster", "127.0.0.1:12000,127.0.0.2:12001,127.0.0.3:12002", "cluster list definition separated by commas")
	raftDataDir := flag.String("raft_data_dir", "./data/var0", "raft data store path")
	logFile := flag.String("logFile", "./log/log0.log", "log file")
	flag.Parse()

	//Init logger
	f := handleFile(*logFile)
	defer f.Close()
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	//Create COnfiguration
	config := server.NewConfig(*addr, *raftAddr, *raftCluster, *raftDataDir)
	//Define serfer from config
	s := server.New(config)
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
	s.Run()
}

func handleFile(logFile string) (f *os.File) {
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Panic("error opening file: %v", err)
	}

	return
}
