package main

import (
	"flag"
	"github.com/marcosQuesada/mesh/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:11000", "Port where Mesh is listen on")
	raftAddr := flag.String("raft_addr", "127.0.0.1:12000", "raft listening address")
	raftCluster := flag.String("raft_cluster", "127.0.0.1:12000", "cluster list definition separated by commas")
	raftDataDir := flag.String("raft_data_dir", "./data/var0", "raft data store path")

	flag.Parse()

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

	go func() {
		<-c
		s.Close()
	}()

	//Server Run
	s.Run()
}
