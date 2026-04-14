terraform {
  backend "gcs" {
    bucket = "my-org-shared-tfstate"
    prefix = "foobar-service"
  }
}
