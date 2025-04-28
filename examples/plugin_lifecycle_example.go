// Example for plugin lifecycle management using the API interface.
package examples

import (
	"fmt"

	"github.com/srediag/plugin-shm/api"
)

func ExamplePluginLifecycle() {
	plugin := &api.ExamplePlugin{}
	if err := plugin.Start(); err != nil {
		fmt.Println("start failed:", err)
	}
	if err := plugin.Reload(); err != nil {
		fmt.Println("reload failed:", err)
	}
	if err := plugin.Stop(); err != nil {
		fmt.Println("stop failed:", err)
	}
}
