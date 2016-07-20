package cells

type CellsHealth map[string]int

func (cells Cells) InspectHealth() (CellsHealth, error) {
	if len(cells) <= 0 {
		return CellsHealth{}, nil
	}
	clusterLoader := cells[0].clusterLoader
	clusters, err := clusterLoader.LoadAllRunningClusters()
	if err != nil {
		return nil, err
	}
	health := CellsHealth{}
	for _, availableCell := range cells {
		health[availableCell.GUID] = 0
	}
	for _, cluster := range clusters {
		for _, clusterNode := range cluster.Nodes {
			cellID := clusterNode.CellGUID
			if _, ok := health[cellID]; ok {
				health[cellID] += 1
			}
		}
	}
	return health, nil
}
