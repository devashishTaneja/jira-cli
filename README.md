go build -o out/jira main.go
out/jira <search_text>


[//]: # (Build steps)
go mod init github.com/devashishTaneja/jira-cli
go mod vendor

[//]: # (Go Releaser Setup)