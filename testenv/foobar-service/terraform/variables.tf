variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "us-east1"
}

variable "image_repository" {
  type    = string
  default = "us-east1-docker.pkg.dev/my-org-shared/my-org-images"
}
