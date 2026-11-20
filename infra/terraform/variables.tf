variable "region" {
  type        = string
  default     = "us-east-1"
  description = "AWS region for the platform (matches the portfolio site)"
}

variable "name" {
  type        = string
  default     = "platform"
  description = "Prefix for resource names"
}

variable "vpc_cidr" {
  type    = string
  default = "10.20.0.0/16"
}

variable "public_subnet_cidr" {
  type    = string
  default = "10.20.1.0/24"
}

variable "node_private_ip" {
  type        = string
  default     = "10.20.1.10"
  description = "Fixed private IP for the node, matches the records in the internal DNS zone"
}

variable "instance_type" {
  type        = string
  default     = "t3.medium"
  description = "4 GiB of memory fits k3s plus Prometheus and Grafana"
}

variable "admin_cidrs" {
  type        = list(string)
  description = "Your public IP as a /32. Only these addresses reach SSH, the Kubernetes API, DNS and NTP."

  validation {
    condition     = alltrue([for c in var.admin_cidrs : can(cidrhost(c, 0)) && c != "0.0.0.0/0"])
    error_message = "admin_cidrs must be valid CIDRs and must not be 0.0.0.0/0."
  }
}

variable "ssh_public_key" {
  type        = string
  description = "Public key Ansible uses to log in"
}

variable "github_repo" {
  type        = string
  default     = "echeekene/cloud-native-platform"
  description = "owner/repo allowed to assume the deploy role"
}

variable "alert_email" {
  type        = string
  default     = ""
  description = "Email for instance health alarms. Leave empty to skip."
}

# Existing portfolio site. These let Terraform adopt the bucket that was
# originally created by hand in the console.
variable "manage_site_bucket" {
  type        = bool
  default     = false
  description = "Import and manage the existing portfolio S3 bucket"
}

variable "site_bucket_name" {
  type    = string
  default = ""
}

variable "site_distribution_id" {
  type        = string
  default     = ""
  description = "Existing CloudFront distribution ID, read only"
}
