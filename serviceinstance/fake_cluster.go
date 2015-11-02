package serviceinstance

// FakeCluster models a fake cluster of nodes
type FakeCluster struct {
	Replicas int
	Size     string
}

// NewFakeCluster creates a fake Cluster
func NewFakeCluster(replicas int, size string) FakeCluster {
	return FakeCluster{Replicas: replicas, Size: size}
}
