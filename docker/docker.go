package docker

import (
	"context"
	"fmt"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/elupevg/golab/topology"
)

type DockerProvider struct {
	dockerClient client.APIClient
	count        int
}

func New(dockerClient client.APIClient) *DockerProvider {
	return &DockerProvider{dockerClient: dockerClient}
}

func (dp *DockerProvider) LinkCreate(ctx context.Context, link topology.Link) (string, error) {
	_, ipv4Net, err := net.ParseCIDR(link.IPv4Subnet)
	_, lastIP := cidr.AddressRange(ipv4Net)
	lastIP = cidr.Dec(lastIP)
	opts := network.CreateOptions{
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet:  link.IPv4Subnet,
					Gateway: lastIP.String(),
				},
			},
		},
	}
	resp, err := dp.dockerClient.NetworkCreate(ctx, fmt.Sprintf("link%d", dp.count+1), opts)
	if err != nil {
		return "", err
	}
	dp.count++
	return resp.ID, nil
}

func (dp *DockerProvider) NodeCreate(_ context.Context, _ topology.Node) (string, error) {
	return "", nil
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
