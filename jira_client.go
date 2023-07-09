package main

import (
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"log"
	"math"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
)

type Credential struct {
	ApiUser string
	ApiKey  string
}

type JiraClient struct {
	Domain     string
	Credential Credential
}

// Issue By Issue Picker - struct
type Issue struct {
	SummaryText, Key string
}

type Section struct {
	Label, Sub, Id string
	Issues         []Issue
}

type IssuePickerResp struct {
	Sections []Section
}

// Issue By JQL - struct

type Fields struct {
	Summary string
}

type IssueJql struct {
	Key    string
	Fields Fields
}

type IssueJqlResp struct {
	Issues []IssueJql
}

func (j *JiraClient) call(req *http.Request) (*http.Response, error) {
	client := http.Client{}
	req.SetBasicAuth(j.Credential.ApiUser, j.Credential.ApiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

func (j *JiraClient) search() {
	args := getUserInput()
	issues := j.getIssuesUsingIssuePicker(args)
	selectedIssueKey := loadIssueListPrompt(issues)
	if len(selectedIssueKey) > 0 {
		j.openIssueInBrowser(selectedIssueKey)
	}
}

func (j *JiraClient) currentUserSearch(jql string) {
	issues := j.getIssuesUsingJql(jql)
	selectedIssueKey := loadIssueListPrompt(issues[:])
	if len(selectedIssueKey) > 0 {
		j.openIssueInBrowser(selectedIssueKey)
	}
}

func (j *JiraClient) advancedJqlSearch(jql string) {
	if len(jql) == 0 {
		jql = getUserInput()
	}
	issues := j.getIssuesUsingJql(jql)
	selectedIssueKey := loadIssueListPrompt(issues[:])
	if len(selectedIssueKey) > 0 {
		j.openIssueInBrowser(selectedIssueKey)
	}
}

func (j *JiraClient) openIssueInBrowser(issueKey string) {
	jiraLink := fmt.Sprintf("%s/browse/%s", j.Domain, issueKey)
	fmt.Printf("Opening %s in browser\n", issueKey)
	openbrowser(jiraLink)
}

// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/#api-rest-api-3-issue-picker-get
func (j *JiraClient) getIssuesUsingIssuePicker(args string) []Issue {
	// Call Rest API
	requestUrl := fmt.Sprintf("%s/rest/api/3/issue/picker?query=%s", j.Domain, url.QueryEscape(args))
	req, err := http.NewRequest("GET", requestUrl, nil)
	resp, err := j.call(req)
	if err != nil || resp.StatusCode != 200 {
		log.Fatal("Error Occurred: ", resp.StatusCode)
	}
	defer resp.Body.Close()

	var issuePickerResp IssuePickerResp
	err = json.NewDecoder(resp.Body).Decode(&issuePickerResp)
	if err != nil {
		log.Fatal(err)
	}
	return issuePickerResp.Sections[0].Issues
}

func (j *JiraClient) getIssuesUsingJql(jql string) []Issue {
	// Call Rest API
	requestUrl := fmt.Sprintf("%s/rest/api/3/search?jql=%s", j.Domain, url.QueryEscape(jql))
	req, err := http.NewRequest("GET", requestUrl, nil)
	resp, err := j.call(req)
	if err != nil || resp.StatusCode != 200 {
		log.Println("Error Occurred: No Issues Found")
		return []Issue{}
	}
	defer resp.Body.Close()

	var issueJqlResp IssueJqlResp
	err = json.NewDecoder(resp.Body).Decode(&issueJqlResp)
	if err != nil {
		log.Fatal(err)
	}

	issuesJql := issueJqlResp.Issues
	issuesLen := math.Min(5, float64(len(issuesJql)))
	issues := make([]Issue, int(issuesLen))
	for i, issueJql := range issuesJql {
		if i == 5 {
			break
		}
		issues[i] = Issue{
			SummaryText: issueJql.Fields.Summary,
			Key:         issueJql.Key,
		}
	}

	return issues[:]
}

func loadIssueListPrompt(issues []Issue) string {
	if len(issues) == 0 {
		return ""
	}

	// Create a prompt for JIRA Link
	descriptions := make(map[string]string)
	issuesLen := math.Min(5, float64(len(issues)))
	options := make([]string, int(issuesLen))

	for i, issue := range issues {
		if i == 5 {
			break
		}
		issueKey := issue.Key
		options[i] = issueKey
		descriptions[issueKey] = issue.SummaryText
	}

	issueKey := ""
	prompt := &survey.Select{
		Message: "Select Issue:",
		Options: options[:],
		Description: func(value string, index int) string {
			return descriptions[value]
		},
	}
	err := survey.AskOne(prompt, &issueKey)

	if err != nil {
		log.Fatal(err)
	}

	return issueKey
}

func openbrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func getUserInput() string {
	var userInput = ""
	prompt := &survey.Input{
		Renderer: survey.Renderer{},
		Message:  "Enter Query Here: ",
	}
	survey.AskOne(prompt, &userInput)
	return userInput
}
