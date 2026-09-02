// Package livepublish renders a 3.0.6 archive as static JSON files in docs/
// not yet implemented
package livepublish

import (
	"context"
	"fmt"
)

// Publish is not yet implemented for 3.0.6 archives
func Publish(ctx context.Context, dbPath, outDir string) error {
	return fmt.Errorf("publish: not yet implemented for v03_00_06")
}
