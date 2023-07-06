package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/savioxavier/termlink"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

type issue struct {
	SummaryText, Key string
}

type section struct {
	Label, Sub, Id string
	Issues         []issue
}

type issuePickerResp struct {
	Sections []section
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

func main() {

	args := os.Args

	if len(args) < 2 {
		log.Fatal("Missing Arguments")
		os.Exit(1)
	}

	jiraDomain := os.Getenv("JIRA_DOMAIN")
	username := os.Getenv("JIRA_API_USER")
	apikey := os.Getenv("JIRA_API_KEY")
	base64Auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(username + ":" + apikey))

	queryUrl := jiraDomain + "/rest/api/3/issue/picker?query=" + os.Args[1]

	client := http.Client{}
	req, err := http.NewRequest("GET", queryUrl, nil)
	req.Header.Add("Authorization", base64Auth)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != 200 {
		fmt.Println("Error Occured: ", resp.StatusCode)
		os.Exit(1)
	}

	defer resp.Body.Close()
	var issuePickerResp issuePickerResp
	err = json.NewDecoder(resp.Body).Decode(&issuePickerResp)

	issues := issuePickerResp.Sections[0].Issues
	len := len(issues)
	if len > 5 {
		len = 5
	}

	if len > 1 {
		for i := 0; i < len; i++ {
			fmt.Println("______________________________________________")
			issue := issues[i]
			fmt.Println(issue.SummaryText)
			jiraLink := jiraDomain + "/browse/" + issue.Key
			fmt.Println(termlink.ColorLink(issue.Key, jiraLink, "red"))

		}
	} else if len == 0 {
		fmt.Println("No related issues found")
	} else {
		openbrowser(jiraDomain + "/browse/" + issues[0].Key)
	}

}
