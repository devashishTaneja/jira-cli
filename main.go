package main

import (
	"github.com/AlecAivazis/survey/v2"
	"log"
	"os"
)

func executePrompt() {
	option := ""
	prompt := &survey.Select{
		Message: "Select option:",
		Options: []string{"My Issues", "Search Issue", "Advanced Search"},
	}
	err := survey.AskOne(prompt, &option)
	if err != nil {
		return
	}

	jiraClient := JiraClient{
		Domain: os.Getenv("JIRA_DOMAIN"),
		Credential: Credential{
			ApiUser: os.Getenv("JIRA_API_USER"),
			ApiKey:  os.Getenv("JIRA_API_KEY"),
		},
	}

	switch option {
	case "My Issues":
		jiraClient.advancedJqlSearch("assignee in (currentUser()) ORDER BY created desc")
	case "Search Issue":
		jiraClient.search()
	case "Advanced Search":
		jiraClient.advancedJqlSearch("")
	default:
		log.Fatal("Invalid option")
	}
}

func main() {
	executePrompt()
}
