workflow "tests" {
  on = "push"
  resolves = ["let's try some docker stuff"]
}

action "let's try some docker stuff" {
  uses = "docker://alpine:latest"
  runs = "ls"
  args = "-ltr"
}
