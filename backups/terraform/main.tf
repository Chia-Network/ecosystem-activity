provider "aws" {
  region = "us-west-2"
}

terraform {
  backend "s3" {
    bucket               = "chia-terraform"
    key                  = "ecosystem-s3-backup/terraform.tfstate"
    region               = "us-west-2"
  }
}

resource "aws_s3_bucket" "chia-ecosystem-github-repo-backups" {
  bucket = "chia-ecosystem-github-repo-backups"
  acl    = "private"

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm     = "aws:kms"
      }
    }
  }

  tags = {
    Name = "Chia Ecosystem Repository Backups"
  }
}

resource "aws_s3_bucket_public_access_block" "chia-ecosystem-github-repo-backups-block-public" {
  bucket = aws_s3_bucket.chia-ecosystem-github-repo-backups.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
