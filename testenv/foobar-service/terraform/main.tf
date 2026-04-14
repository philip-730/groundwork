terraform {
  required_version = ">= 1.5"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

data "google_project" "env" {}

resource "google_service_account" "service" {
  account_id   = "foobar-service"
  display_name = "foobar-service service account"
}

resource "google_artifact_registry_repository_iam_member" "service_reader" {
  location   = "us-east1"
  repository = "my-org-images"
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.service.email}"
}

resource "google_artifact_registry_repository_iam_member" "cloudbuild_writer" {
  location   = "us-east1"
  repository = "my-org-images"
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${data.google_project.env.number}@cloudbuild.gserviceaccount.com"
}

resource "google_cloud_run_v2_service" "service" {
  name     = "foobar-service"
  location = var.region

  template {
    service_account = google_service_account.service.email

    containers {
      image = "${var.image_repository}/foobar-service:latest"

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }
  }

  lifecycle {
    ignore_changes = [
      template[0].containers[0].image,
    ]
  }

  depends_on = [
    google_artifact_registry_repository_iam_member.service_reader,
  ]
}
