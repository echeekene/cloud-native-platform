terraform {
  required_version = ">= 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.80"
    }
  }

  # Partial config: bucket and key come from backend.hcl so nothing
  # account specific is committed. S3 native locking needs TF 1.10+.
  backend "s3" {
    use_lockfile = true
    encrypt      = true
  }
}
