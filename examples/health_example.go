// Example for health/liveness monitoring using the API interface.
package examples

import (
	"fmt"

	"github.com/srediag/plugin-shm/api"
)

func ExamplePluginHealth() {
	health := &api.ExampleHealth{}
	if err := health.Heartbeat("plugin-1"); err != nil {
		fmt.Println("heartbeat failed:", err)
	}
	alive, err := health.LivenessCheck("plugin-1")
	if err != nil {
		fmt.Println("liveness check failed:", err)
	}
	fmt.Println("plugin alive:", alive)
}
