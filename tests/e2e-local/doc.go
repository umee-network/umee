// package e2e defines an integration testing suite used for full end-to-end
// testing functionality.
//
// The file e2e_suite_test.go defines the testing suite and contains the core
// bootrapping logic that creates a testing environment via Docker containers.
// A testing network is created dynamically and contains multiple Docker
// containers:
//
// 1. A single validator Gaia network
// 2. A configurable number of Umee validator processes
// 3. A hermes relayer connecting the Umee and Gaia networks over IBC
// 4. A single Ethereum node
// 5. A configurable number of Peggy orchestrator processes
//
// The file e2e_test.go contains the actual end-to-end integration tests that
// utilize the testing suite.
package e2e
