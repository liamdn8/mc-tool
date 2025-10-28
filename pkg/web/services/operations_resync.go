package services

import (
	"fmt"
	"sort"
)

// GetResyncOptions returns available replication resync options and cluster metadata.
func (os *OperationsService) GetResyncOptions() (map[string]interface{}, error) {
	aliases, err := os.minioService.GetAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to get aliases: %w", err)
	}

	aliasIndex, clusters, err := os.replicationService.GetClusterAliasIndex()
	if err != nil {
		return nil, err
	}

	// Build map of cluster identifiers to alias names for quick lookups.
	clusterAliasCounts := make(map[string]int)
	for _, info := range aliasIndex {
		clusterID := clusterIdentifier(info.Cluster)
		if clusterID == "" {
			continue
		}
		clusterAliasCounts[clusterID]++
	}

	aliasOptions := make([]map[string]interface{}, 0, len(aliases))
	for _, alias := range aliases {
		option := map[string]interface{}{
			"alias": alias.Name,
			"url":   alias.URL,
		}

		if info, ok := aliasIndex[alias.Name]; ok {
			clusterID := clusterIdentifier(info.Cluster)
			option["clusterId"] = clusterID
			option["clusterName"] = info.Cluster.Name
			option["deploymentId"] = info.Cluster.DeploymentID
			option["site"] = sanitizeSiteMetadata(info.Site)

			peerCount := clusterAliasCounts[clusterID]
			if peerCount > 0 {
				option["peerCount"] = peerCount - 1
			}

			if clusterAliasCounts[clusterID] > 1 {
				option["canResync"] = true
			} else {
				option["canResync"] = false
				option["reason"] = "Cluster has no peer targets for resync"
			}
		} else {
			option["canResync"] = false
			option["reason"] = "Alias is not part of a replication cluster"
		}

		aliasOptions = append(aliasOptions, option)
	}

	sort.Slice(aliasOptions, func(i, j int) bool {
		ai := aliasOptions[i]["alias"].(string)
		aj := aliasOptions[j]["alias"].(string)
		return ai < aj
	})

	clusterSummaries := buildClusterSummaries(clusters)

	return map[string]interface{}{
		"aliases":  aliasOptions,
		"clusters": clusterSummaries,
	}, nil
}

// StartReplicationResync validates selection and triggers the resync operation.
func (os *OperationsService) StartReplicationResync(sourceAlias, targetAlias string) (map[string]interface{}, error) {
	if sourceAlias == "" || targetAlias == "" {
		return nil, fmt.Errorf("sourceAlias and targetAlias are required")
	}

	if sourceAlias == targetAlias {
		return nil, fmt.Errorf("source and target alias must be different")
	}

	aliasIndex, _, err := os.replicationService.GetClusterAliasIndex()
	if err != nil {
		return nil, err
	}

	clusterAliasCounts := make(map[string]int)
	for _, info := range aliasIndex {
		clusterID := clusterIdentifier(info.Cluster)
		if clusterID == "" {
			continue
		}
		clusterAliasCounts[clusterID]++
	}

	sourceInfo, ok := aliasIndex[sourceAlias]
	if !ok {
		return nil, fmt.Errorf("source alias %s is not part of a replication cluster", sourceAlias)
	}

	targetInfo, ok := aliasIndex[targetAlias]
	if !ok {
		return nil, fmt.Errorf("target alias %s is not part of a replication cluster", targetAlias)
	}

	sourceClusterID := clusterIdentifier(sourceInfo.Cluster)
	targetClusterID := clusterIdentifier(targetInfo.Cluster)

	if sourceClusterID == "" || targetClusterID == "" || sourceClusterID != targetClusterID {
		return nil, fmt.Errorf("aliases must belong to the same replication cluster")
	}

	if clusterAliasCounts[sourceClusterID] <= 1 {
		return nil, fmt.Errorf("cluster %s does not have a peer target available for resync", sourceClusterID)
	}

	output, err := os.replicationService.StartReplicationResync(sourceAlias, targetAlias)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success":     true,
		"message":     fmt.Sprintf("Resync triggered from %s to %s", sourceAlias, targetAlias),
		"output":      output,
		"clusterId":   sourceClusterID,
		"sourceAlias": sourceAlias,
		"targetAlias": targetAlias,
	}, nil
}

// GetReplicationResyncStatus retrieves the current resync status for the selected aliases.
func (os *OperationsService) GetReplicationResyncStatus(sourceAlias, targetAlias string) (map[string]interface{}, error) {
	if sourceAlias == "" || targetAlias == "" {
		return nil, fmt.Errorf("sourceAlias and targetAlias are required")
	}

	status, err := os.replicationService.GetReplicationResyncStatus(sourceAlias, targetAlias)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"status":  status,
	}, nil
}

func clusterIdentifier(cluster *ClusterInfo) string {
	if cluster == nil {
		return ""
	}
	if cluster.DeploymentID != "" {
		return cluster.DeploymentID
	}
	return cluster.Name
}

func sanitizeSiteMetadata(site map[string]interface{}) map[string]interface{} {
	if site == nil {
		return nil
	}

	cleaned := map[string]interface{}{}
	if v, ok := site["name"].(string); ok && v != "" {
		cleaned["name"] = v
	}
	if v, ok := site["deploymentID"].(string); ok && v != "" {
		cleaned["deploymentId"] = v
	}
	if v, ok := site["endpoint"].(string); ok && v != "" {
		cleaned["endpoint"] = v
	}
	if v, ok := site["status"].(string); ok && v != "" {
		cleaned["status"] = v
	}
	if v, ok := site["state"].(string); ok && v != "" {
		cleaned["state"] = v
	}
	if v, ok := site["syncStatus"].(string); ok && v != "" {
		cleaned["syncStatus"] = v
	}

	return cleaned
}

func buildClusterSummaries(clusters []ClusterInfo) []map[string]interface{} {
	summaries := make([]map[string]interface{}, 0, len(clusters))

	for _, cluster := range clusters {
		clusterID := clusterIdentifier(&cluster)

		siteSummaries := make([]map[string]interface{}, 0, len(cluster.Sites))
		aliasNames := make([]string, 0, len(cluster.Sites))

		for _, site := range cluster.Sites {
			if site == nil {
				continue
			}
			name, _ := site["name"].(string)
			aliasNames = append(aliasNames, name)
			siteSummaries = append(siteSummaries, sanitizeSiteMetadata(site))
		}

		sort.Strings(aliasNames)

		summaries = append(summaries, map[string]interface{}{
			"clusterId":    clusterID,
			"name":         cluster.Name,
			"deploymentId": cluster.DeploymentID,
			"aliasCount":   len(aliasNames),
			"aliases":      aliasNames,
			"sites":        siteSummaries,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		ci := summaries[i]["clusterId"].(string)
		cj := summaries[j]["clusterId"].(string)
		return ci < cj
	})

	return summaries
}
