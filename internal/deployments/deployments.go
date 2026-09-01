// Package deployments lists known Live! by BULA installations to archive, as documented
// in the sister repo live-by-bula-openapi's README known-deployments table.
package deployments

// Deployment identifies one known Live! by BULA installation to archive.
type Deployment struct {
	Slug        string // our own short identifier, used for the archive filename
	Host        string // e.g. "wbuc.wfdf.sport"
	BasePath    string // e.g. "/live/data/"; heartbeat and every other endpoint hang off this
	SpecVersion string // "1.9.14" | "1.9.17" | "3.0.6" -- which spec shape this deployment produces
}

// WBUC2025 is the 1.9.14 reference deployment used to build and verify internal/bula/v1914.
var WBUC2025 = Deployment{
	Slug:        "wbuc2025",
	Host:        "wbuc.wfdf.sport",
	BasePath:    "/live/data/",
	SpecVersion: "1.9.14",
}
