// Example for transport usage (send/receive) using the API interface.
package examples

import (
	"fmt"

	"github.com/srediag/plugin-shm/api"
)

func ExamplePluginTransport() {
	tr := &api.ExampleTransport{}
	if err := tr.Send([]byte("ping")); err != nil {
		fmt.Println("send failed:", err)
	}
	resp, err := tr.Receive()
	if err != nil {
		fmt.Println("receive failed:", err)
	}
	fmt.Println("received:", resp)
}
