provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project   = "cloud-native-platform"
      Owner     = "ekene-eche"
      ManagedBy = "terraform"
      Repo      = "echeekene/cloud-native-platform"
    }
  }
}
