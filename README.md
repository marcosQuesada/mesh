Mesh [WIP]
==========

 Proof of concept on how to build a basic cluster structure using Raft

Milestones
===========
-Basic Raft cluster
-Command Quorum execution

Cluster definition will be something like that
==============================================

mesh -addr=127.0.0.1:11000 -raft_addr=127.0.0.1:12000 -raft_data_dir=./data/var0 -raft_cluster=127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002

mesh -addr=127.0.0.1:11001 -raft_addr=127.0.0.1:12001 -raft_data_dir=./data/var1 -raft_cluster=127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002

mesh -addr=127.0.0.1:11002 -raft_addr=127.0.0.1:12002 -raft_data_dir=./data/var2 -raft_cluster=127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002

Where:
	addr: Port where Mesh is listen on
	raft_addr: raft listening address
	raft_data_dir: raft data store path
	raft_cluster: cluster list definition separated by commas
