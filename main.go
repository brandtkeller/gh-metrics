package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Metrics struct {
	RepositoryName        string `json:"full_name"`
	Stars                 int    `json:"stargazers_count"`
	Forks                 int    `json:"forks_count"`
	OpenIssues            int    `json:"open_issues_count"`
	Watchers              int    `json:"subscribers_count"`
	TotalIssues           int
	TotalReleaseDownloads int
	LastUpdated           string `json:"updated_at"`
}

func getAllIssuesCount(baseURL string) (int, error) {
	totalIssues := 0
	page := 1

	for {
		issuesResp, err := http.Get(fmt.Sprintf("%s/issues?state=all&per_page=100&page=%d", baseURL, page))
		if err != nil {
			return 0, err
		}
		defer issuesResp.Body.Close()

		issuesData, err := ioutil.ReadAll(issuesResp.Body)
		if err != nil {
			return 0, err
		}

		issues := []map[string]interface{}{}
		if err := json.Unmarshal(issuesData, &issues); err != nil {
			return 0, err
		}

		if len(issues) == 0 {
			break
		}

		for _, issue := range issues {
			// Exclude pull requests by checking for the "pull_request" field
			if _, isPR := issue["pull_request"]; !isPR {
				totalIssues++
			}
		}

		page++
	}

	return totalIssues, nil
}

func getGitHubMetrics(owner, repo string) (Metrics, error) {
	baseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	// Retrieve repository details
	repoResp, err := http.Get(baseURL)
	if err != nil {
		return Metrics{}, err
	}
	defer repoResp.Body.Close()
	repoData, err := ioutil.ReadAll(repoResp.Body)
	if err != nil {
		return Metrics{}, err
	}

	repoMetrics := Metrics{}
	if err := json.Unmarshal(repoData, &repoMetrics); err != nil {
		return Metrics{}, err
	}

	// Retrieve release details
	releasesResp, err := http.Get(baseURL + "/releases")
	if err != nil {
		return Metrics{}, err
	}
	defer releasesResp.Body.Close()

	releasesData, err := ioutil.ReadAll(releasesResp.Body)
	if err != nil {
		return Metrics{}, err
	}

	releases := []map[string]interface{}{}
	if err := json.Unmarshal(releasesData, &releases); err != nil {
		if json.Unmarshal(releasesData, &map[string]interface{}{}) == nil {
			releases = []map[string]interface{}{}
		} else {
			return Metrics{}, err
		}
	}

	totalDownloads := 0
	for _, release := range releases {
		if assets, ok := release["assets"].([]interface{}); ok {
			for _, asset := range assets {
				if assetMap, ok := asset.(map[string]interface{}); ok {
					if count, ok := assetMap["download_count"].(float64); ok {
						totalDownloads += int(count)
					}
				}
			}
		}
	}

	repoMetrics.TotalReleaseDownloads = totalDownloads

	// Retrieve total issues count
	totalIssues, err := getAllIssuesCount(baseURL)
	if err != nil {
		return Metrics{}, err
	}
	repoMetrics.TotalIssues = totalIssues

	return repoMetrics, nil
}

func formatMonthlyReport(current, previous Metrics) string {
	deltas := Metrics{
		Stars:                 current.Stars - previous.Stars,
		Forks:                 current.Forks - previous.Forks,
		OpenIssues:            current.OpenIssues - previous.OpenIssues,
		Watchers:              current.Watchers - previous.Watchers,
		TotalIssues:           current.TotalIssues - previous.TotalIssues,
		TotalReleaseDownloads: current.TotalReleaseDownloads - previous.TotalReleaseDownloads,
	}

	return fmt.Sprintf(
		"GitHub Repository Monthly Report\n"+
			"===================================\n"+
			"Repository: %s\n"+
			"Stars: %d (Delta: %+d)\n"+
			"Forks: %d (Delta: %+d)\n"+
			"Watchers: %d (Delta: %+d)\n"+
			"Open Issues: %d (Delta: %+d)\n"+
			"Total Issues: %d (Delta: %+d)\n"+
			"Total Release Downloads: %d (Delta: %+d)\n"+
			"Last Updated: %s\n"+
			"===================================\n",
		current.RepositoryName,
		current.Stars, deltas.Stars,
		current.Forks, deltas.Forks,
		current.Watchers, deltas.Watchers,
		current.OpenIssues, deltas.OpenIssues,
		current.TotalIssues, deltas.TotalIssues,
		current.TotalReleaseDownloads, deltas.TotalReleaseDownloads,
		current.LastUpdated,
	)
}

func getMetricsFilePath(repoName string, timestamp string) string {
	return filepath.Join("metrics", fmt.Sprintf("%s_metrics_%s.json", repoName, timestamp))
}

func main() {
	owner := flag.String("owner", "octocat", "The owner of the GitHub repository")
	repo := flag.String("repo", "Hello-World", "The name of the GitHub repository")
	previousFile := flag.String("previous", "", "Path to the previous metrics file")
	flag.Parse()

	// Ensure the metrics directory exists
	os.MkdirAll("metrics", os.ModePerm)

	// Get current timestamp for file naming
	timestamp := time.Now().Format("20060102_150405")
	metricsFile := getMetricsFilePath(*repo, timestamp)

	// Load previous metrics
	var previous Metrics
	if *previousFile != "" {
		file, err := os.Open(*previousFile)
		if err == nil {
			defer file.Close()
			data, _ := ioutil.ReadAll(file)
			json.Unmarshal(data, &previous)
		}
	}

	current, err := getGitHubMetrics(*owner, *repo)
	if err != nil {
		fmt.Printf("Error fetching data from GitHub: %v\n", err)
		return
	}

	report := formatMonthlyReport(current, previous)
	fmt.Println(report)

	// Save current metrics for future comparison
	data, _ := json.Marshal(current)
	ioutil.WriteFile(metricsFile, data, 0644)
}
