variable "preview_name" {
  type        = string
  description = "The preview environment's name"
}

variable "kubeconfig_path" {
  type        = string
  default     = "/home/gitpod/.kube/config"
  description = "The path to the kubernetes config"
}

variable "harvester_kube_context" {
  type        = string
  default     = "harvester"
  description = "The name of the harvester kube context"
}

variable "dev_kube_context" {
  type        = string
  default     = "dev"
  description = "The name of the dev kube context"
}

variable "vmi" {
  type        = string
  description = "The storage class for the VM"
  default     = "gitpod-k3s-202209251218"
}

variable "cert_issuer" {
  type        = string
  default     = "zerossl-issuer-gitpod-core-dev"
  description = "Certificate issuer"
}
