package ebmlib

// RepairOptions configures repair behavior for EPUB/PDF.
type RepairOptions struct {
	// Aggressive enables destructive, best-effort repairs that may alter structure.
	Aggressive bool
}
