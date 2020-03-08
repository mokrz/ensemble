package node

import "github.com/containerd/containerd"

type Worker struct {
	Node
}

func NewWorker(cfg Config, ctr *containerd.Client) (w *Worker) {
	return &Worker{
		Node: Node{
			Cfg: cfg,
			Ctr: ctr,
		},
	}
}
