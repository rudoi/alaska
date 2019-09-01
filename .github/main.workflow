workflow "tests" {
  on = "push"
  resolves = ["let's try some docker stuff"]
}

action "let's try some docker stuff" {
  uses = "docker://golang:1.12.7"
  runs = "sh"
  args = "-c 'go mod download && go test -v ./pkg/...'"
}
