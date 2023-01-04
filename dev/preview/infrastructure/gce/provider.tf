terraform {

  backend "gcs" {
    bucket = "3f4745df-preview-tf-state"
    prefix = "preview-gce"
  }

  required_version = ">= 1.2"
  required_providers {
    k8s = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.0"
    }
    google = {
      source  = "hashicorp/google"
      version = ">=4.47.0"
    }
  }
}
provider "k8s" {
  alias          = "dev"
  config_path    = var.kubeconfig_path
  config_context = var.dev_kube_context
}

provider "k8s" {
  alias          = "harvester"
  config_path    = var.kubeconfig_path
  config_context = var.harvester_kube_context
}

provider "google" {
  project = "gitpod-core-dev"
  region  = "us-central1"
}
