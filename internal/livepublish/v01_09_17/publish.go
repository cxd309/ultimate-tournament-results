// Package livepublish renders a 1.9.17 archive as static JSON files in docs/
// not yet implemented
package livepublish

import (
	"context"
	"fmt"
)

// Publish is not yet implemented for 1.9.17 archives
func Publish(ctx context.Context, dbPath, outDir string) error {
	return fmt.Errorf("publish: not yet implemented for v01_09_17")
}
