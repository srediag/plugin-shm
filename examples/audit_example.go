// Example for audit logging and governance using the API interface.
package examples

import (
	"fmt"

	"github.com/srediag/plugin-shm/api"
)

func ExamplePluginAudit() {
	audit := &api.ExampleAudit{}
	event := "plugin_loaded"
	details := map[string]interface{}{"plugin": "plugin-1", "user": "admin"}
	if err := audit.LogEvent(event, details); err != nil {
		fmt.Println("audit log failed:", err)
	}
}
