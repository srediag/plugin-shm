// Example for lifecycle management (start, stop, reload) using the API interface.
package examples

import (
	"fmt"

	"github.com/srediag/plugin-shm/api"
)

func ExamplePluginLifecycleManager() {
	lc := &api.ExampleLifecycle{}
	if err := lc.StartPlugin("plugin-1"); err != nil {
		fmt.Println("start failed:", err)
	}
	if err := lc.ReloadPlugin("plugin-1"); err != nil {
		fmt.Println("reload failed:", err)
	}
	if err := lc.StopPlugin("plugin-1"); err != nil {
		fmt.Println("stop failed:", err)
	}
}
