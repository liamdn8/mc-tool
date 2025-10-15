package services

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/liamdn8/mc-tool/pkg/logger"
)

// ReplicationService handles site replication operations
type ReplicationService struct {
	minioService *MinIOService
}

// NewReplicationService creates a new replication service
func NewReplicationService(minioService *MinIOService) *ReplicationService {
	return &ReplicationService{
		minioService: minioService,
	}
}

// ReplicationInfo represents replication information for a site
type ReplicationInfo struct {
	Enabled      bool                     `json:"enabled"`
	Sites        []map[string]interface{} `json:"sites"`
	DeploymentID string                   `json:"deployment_id"`
}

// GetReplicationInfo retrieves replication information for all aliases
func (rs *ReplicationService) GetReplicationInfo() (map[string]interface{}, error) {
	aliases, err := rs.minioService.GetAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to get aliases: %v", err)
	}

	if len(aliases) == 0 {
		return map[string]interface{}{
			"enabled": false,
			"sites":   []interface{}{},
			"message": "No MinIO aliases configured",
		}, nil
	}

	var sites []map[string]interface{}
	var replicationEnabled = false
	var replicationGroupInfo map[string]interface{}

	for _, alias := range aliases {
		siteInfo := map[string]interface{}{
			"name":               alias.Name,
			"url":                alias.URL,
			"healthy":            false,
			"replicationEnabled": false,
			"replicationStatus":  "not_configured",
			"deploymentID":       "",
			"siteName":           "",
		}

		// Check replication status
		replicateCmd := exec.Command("mc", "admin", "replicate", "info", alias.Name, "--json")
		replicateOutput, replicateErr := replicateCmd.CombinedOutput()

		if replicateErr == nil {
			var replicateInfo map[string]interface{}
			if json.Unmarshal(replicateOutput, &replicateInfo) == nil {
				if enabled, ok := replicateInfo["enabled"].(bool); ok && enabled {
					siteInfo["replicationEnabled"] = true
					siteInfo["replicationStatus"] = "configured"
					replicationEnabled = true

					if replicationGroupInfo == nil {
						replicationGroupInfo = replicateInfo
					}

					// Extract site information from peer sites
					if sitesList, ok := replicateInfo["sites"].([]interface{}); ok {
						for _, peerSite := range sitesList {
							if peer, ok := peerSite.(map[string]interface{}); ok {
								peerEndpoint, _ := peer["endpoint"].(string)
								peerName, _ := peer["name"].(string)

								matched := false
								if peerEndpoint == alias.URL {
									matched = true
								}
								if !matched && peerEndpoint != "" && alias.URL != "" {
									if strings.Contains(alias.URL, peerEndpoint) || strings.Contains(peerEndpoint, alias.URL) {
										matched = true
									}
								}
								if !matched && peerName == alias.Name {
									matched = true
								}

								if matched {
									if name, ok := peer["name"].(string); ok && name != "" {
										siteInfo["siteName"] = name
									}
									if deployID, ok := peer["deploymentID"].(string); ok && deployID != "" {
										siteInfo["deploymentID"] = deployID
									}
									break
								}
							}
						}
					}
				} else {
					siteInfo["replicationStatus"] = "disabled"
				}
			}
		} else {
			siteInfo["replicationStatus"] = "not_configured"
		}

		// Get health status
		healthInfo := rs.minioService.GetAliasHealth(alias.Name)
		siteInfo["healthy"] = healthInfo["healthy"]

		sites = append(sites, siteInfo)
	}

	return map[string]interface{}{
		"enabled":          replicationEnabled,
		"sites":            sites,
		"totalAliases":     len(sites),
		"replicationGroup": replicationGroupInfo,
		"configuredSites":  len(sites),
	}, nil
}

// ClusterInfo represents information about a replication cluster
type ClusterInfo struct {
	Name         string                   `json:"name"`
	Sites        []map[string]interface{} `json:"sites"`
	DeploymentID string                   `json:"deployment_id"`
	Enabled      bool                     `json:"enabled"`
}

// DetectReplicationClusters detects existing replication clusters and potential split brain scenarios
func (rs *ReplicationService) DetectReplicationClusters() ([]ClusterInfo, error) {
	aliases, err := rs.minioService.GetAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to get aliases: %v", err)
	}

	var clusters []ClusterInfo
	processedAliases := make(map[string]bool)

	for _, alias := range aliases {
		if processedAliases[alias.Name] {
			continue
		}

		// Check if this alias has replication enabled
		replicateCmd := exec.Command("mc", "admin", "replicate", "info", alias.Name, "--json")
		replicateOutput, err := replicateCmd.CombinedOutput()
		if err != nil {
			continue
		}

		var replicateInfo map[string]interface{}
		if json.Unmarshal(replicateOutput, &replicateInfo) != nil {
			continue
		}

		if enabled, ok := replicateInfo["enabled"].(bool); !ok || !enabled {
			continue
		}

		// Extract cluster information
		cluster := ClusterInfo{
			Name:    alias.Name,
			Enabled: true, // If we reach here, replication is enabled
		}

		if sites, ok := replicateInfo["sites"].([]interface{}); ok {
			for _, site := range sites {
				if siteMap, ok := site.(map[string]interface{}); ok {
					cluster.Sites = append(cluster.Sites, siteMap)

					// Mark site names as processed to avoid duplicates
					if siteName, ok := siteMap["name"].(string); ok {
						processedAliases[siteName] = true
					}
				}
			}
		}

		if deploymentID, ok := replicateInfo["deploymentID"].(string); ok {
			cluster.DeploymentID = deploymentID
		}

		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

// CheckSplitBrainStatus checks for split brain scenarios and provides detailed warnings
func (rs *ReplicationService) CheckSplitBrainStatus() (map[string]interface{}, error) {
	existingClusters, err := rs.DetectReplicationClusters()
	if err != nil {
		return nil, fmt.Errorf("failed to detect clusters: %v", err)
	}

	result := map[string]interface{}{
		"splitBrainDetected": len(existingClusters) > 1,
		"clusterCount":       len(existingClusters),
		"clusters":           existingClusters,
		"warnings":           []string{},
		"recommendations":    []string{},
	}

	if len(existingClusters) > 1 {
		logger.GetLogger().Warn("Split brain scenario detected", map[string]interface{}{
			"clusterCount": len(existingClusters),
			"clusters":     existingClusters,
		})

		warnings := []string{
			fmt.Sprintf("⚠️ SPLIT BRAIN DETECTED: %d separate replication clusters found", len(existingClusters)),
			"🔥 This configuration can lead to data inconsistency and conflicts",
			"💥 New site additions will be blocked until resolved",
		}

		recommendations := []string{
			"1. 🛠️ Choose one cluster as the primary and remove others",
			"2. 📋 Backup data from all clusters before making changes",
			"3. 🔄 Re-establish single cluster using mc admin replicate add",
			"4. ✅ Verify data consistency after resolution",
		}

		// Add detailed cluster information
		for i, cluster := range existingClusters {
			sites := []string{}
			for _, site := range cluster.Sites {
				if name, ok := site["name"].(string); ok {
					sites = append(sites, name)
				}
			}
			warnings = append(warnings, fmt.Sprintf("   Cluster %d: %s", i+1, strings.Join(sites, ", ")))
		}

		result["warnings"] = warnings
		result["recommendations"] = recommendations
		result["severity"] = "critical"
		result["actionRequired"] = true
	} else if len(existingClusters) == 1 {
		result["status"] = "healthy"
		result["message"] = "Single replication cluster detected - configuration is healthy"
		result["severity"] = "info"
		result["actionRequired"] = false
	} else {
		result["status"] = "no_replication"
		result["message"] = "No replication clusters configured"
		result["severity"] = "info"
		result["actionRequired"] = false
	}

	return result, nil
}
func (rs *ReplicationService) AddSiteReplicationSmart(aliases []string) (map[string]interface{}, error) {
	if len(aliases) == 0 {
		return nil, fmt.Errorf("no aliases provided")
	}

	logger.GetLogger().Info("Smart adding site replication", map[string]interface{}{
		"aliases": aliases,
	})

	// Step 1: Detect existing clusters
	existingClusters, err := rs.DetectReplicationClusters()
	if err != nil {
		return nil, fmt.Errorf("failed to detect existing clusters: %v", err)
	}

	result := map[string]interface{}{
		"action":           "",
		"clustersFound":    len(existingClusters),
		"existingClusters": existingClusters,
		"aliases":          aliases,
	}

	// Step 2: Check for split brain scenario (multiple clusters)
	if len(existingClusters) > 1 {
		logger.GetLogger().Warn("Multiple replication clusters detected - potential split brain", map[string]interface{}{
			"clusterCount": len(existingClusters),
			"clusters":     existingClusters,
		})

		result["action"] = "split_brain_detected"
		result["error"] = fmt.Sprintf("Split brain detected: %d separate replication clusters found. Please resolve this before adding new sites.", len(existingClusters))
		return result, fmt.Errorf("split brain scenario detected: %d separate clusters exist", len(existingClusters))
	}

	// Step 3: No existing cluster - create new replication
	if len(existingClusters) == 0 {
		if len(aliases) < 2 {
			return nil, fmt.Errorf("at least 2 aliases are required to create new replication cluster")
		}

		result["action"] = "create_new_cluster"

		err := rs.AddSiteReplication(aliases)
		if err != nil {
			result["error"] = err.Error()
			return result, err
		}

		result["success"] = true
		result["message"] = fmt.Sprintf("New replication cluster created with %d sites", len(aliases))
		return result, nil
	}

	// Step 4: Existing cluster found - add to existing cluster
	existingCluster := existingClusters[0]
	result["action"] = "add_to_existing_cluster"
	result["existingCluster"] = existingCluster

	// Get existing site names
	existingSites := make([]string, 0)
	for _, site := range existingCluster.Sites {
		if siteName, ok := site["name"].(string); ok {
			existingSites = append(existingSites, siteName)
		}
	}

	// Filter out aliases that are already in the cluster
	newAliases := make([]string, 0)
	alreadyInCluster := make([]string, 0)

	for _, alias := range aliases {
		found := false
		for _, existing := range existingSites {
			if alias == existing {
				found = true
				alreadyInCluster = append(alreadyInCluster, alias)
				break
			}
		}
		if !found {
			newAliases = append(newAliases, alias)
		}
	}

	result["newAliases"] = newAliases
	result["alreadyInCluster"] = alreadyInCluster

	if len(newAliases) == 0 {
		result["message"] = "All specified aliases are already in the replication cluster"
		result["success"] = true
		return result, nil
	}

	// Add new aliases to existing cluster
	allAliases := append(existingSites, newAliases...)

	err = rs.AddSiteReplication(allAliases)
	if err != nil {
		result["error"] = err.Error()
		return result, err
	}

	result["success"] = true
	result["message"] = fmt.Sprintf("Added %d new sites to existing cluster. Total sites: %d", len(newAliases), len(allAliases))

	return result, nil
}

// AddSiteReplication adds sites to replication
func (rs *ReplicationService) AddSiteReplication(aliases []string) error {
	if len(aliases) < 2 {
		return fmt.Errorf("at least 2 aliases are required")
	}

	logger.GetLogger().Info("Adding site replication", map[string]interface{}{
		"aliases": aliases,
	})

	args := []string{"admin", "replicate", "add"}
	args = append(args, aliases...)

	cmd := exec.Command("mc", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.GetLogger().Error("Failed to add site replication", map[string]interface{}{
			"error":  err.Error(),
			"output": string(output),
		})
		return fmt.Errorf("failed to add replication: %s", string(output))
	}

	logger.GetLogger().Info("Site replication added successfully", map[string]interface{}{
		"aliases": aliases,
		"output":  string(output),
	})

	return nil
}

// RemoveEntireReplication removes entire replication configuration
func (rs *ReplicationService) RemoveEntireReplication(alias string) error {
	// If no alias provided, find any alias with replication enabled
	if alias == "" {
		aliases, err := rs.minioService.GetAliases()
		if err != nil || len(aliases) == 0 {
			return fmt.Errorf("no aliases available for replication removal")
		}

		for _, a := range aliases {
			replicateCmd := exec.Command("mc", "admin", "replicate", "info", a.Name, "--json")
			if replicateOutput, err := replicateCmd.CombinedOutput(); err == nil {
				var replicateInfo map[string]interface{}
				if json.Unmarshal(replicateOutput, &replicateInfo) == nil {
					if enabled, ok := replicateInfo["enabled"].(bool); ok && enabled {
						alias = a.Name
						break
					}
				}
			}
		}

		if alias == "" {
			return fmt.Errorf("no site replication found to remove")
		}
	}

	logger.GetLogger().Info("Removing entire site replication", map[string]interface{}{
		"alias": alias,
	})

	cmd := exec.Command("mc", "admin", "replicate", "rm", alias, "--all", "--force")
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.GetLogger().Error("Failed to remove site replication", map[string]interface{}{
			"error":  err.Error(),
			"output": string(output),
		})
		return fmt.Errorf("failed to remove site replication: %s", string(output))
	}

	logger.GetLogger().Info("Site replication removed successfully", map[string]interface{}{
		"alias":  alias,
		"output": string(output),
	})

	return nil
}

// RemoveIndividualSitesSmart removes individual sites with split-brain awareness
func (rs *ReplicationService) RemoveIndividualSitesSmart(aliasesToRemove []string) ([]map[string]interface{}, []string, error) {
	if len(aliasesToRemove) == 0 {
		return nil, nil, fmt.Errorf("no aliases specified for removal")
	}

	logger.GetLogger().Info("Smart removing sites from replication", map[string]interface{}{
		"aliases": aliasesToRemove,
		"count":   len(aliasesToRemove),
	})

	// Detect existing clusters to handle split brain scenarios
	existingClusters, err := rs.DetectReplicationClusters()
	if err != nil {
		logger.GetLogger().Warn("Failed to detect clusters, falling back to standard removal", map[string]interface{}{
			"error": err.Error(),
		})
		return rs.RemoveIndividualSites(aliasesToRemove)
	}

	var results []map[string]interface{}
	var failed []string

	// Process each site to remove
	for _, aliasToRemove := range aliasesToRemove {
		// Find which cluster this site belongs to
		var targetCluster *ClusterInfo
		for i := range existingClusters {
			cluster := &existingClusters[i]
			for _, site := range cluster.Sites {
				if siteName, ok := site["name"].(string); ok && siteName == aliasToRemove {
					targetCluster = cluster
					break
				}
			}
			if targetCluster != nil {
				break
			}
		}

		if targetCluster == nil {
			// Site not found in any cluster
			failed = append(failed, aliasToRemove)
			results = append(results, map[string]interface{}{
				"alias":   aliasToRemove,
				"success": false,
				"error":   fmt.Sprintf("Site %s not found in any replication cluster", aliasToRemove),
			})
			continue
		}

		// Find a reference alias in the same cluster
		var referenceAlias string
		for _, site := range targetCluster.Sites {
			if siteName, ok := site["name"].(string); ok && siteName != aliasToRemove {
				referenceAlias = siteName
				break
			}
		}

		// Check if this would remove all sites from cluster
		if referenceAlias == "" {
			logger.GetLogger().Info("Removing entire cluster as only one site remains", map[string]interface{}{
				"site":    aliasToRemove,
				"cluster": targetCluster.Name,
			})

			err := rs.RemoveEntireReplication(aliasToRemove)
			if err != nil {
				failed = append(failed, aliasToRemove)
				results = append(results, map[string]interface{}{
					"alias":   aliasToRemove,
					"success": false,
					"error":   err.Error(),
				})
			} else {
				results = append(results, map[string]interface{}{
					"alias":   aliasToRemove,
					"success": true,
					"message": "Last site in cluster - entire replication configuration removed",
				})
			}
			continue
		}

		// Remove site using reference from same cluster
		logger.GetLogger().Info("Removing site from cluster", map[string]interface{}{
			"site":      aliasToRemove,
			"reference": referenceAlias,
			"cluster":   targetCluster.Name,
		})

		cmd := exec.Command("mc", "admin", "replicate", "rm", referenceAlias, aliasToRemove, "--force")
		output, err := cmd.CombinedOutput()

		if err != nil {
			logger.GetLogger().Error("Failed to remove site from cluster", map[string]interface{}{
				"site":   aliasToRemove,
				"error":  err.Error(),
				"output": string(output),
			})
			failed = append(failed, aliasToRemove)
			results = append(results, map[string]interface{}{
				"alias":   aliasToRemove,
				"success": false,
				"error":   string(output),
			})
		} else {
			logger.GetLogger().Info("Site removed from cluster successfully", map[string]interface{}{
				"site":   aliasToRemove,
				"output": string(output),
			})
			results = append(results, map[string]interface{}{
				"alias":   aliasToRemove,
				"success": true,
				"output":  string(output),
			})
		}
	}

	return results, failed, nil
}

// RemoveIndividualSites removes individual sites from replication
func (rs *ReplicationService) RemoveIndividualSites(aliasesToRemove []string) ([]map[string]interface{}, []string, error) {
	if len(aliasesToRemove) == 0 {
		return nil, nil, fmt.Errorf("no aliases specified for removal")
	}

	logger.GetLogger().Info("Removing sites from replication", map[string]interface{}{
		"aliases": aliasesToRemove,
		"count":   len(aliasesToRemove),
	})

	// Find reference alias
	referenceAlias, err := rs.findReferenceAlias(aliasesToRemove)
	if err != nil {
		return nil, nil, err
	}

	// Check if removing all sites
	if referenceAlias == "" {
		if len(aliasesToRemove) > 0 {
			// Remove entire replication
			err := rs.RemoveEntireReplication(aliasesToRemove[0])
			if err != nil {
				return nil, aliasesToRemove, err
			}

			results := make([]map[string]interface{}, len(aliasesToRemove))
			for i, alias := range aliasesToRemove {
				results[i] = map[string]interface{}{
					"alias":   alias,
					"success": true,
					"message": "All sites removed - entire replication configuration removed",
				}
			}
			return results, nil, nil
		}
		return nil, nil, fmt.Errorf("no suitable reference alias found for site removal")
	}

	// Remove each site individually
	var results []map[string]interface{}
	var failed []string

	for _, aliasToRemove := range aliasesToRemove {
		logger.GetLogger().Info("Removing individual site from replication", map[string]interface{}{
			"site":      aliasToRemove,
			"reference": referenceAlias,
		})

		cmd := exec.Command("mc", "admin", "replicate", "rm", referenceAlias, aliasToRemove, "--force")
		output, err := cmd.CombinedOutput()

		if err != nil {
			logger.GetLogger().Error("Failed to remove site from replication", map[string]interface{}{
				"site":   aliasToRemove,
				"error":  err.Error(),
				"output": string(output),
			})
			failed = append(failed, aliasToRemove)
			results = append(results, map[string]interface{}{
				"alias":   aliasToRemove,
				"success": false,
				"error":   string(output),
			})
		} else {
			logger.GetLogger().Info("Site removed from replication successfully", map[string]interface{}{
				"site":   aliasToRemove,
				"output": string(output),
			})
			results = append(results, map[string]interface{}{
				"alias":   aliasToRemove,
				"success": true,
				"output":  string(output),
			})
		}
	}

	return results, failed, nil
}

// findReferenceAlias finds a suitable reference alias for site removal
func (rs *ReplicationService) findReferenceAlias(aliasesToRemove []string) (string, error) {
	aliases, err := rs.minioService.GetAliases()
	if err != nil {
		return "", fmt.Errorf("failed to get aliases list: %v", err)
	}

	for _, alias := range aliases {
		aliasName := alias.Name

		// Check if this alias is NOT in the removal list
		shouldSkip := false
		for _, toRemove := range aliasesToRemove {
			if aliasName == toRemove {
				shouldSkip = true
				break
			}
		}

		if !shouldSkip {
			// Check if it has replication enabled
			replicateCmd := exec.Command("mc", "admin", "replicate", "info", aliasName, "--json")
			if replicateOutput, err := replicateCmd.CombinedOutput(); err == nil {
				var replicateInfo map[string]interface{}
				if json.Unmarshal(replicateOutput, &replicateInfo) == nil {
					if enabled, ok := replicateInfo["enabled"].(bool); ok && enabled {
						return aliasName, nil
					}
				}
			}
		}
	}

	return "", nil // No reference alias found, might need to remove entire replication
}

// ResyncSites handles resyncing data between sites
func (rs *ReplicationService) ResyncSites(sourceAlias, targetAlias, direction string) error {
	if sourceAlias == "" {
		return fmt.Errorf("source alias is required")
	}

	if direction != "resync-from" && direction != "resync-to" {
		return fmt.Errorf("direction must be 'resync-from' or 'resync-to'")
	}

	logger.GetLogger().Info("Starting site replication resync", map[string]interface{}{
		"source_alias": sourceAlias,
		"target_alias": targetAlias,
		"direction":    direction,
	})

	var cmd *exec.Cmd
	if targetAlias == "" {
		return fmt.Errorf("target alias is required")
	}

	cmd = exec.Command("mc", "admin", "replicate", "resync", "start",
		"--deployment-id", targetAlias, sourceAlias)

	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.GetLogger().Error("Failed to start resync", map[string]interface{}{
			"error":  err.Error(),
			"output": string(output),
		})
		return fmt.Errorf("failed to start resync: %s", string(output))
	}

	logger.GetLogger().Info("Resync started successfully", map[string]interface{}{
		"source_alias": sourceAlias,
		"target_alias": targetAlias,
		"direction":    direction,
		"output":       string(output),
	})

	return nil
}
