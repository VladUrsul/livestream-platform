variable "do_token" {
  description = "DigitalOcean API token"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "DigitalOcean region"
  type        = string
  default     = "fra1"  # Frankfurt
}

variable "cluster_name" {
  description = "Kubernetes cluster name"
  type        = string
  default     = "streamr-cluster"
}

variable "node_size" {
  description = "Droplet size for cluster nodes"
  type        = string
  default     = "s-2vcpu-4gb" 
}

variable "node_count" {
  description = "Number of cluster nodes"
  type        = number
  default     = 2
}

variable "db_size" {
  description = "Managed database node size"
  type        = string
  default     = "db-s-1vcpu-1gb"  # cheapest managed DB
}

variable "environment" {
  type    = string
  default = "production"
}