// Example for using the shared memory buffer in plugin-shm.
package examples

import (
	"context"
	"fmt"

	"github.com/srediag/plugin-shm/pkg/shm"
)

func ExampleSharedMemoryBuffer() {
	ctx := context.Background()
	buf, err := shm.Open(ctx, shm.OpenOptions{Name: "example", Size: 4096, Create: true})
	if err != nil {
		fmt.Println("failed to open buffer:", err)
		return
	}
	defer buf.Close()
	// Write data
	_, _ = buf.Write(ctx, []byte("hello world"))
	// Read data
	out := make([]byte, 64)
	_, _ = buf.Read(ctx, out)
	fmt.Println(string(out))
}
