package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Comcast/go-leaderelection"
	"github.com/go-zookeeper/zk"
)

const electionNode = "/election/defy-fsrobot"

var node string

var zkaddr string

func init() {
	flag.StringVar(&node, "node", os.Getenv("HOSTNAME"), "node name of the cluster")
	flag.StringVar(&zkaddr, "zk", "localhost:2181", "zookeeper server")
}

func main() {
	flag.Parse()

	zkconn, _, err := zk.Connect([]string{zkaddr}, 10*time.Second)
	if err != nil {
		log.Fatal("zk.COnnect failed:", err)
	}
	defer zkconn.Close()

	v, _, err := zkconn.Get(electionNode)
	if err != nil {
		fmt.Printf("zk: get %s failed: %v\n", electionNode, err)
	}
	fmt.Printf("zk: get %s: %v\n", electionNode, string(v))

	if len(v) == 0 {
		s, err := zkconn.Create(electionNode, []byte(node), 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			log.Fatal("zk.create failed:", err)
		}

		fmt.Println("create data:", s)
	}

	election, err := leaderelection.NewElection(zkconn, electionNode, node)
	if err != nil {
		log.Fatal("leaderelection.NewElection failed:", err)
	}
	defer election.Resign()

	go election.ElectLeader()

	for {
		select {
		case status := <-election.Status():
			log.Println("electleader status:", status)
		case <-time.After(10 * time.Second):
			log.Println("Tick")
		}
	}
}
