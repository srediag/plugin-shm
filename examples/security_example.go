// Example for secure plugin loading and validation using the API interface.
package examples

import (
	"fmt"

	"github.com/srediag/plugin-shm/api"
)

func ExamplePluginSecurity() {
	sec := &api.ExampleSecurity{}
	if err := sec.ValidateSignature("/path/to/plugin.so"); err != nil {
		fmt.Println("signature validation failed:", err)
	}
	if err := sec.SecureChannel("peer-host"); err != nil {
		fmt.Println("secure channel failed:", err)
	}
}
