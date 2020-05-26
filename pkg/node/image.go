package node

import (
	"github.com/containerd/containerd"
)

// Image wraps containerd.Image
type Image interface {
	Name() string
}

func newImage(i containerd.Image) Image {
	return &image{
		ctrImage: i,
	}
}

type image struct {
	ctrImage containerd.Image
}

func (i *image) Name() string {
	return i.ctrImage.Name()
}
