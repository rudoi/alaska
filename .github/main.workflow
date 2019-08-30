workflow "tests" {
  on = "push"
  resolves = ["let's try some docker stuff"]
}

action "let's try some docker stuff" {
  uses = "docker://alpine:latest"
  runs = "sh"
  args = "-c \"ls -ltr\""
}
