# Bring the hand built portfolio bucket under Terraform without
# recreating it. Import blocks need TF 1.7+; with manage_site_bucket = false
# nothing here is created or imported.

locals {
  site_buckets = var.manage_site_bucket ? toset([var.site_bucket_name]) : toset([])
}

import {
  for_each = local.site_buckets
  to       = aws_s3_bucket.site[each.key]
  id       = each.key
}

import {
  for_each = local.site_buckets
  to       = aws_s3_bucket_public_access_block.site[each.key]
  id       = each.key
}

import {
  for_each = local.site_buckets
  to       = aws_s3_bucket_server_side_encryption_configuration.site[each.key]
  id       = each.key
}

resource "aws_s3_bucket" "site" {
  for_each = local.site_buckets
  bucket   = each.key

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_public_access_block" "site" {
  for_each                = local.site_buckets
  bucket                  = aws_s3_bucket.site[each.key].id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "site" {
  for_each = local.site_buckets
  bucket   = aws_s3_bucket.site[each.key].id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# The distribution stays managed in the console for now; read it so its
# domain shows up in outputs.
data "aws_cloudfront_distribution" "site" {
  count = var.site_distribution_id == "" ? 0 : 1
  id    = var.site_distribution_id
}
