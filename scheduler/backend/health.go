package backend

type CellsHealth map[string]int

func (b Backends) InspectHealth() (CellsHealth, error) {
	if len(b) <= 0 {
		return CellsHealth{}, nil
	}
	clusterLoader := b[0].clusterLoader
	clusters, err := clusterLoader.LoadAllClusters()
	if err != nil {
		return nil, err
	}
	health := CellsHealth{}
	for _, availableCell := range b {
		health[availableCell.ID] = 0
	}
	for _, cluster := range clusters {
		for _, clusterNode := range cluster.Nodes {
			backendID := clusterNode.BackendID
			if _, ok := health[backendID]; ok {
				health[backendID] += 1
			}
		}
	}
	return health, nil
}
