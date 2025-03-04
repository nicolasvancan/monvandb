package executor

import (
	"context"
	"fmt"
	"sync"

	contexts "github.com/nicolasvancan/monvandb/src/dbvengine/contexts"
)

type ExecutionLayer struct {
	Nodes        map[string]*ExecutionNode
	Context      *contexts.ExecutionContext
	LayerChannel chan OpNotification
	WaitGroup    sync.WaitGroup
}

func (el *ExecutionLayer) Start() error {
	ctx, cancel := context.WithCancel(context.Background())

	// Start the execution layer
	for _, node := range el.Nodes {
		if len(node.DependsOn) == 0 {
			// Start the node
			el.WaitGroup.Add(1)
			go node.Run(ctx, &el.WaitGroup, el.LayerChannel)
		}
	}

	// Function to run until all nodes are finished
	go func() {
		for notification := range el.LayerChannel {
			if notification.Err != nil {
				fmt.Printf("Error on node %s: %v\n", notification.NodeId, notification.Err)
				cancel()
				return
			}

			fmt.Printf("node %s has Finished\n", notification.NodeId)
			for _, node := range el.Nodes {
				fmt.Printf("Checking node %s\n", node.Id)
				node.OnNotified(notification.NodeId)

				if len(node.DependsOn) == 0 && node.State == NodeIdle {
					el.WaitGroup.Add(1)
					go node.Run(ctx, &el.WaitGroup, el.LayerChannel)
				}
			}
			el.WaitGroup.Done()
		}
	}()

	el.WaitGroup.Wait()
	close(el.LayerChannel)
	fmt.Println("Execution Layer Finished")
	return nil
}

func (el *ExecutionLayer) AddNode(node *ExecutionNode, final bool) {
	node.Layer = el
	// Tell the node to insert its result on the context
	node.Final = final
	el.Nodes[node.Id] = node
}

func NewExecutionLayer(queryContext *contexts.ExecutionContext) *ExecutionLayer {
	el := new(ExecutionLayer)
	el.Nodes = make(map[string]*ExecutionNode)
	el.Context = queryContext
	el.LayerChannel = make(chan OpNotification)
	el.WaitGroup = sync.WaitGroup{}
	return el
}
