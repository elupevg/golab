package docker

import (
	"context"
	"strconv"

	"github.com/elupevg/golab/topology"
)

type StubVP struct {
	linkCount int
	nodeCount int
	linkErr   error
	nodeErr   error
}

func (s *StubVP) LinkCreate(_ context.Context, _ topology.Link) (string, error) {
	if s.linkErr != nil {
		return "", s.linkErr
	}
	s.linkCount++
	return strconv.Itoa(s.linkCount), nil
}

func (s *StubVP) NodeCreate(_ context.Context, _ topology.Node) (string, error) {
	if s.nodeErr != nil {
		return "", s.nodeErr
	}
	s.nodeCount++
	return strconv.Itoa(s.nodeCount), nil
}

func New() *StubVP {
	return &StubVP{}
}

/*
import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

func main() {
	dc, err := client.NewClientWithOpts(client.WithVersion("1.47"))
	if err != nil {
		panic(err)
	}
	defer dc.Close()

	opts := network.CreateOptions{
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet:  "100.77.0.0/24",
					Gateway: "100.77.0.254",
				},
			},
		},
	}
	resp, err := dc.NetworkCreate(context.Background(), "ptpbygo", opts)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}
*/
