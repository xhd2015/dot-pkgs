//go:build !darwin

package chrome

func snapStep(opts LoadUnpackedOpts, res *LoadUnpackedResult, step string, n *int) {
	// no-op off darwin
}
