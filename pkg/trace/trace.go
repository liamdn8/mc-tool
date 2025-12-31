package trace

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Options holds configuration for running an mc admin trace capture.
type Options struct {
	Alias               string
	MCPath              string
	Duration            time.Duration
	Insecure            bool
	Verbose             bool
	StatusCodeFilters   []int
	ErrorMessageFilters []string
	GroupByAPI          bool
	GroupByClient       bool
	GroupByVersions     bool
	TraceErrorsOnly     bool // If true, only trace errors; otherwise trace all requests
}

// ObjectStat represents aggregated error occurrences for a given object.
type ObjectStat struct {
	Name         string
	Count        int
	SampleErrors []string
	ErrorCounts  map[string]int
}

// ObjectCount represents aggregated event counts for a specific object key.
type ObjectCount struct {
	Name  string
	Count int
}

// ErrorStat groups events by shared error message/code.
type ErrorStat struct {
	Message string
	Count   int
	Objects []ObjectCount
}

// Result captures the overall trace summary output.
type Result struct {
	Start       time.Time
	End         time.Time
	Duration    time.Duration
	TotalEvents int
	Stats       []ObjectStat
	RawErrors   []string
	ErrorStats  []ErrorStat
	APIStats    []APIStat
	ClientStats []ClientStat
}

type errorAccumulator struct {
	Message string
	Count   int
	Objects map[string]int
}

type apiAccumulator struct {
	Name        string
	Count       int
	ErrorCounts map[string]int
	Objects     map[string]int
}

type APIStat struct {
	Name        string
	Count       int
	ErrorCounts map[string]int
	Objects     []ObjectCount
}

type clientAccumulator struct {
	Name        string
	Count       int
	ErrorCounts map[string]int
	Objects     map[string]int
}

type ClientStat struct {
	Name        string
	Count       int
	ErrorCounts map[string]int
	Objects     []ObjectCount
}

// Run executes mc admin trace according to opts and returns aggregated results.
func Run(opts Options) (Result, error) {
	if opts.Alias == "" {
		return Result{}, errors.New("alias is required")
	}
	if opts.Duration <= 0 {
		return Result{}, errors.New("duration must be > 0")
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.Duration)
	defer cancel()

	args := []string{"admin", "trace"}
	if opts.TraceErrorsOnly {
		args = append(args, "--errors")
	}
	args = append(args, "--json")
	if opts.Insecure {
		args = append(args, "--insecure")
	}
	args = append(args, opts.Alias)

	cmd := exec.CommandContext(ctx, opts.MCPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{}, fmt.Errorf("failed to get stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return Result{}, fmt.Errorf("failed to get stderr: %w", err)
	}

	if err = cmd.Start(); err != nil {
		return Result{}, fmt.Errorf("failed to start mc admin trace: %w", err)
	}

	start := time.Now()
	statsMap := map[string]*ObjectStat{}
	errorGroups := map[string]*errorAccumulator{}
	var apiGroups map[string]*apiAccumulator
	if opts.GroupByAPI {
		apiGroups = map[string]*apiAccumulator{}
	}
	var clientGroups map[string]*clientAccumulator
	if opts.GroupByClient {
		clientGroups = map[string]*clientAccumulator{}
	}
	rawErrors := make([]string, 0)
	totalEvents := 0

	statusFilter := make(map[int]struct{}, len(opts.StatusCodeFilters))
	for _, code := range opts.StatusCodeFilters {
		statusFilter[code] = struct{}{}
	}

	messageFilters := make([]string, 0, len(opts.ErrorMessageFilters))
	for _, filter := range opts.ErrorMessageFilters {
		trimmed := strings.ToLower(strings.TrimSpace(filter))
		if trimmed != "" {
			messageFilters = append(messageFilters, trimmed)
		}
	}

	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 256*1024)
	scanner.Buffer(buf, 2*1024*1024)

	readErrCh := make(chan error, 1)
	go func() {
		readErrCh <- readTrace(scanner, statsMap, errorGroups, apiGroups, clientGroups, statusFilter, messageFilters, &totalEvents, &rawErrors, opts.GroupByVersions, opts.Verbose)
	}()

	stderrBuf, _ := io.ReadAll(stderr)

	readErr := <-readErrCh
	waitErr := cmd.Wait()
	end := time.Now()

	// If the context deadline triggered, cmd.Wait will return an error wrapping context deadline exceeded.
	if ctx.Err() == context.DeadlineExceeded {
		// treat deadline exceeded as expected. overwrite waitErr.
		waitErr = nil
	}

	if readErr != nil && !errors.Is(readErr, context.DeadlineExceeded) {
		return Result{}, readErr
	}

	if waitErr != nil {
		// provide stderr output to help troubleshooting
		if len(stderrBuf) > 0 {
			return Result{}, fmt.Errorf("mc admin trace failed: %w: %s", waitErr, strings.TrimSpace(string(stderrBuf)))
		}
		return Result{}, fmt.Errorf("mc admin trace failed: %w", waitErr)
	}

	// Convert map to sorted slice
	stats := make([]ObjectStat, 0, len(statsMap))
	for _, entry := range statsMap {
		stats = append(stats, *entry)
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count == stats[j].Count {
			return stats[i].Name < stats[j].Name
		}
		return stats[i].Count > stats[j].Count
	})

	// Convert error groups to sorted slice
	errorStats := make([]ErrorStat, 0, len(errorGroups))
	for _, group := range errorGroups {
		objectCounts := make([]ObjectCount, 0, len(group.Objects))
		for name, count := range group.Objects {
			objectCounts = append(objectCounts, ObjectCount{Name: name, Count: count})
		}
		sort.Slice(objectCounts, func(i, j int) bool {
			if objectCounts[i].Count == objectCounts[j].Count {
				return objectCounts[i].Name < objectCounts[j].Name
			}
			return objectCounts[i].Count > objectCounts[j].Count
		})

		errorStats = append(errorStats, ErrorStat{
			Message: group.Message,
			Count:   group.Count,
			Objects: objectCounts,
		})
	}

	sort.Slice(errorStats, func(i, j int) bool {
		if errorStats[i].Count == errorStats[j].Count {
			return errorStats[i].Message < errorStats[j].Message
		}
		return errorStats[i].Count > errorStats[j].Count
	})

	var apiStats []APIStat
	if opts.GroupByAPI && len(apiGroups) > 0 {
		apiStats = make([]APIStat, 0, len(apiGroups))
		for _, group := range apiGroups {
			objectCounts := make([]ObjectCount, 0, len(group.Objects))
			for name, count := range group.Objects {
				objectCounts = append(objectCounts, ObjectCount{Name: name, Count: count})
			}
			sort.Slice(objectCounts, func(i, j int) bool {
				if objectCounts[i].Count == objectCounts[j].Count {
					return objectCounts[i].Name < objectCounts[j].Name
				}
				return objectCounts[i].Count > objectCounts[j].Count
			})

			errorCounts := make(map[string]int, len(group.ErrorCounts))
			for msg, count := range group.ErrorCounts {
				errorCounts[msg] = count
			}

			apiStats = append(apiStats, APIStat{
				Name:        group.Name,
				Count:       group.Count,
				ErrorCounts: errorCounts,
				Objects:     objectCounts,
			})
		}

		sort.Slice(apiStats, func(i, j int) bool {
			if apiStats[i].Count == apiStats[j].Count {
				return apiStats[i].Name < apiStats[j].Name
			}
			return apiStats[i].Count > apiStats[j].Count
		})
	}

	var clientStats []ClientStat
	if opts.GroupByClient && len(clientGroups) > 0 {
		clientStats = make([]ClientStat, 0, len(clientGroups))
		for _, group := range clientGroups {
			objectCounts := make([]ObjectCount, 0, len(group.Objects))
			for name, count := range group.Objects {
				objectCounts = append(objectCounts, ObjectCount{Name: name, Count: count})
			}
			sort.Slice(objectCounts, func(i, j int) bool {
				if objectCounts[i].Count == objectCounts[j].Count {
					return objectCounts[i].Name < objectCounts[j].Name
				}
				return objectCounts[i].Count > objectCounts[j].Count
			})

			errorCounts := make(map[string]int, len(group.ErrorCounts))
			for msg, count := range group.ErrorCounts {
				errorCounts[msg] = count
			}

			clientStats = append(clientStats, ClientStat{
				Name:        group.Name,
				Count:       group.Count,
				ErrorCounts: errorCounts,
				Objects:     objectCounts,
			})
		}

		sort.Slice(clientStats, func(i, j int) bool {
			if clientStats[i].Count == clientStats[j].Count {
				return clientStats[i].Name < clientStats[j].Name
			}
			return clientStats[i].Count > clientStats[j].Count
		})
	}

	result := Result{
		Start:       start,
		End:         end,
		Duration:    end.Sub(start),
		TotalEvents: totalEvents,
		Stats:       stats,
		RawErrors:   rawErrors,
		ErrorStats:  errorStats,
		APIStats:    apiStats,
		ClientStats: clientStats,
	}

	return result, nil
}

func readTrace(
	scanner *bufio.Scanner,
	statsMap map[string]*ObjectStat,
	errorGroups map[string]*errorAccumulator,
	apiGroups map[string]*apiAccumulator,
	clientGroups map[string]*clientAccumulator,
	statusFilter map[int]struct{},
	messageFilters []string,
	totalEvents *int,
	rawErrors *[]string,
	groupByVersions bool,
	verbose bool,
) error {
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		entry := map[string]interface{}{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			if len(statusFilter) > 0 || len(messageFilters) > 0 {
				continue
			}
			*totalEvents = *totalEvents + 1
			*rawErrors = append(*rawErrors, line)
			recordEvent(statsMap, errorGroups, apiGroups, clientGroups, "<unknown>", "", "JSON parse error", "<unknown>", "<unknown>", groupByVersions)
			continue
		}

		statusCode := extractStatusCode(entry)
		if len(statusFilter) > 0 {
			if _, ok := statusFilter[statusCode]; !ok {
				continue
			}
		}

		objectName := extractObject(entry)
		versionID := extractVersionID(entry)
		errorMsg := extractError(entry)

		if len(messageFilters) > 0 {
			lowerMsg := strings.ToLower(errorMsg)
			matched := false
			for _, filter := range messageFilters {
				if strings.Contains(lowerMsg, filter) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		apiName := extractAPI(entry)
		clientName := extractClient(entry)

		*totalEvents = *totalEvents + 1
		*rawErrors = append(*rawErrors, line)
		recordEvent(statsMap, errorGroups, apiGroups, clientGroups, objectName, versionID, errorMsg, apiName, clientName, groupByVersions)
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func recordEvent(
	statsMap map[string]*ObjectStat,
	errorGroups map[string]*errorAccumulator,
	apiGroups map[string]*apiAccumulator,
	clientGroups map[string]*clientAccumulator,
	name, versionID, errMsg, apiName, clientName string,
	groupByVersions bool,
) {
	key := strings.TrimSpace(name)
	if key == "" {
		key = "<unknown>"
	}

	// If grouping by versions and we have a version ID, append it to the key
	if groupByVersions && versionID != "" {
		key = fmt.Sprintf("%s (version: %s)", key, versionID)
	}

	msg := strings.TrimSpace(errMsg)
	if msg == "" {
		msg = "<unknown error>"
	}

	stat, ok := statsMap[key]
	if !ok {
		stat = &ObjectStat{Name: key, ErrorCounts: map[string]int{}}
		statsMap[key] = stat
	}
	stat.Count++
	if stat.ErrorCounts == nil {
		stat.ErrorCounts = map[string]int{}
	}
	stat.ErrorCounts[msg] = stat.ErrorCounts[msg] + 1
	if msg != "" && len(stat.SampleErrors) < 3 {
		found := false
		for _, existing := range stat.SampleErrors {
			if existing == msg {
				found = true
				break
			}
		}
		if !found {
			stat.SampleErrors = append(stat.SampleErrors, msg)
		}
	}

	group, ok := errorGroups[msg]
	if !ok {
		group = &errorAccumulator{
			Message: msg,
			Objects: map[string]int{},
		}
		errorGroups[msg] = group
	}
	group.Count++
	group.Objects[key] = group.Objects[key] + 1

	if apiGroups != nil {
		apiKey := strings.TrimSpace(apiName)
		if apiKey == "" {
			apiKey = "<unknown>"
		}

		apiGroup, ok := apiGroups[apiKey]
		if !ok {
			apiGroup = &apiAccumulator{
				Name:        apiKey,
				ErrorCounts: map[string]int{},
				Objects:     map[string]int{},
			}
			apiGroups[apiKey] = apiGroup
		}
		apiGroup.Count++
		apiGroup.Objects[key] = apiGroup.Objects[key] + 1
		if apiGroup.ErrorCounts == nil {
			apiGroup.ErrorCounts = map[string]int{}
		}
		apiGroup.ErrorCounts[msg] = apiGroup.ErrorCounts[msg] + 1
	}

	if clientGroups != nil {
		clientKey := strings.TrimSpace(clientName)
		if clientKey == "" {
			clientKey = "<unknown>"
		}

		clientGroup, ok := clientGroups[clientKey]
		if !ok {
			clientGroup = &clientAccumulator{
				Name:        clientKey,
				ErrorCounts: map[string]int{},
				Objects:     map[string]int{},
			}
			clientGroups[clientKey] = clientGroup
		}
		clientGroup.Count++
		clientGroup.Objects[key] = clientGroup.Objects[key] + 1
		if clientGroup.ErrorCounts == nil {
			clientGroup.ErrorCounts = map[string]int{}
		}
		clientGroup.ErrorCounts[msg] = clientGroup.ErrorCounts[msg] + 1
	}
}

func extractObject(entry map[string]interface{}) string {
	// Common fields in trace output
	candidates := []string{"object", "objectName", "resource", "key", "Key", "name", "path", "Path", "bucket", "Bucket", "bucketName"}
	for _, field := range candidates {
		if v, ok := entry[field]; ok {
			if s := toString(v); s != "" {
				// Remove version ID from path if present
				return normalizePathAndStripVersion(s)
			}
		}
	}

	if v, ok := entry["req"]; ok {
		if reqMap, ok := v.(map[string]interface{}); ok {
			for _, field := range []string{"path", "Path"} {
				if pathVal, ok := reqMap[field]; ok {
					if s := toString(pathVal); s != "" {
						return normalizePathAndStripVersion(s)
					}
				}
			}
			if urlVal, ok := reqMap["url"]; ok {
				if s := toString(urlVal); s != "" {
					return normalizePathAndStripVersion(s)
				}
			}
		}
	}

	if v, ok := entry["resourceType"]; ok {
		if s := toString(v); s != "" {
			return normalizePathAndStripVersion(s)
		}
	}

	return ""
}

func extractError(entry map[string]interface{}) string {
	candidates := []string{"error", "Error", "message", "msg", "err"}
	for _, field := range candidates {
		if v, ok := entry[field]; ok {
			if s := toString(v); s != "" {
				return s
			}
		}
	}

	if v, ok := entry["resp"]; ok {
		if respMap, ok := v.(map[string]interface{}); ok {
			if errVal, ok := respMap["error"]; ok {
				if s := toString(errVal); s != "" {
					return s
				}
			}
			if code, ok := respMap["status"].(float64); ok {
				if code >= 400 {
					return fmt.Sprintf("HTTP status %d", int(code))
				}
			}
		}
	}

	if v, ok := entry["status"].(float64); ok {
		if v >= 400 {
			return fmt.Sprintf("HTTP status %d", int(v))
		}
	}

	if codeVal, ok := entry["statusCode"].(float64); ok {
		code := int(codeVal)
		if code >= 400 {
			statusMsg := toString(entry["statusMsg"])
			if statusMsg != "" {
				return fmt.Sprintf("HTTP %d %s", code, statusMsg)
			}
			if api := toString(entry["api"]); api != "" {
				return fmt.Sprintf("%s returned HTTP %d", api, code)
			}
			return fmt.Sprintf("HTTP status %d", code)
		}
	}

	return ""
}

func extractStatusCode(entry map[string]interface{}) int {
	if codeVal, ok := entry["statusCode"].(float64); ok {
		return int(codeVal)
	}
	if codeStr := toString(entry["statusCode"]); codeStr != "" {
		if parsed, err := strconv.Atoi(codeStr); err == nil {
			return parsed
		}
	}

	if statusVal, ok := entry["status"].(float64); ok {
		return int(statusVal)
	}
	if statusStr := toString(entry["status"]); statusStr != "" {
		if parsed, err := strconv.Atoi(statusStr); err == nil {
			return parsed
		}
	}

	if resp, ok := entry["resp"]; ok {
		if respMap, ok := resp.(map[string]interface{}); ok {
			if statusVal, ok := respMap["status"].(float64); ok {
				return int(statusVal)
			}
			if statusStr := toString(respMap["status"]); statusStr != "" {
				if parsed, err := strconv.Atoi(statusStr); err == nil {
					return parsed
				}
			}

			if codeVal, ok := respMap["statusCode"].(float64); ok {
				return int(codeVal)
			}
			if codeStr := toString(respMap["statusCode"]); codeStr != "" {
				if parsed, err := strconv.Atoi(codeStr); err == nil {
					return parsed
				}
			}
		}
	}

	if traceInfo, ok := entry["trace"].(map[string]interface{}); ok {
		if codeVal, ok := traceInfo["statusCode"].(float64); ok {
			return int(codeVal)
		}
		if codeStr := toString(traceInfo["statusCode"]); codeStr != "" {
			if parsed, err := strconv.Atoi(codeStr); err == nil {
				return parsed
			}
		}
	}

	return 0
}

func extractAPI(entry map[string]interface{}) string {
	candidates := []string{"api", "API", "action", "Action", "method", "operation"}
	for _, field := range candidates {
		if v, ok := entry[field]; ok {
			if s := toString(v); s != "" {
				return s
			}
		}
	}

	if req, ok := entry["req"]; ok {
		if reqMap, ok := req.(map[string]interface{}); ok {
			for _, field := range []string{"api", "API", "method", "action"} {
				if v, ok := reqMap[field]; ok {
					if s := toString(v); s != "" {
						return s
					}
				}
			}
		}
	}

	if traceInfo, ok := entry["trace"].(map[string]interface{}); ok {
		for _, field := range []string{"api", "API", "method"} {
			if v, ok := traceInfo[field]; ok {
				if s := toString(v); s != "" {
					return s
				}
			}
		}
	}

	return ""
}

func extractClient(entry map[string]interface{}) string {
	candidates := []string{
		"remoteHost", "remotehost", "remote_host", "host", "Host", "client", "clientHost", "clientAddr", "clientAddress",
		"remoteAddr", "remoteaddr", "RemoteAddr", "source", "Source", "address", "Address",
	}
	for _, field := range candidates {
		if v, ok := entry[field]; ok {
			if s := toString(v); s != "" {
				return s
			}
		}
	}

	if req, ok := entry["req"]; ok {
		if reqMap, ok := req.(map[string]interface{}); ok {
			for _, field := range []string{"remoteHost", "remotehost", "remoteAddr", "client", "clientAddr", "ip"} {
				if v, ok := reqMap[field]; ok {
					if s := toString(v); s != "" {
						return s
					}
				}
			}
		}
	}

	if traceInfo, ok := entry["trace"].(map[string]interface{}); ok {
		for _, field := range []string{"remoteHost", "remoteAddr", "client"} {
			if v, ok := traceInfo[field]; ok {
				if s := toString(v); s != "" {
					return s
				}
			}
		}
	}

	return ""
}

func extractVersionID(entry map[string]interface{}) string {
	// First, check if version ID is in the path/URL
	pathCandidates := []string{"path", "Path", "resource", "object", "objectName", "url"}
	for _, field := range pathCandidates {
		if v, ok := entry[field]; ok {
			if s := toString(v); s != "" {
				if ver := extractVersionFromPath(s); ver != "" {
					return ver
				}
			}
		}
	}

	// Check in req.path or req.url
	if req, ok := entry["req"]; ok {
		if reqMap, ok := req.(map[string]interface{}); ok {
			for _, field := range []string{"path", "Path", "url"} {
				if pathVal, ok := reqMap[field]; ok {
					if s := toString(pathVal); s != "" {
						if ver := extractVersionFromPath(s); ver != "" {
							return ver
						}
					}
				}
			}
		}
	}

	// Common version ID fields in trace output
	candidates := []string{"versionId", "versionID", "version-id", "VersionId", "VersionID"}
	for _, field := range candidates {
		if v, ok := entry[field]; ok {
			if s := toString(v); s != "" {
				return s
			}
		}
	}

	// Check in query parameters
	if req, ok := entry["req"]; ok {
		if reqMap, ok := req.(map[string]interface{}); ok {
			// Check query parameters
			if query, ok := reqMap["query"]; ok {
				if queryMap, ok := query.(map[string]interface{}); ok {
					for _, field := range []string{"versionId", "versionID", "version-id"} {
						if v, ok := queryMap[field]; ok {
							if s := toString(v); s != "" {
								return s
							}
						}
					}
				}
			}
			// Check headers
			if headers, ok := reqMap["headers"]; ok {
				if headersMap, ok := headers.(map[string]interface{}); ok {
					for _, field := range []string{"X-Amz-Version-Id", "x-amz-version-id"} {
						if v, ok := headersMap[field]; ok {
							if s := toString(v); s != "" {
								return s
							}
						}
					}
				}
			}
		}
	}

	// Check in response
	if resp, ok := entry["resp"]; ok {
		if respMap, ok := resp.(map[string]interface{}); ok {
			if headers, ok := respMap["headers"]; ok {
				if headersMap, ok := headers.(map[string]interface{}); ok {
					for _, field := range []string{"X-Amz-Version-Id", "x-amz-version-id"} {
						if v, ok := headersMap[field]; ok {
							if s := toString(v); s != "" {
								return s
							}
						}
					}
				}
			}
		}
	}

	return ""
}

func toString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return strings.TrimSpace(v.String())
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	default:
		return ""
	}
}

func normalizePath(s string) string {
	if s == "" {
		return s
	}
	trimmed := strings.TrimSpace(s)
	trimmed = strings.TrimPrefix(trimmed, "http://")
	trimmed = strings.TrimPrefix(trimmed, "https://")
	trimmed = strings.TrimPrefix(trimmed, "/")
	return trimmed
}

// normalizePathAndStripVersion removes version ID from path and normalizes it
// Handles both "path?versionId=xxx" and URL-encoded "path%3FversionId%3Dxxx"
func normalizePathAndStripVersion(s string) string {
	normalized := normalizePath(s)

	// Handle URL-encoded query strings (%3F = ?, %3D = =)
	decoded := strings.ReplaceAll(normalized, "%3F", "?")
	decoded = strings.ReplaceAll(decoded, "%3f", "?")
	decoded = strings.ReplaceAll(decoded, "%3D", "=")
	decoded = strings.ReplaceAll(decoded, "%3d", "=")

	// Strip query parameters (including versionId)
	if idx := strings.Index(decoded, "?"); idx >= 0 {
		return decoded[:idx]
	}

	return normalized
}

// extractVersionFromPath extracts version ID from URL path or query string
// Handles both "path?versionId=xxx" and URL-encoded "path%3FversionId%3Dxxx"
func extractVersionFromPath(s string) string {
	if s == "" {
		return ""
	}

	// Handle URL-encoded query strings
	decoded := strings.ReplaceAll(s, "%3F", "?")
	decoded = strings.ReplaceAll(decoded, "%3f", "?")
	decoded = strings.ReplaceAll(decoded, "%3D", "=")
	decoded = strings.ReplaceAll(decoded, "%3d", "=")
	decoded = strings.ReplaceAll(decoded, "%26", "&")
	decoded = strings.ReplaceAll(decoded, "%26", "&")

	// Look for version ID in query string
	if idx := strings.Index(decoded, "?"); idx >= 0 {
		queryString := decoded[idx+1:]
		params := strings.Split(queryString, "&")
		for _, param := range params {
			if strings.HasPrefix(strings.ToLower(param), "versionid=") {
				parts := strings.SplitN(param, "=", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}

	return ""
}
