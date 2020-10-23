project = "Example TIG deployment"

app "TIG" {
  labels = {
    "service" = "tig",
    "env" = "dev"
  }

  // This doesn't actually do anything
  build {
    use "docker-pull" {
      image = "alpine"
      tag = "latest"
    }
  }

  deploy { 
    use "bolt" {
      plan = "facts"
      targets = ["localhost"]
    }
  }
}
