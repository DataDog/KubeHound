resource "aws_s3_bucket" "khaas_s3_instance" {
  bucket = var.bucket_name
}

resource "aws_s3_bucket_acl" "khaas_s3_acl" {
  bucket = aws_s3_bucket.khaas_s3_instance.id
  acl    = "private"
}

// Reference: https://aws.amazon.com/blogs/aws/heads-up-amazon-s3-security-changes-are-coming-in-april-of-2023/
resource "aws_s3_bucket_ownership_controls" "khaas_s3_ownership" {
  bucket = aws_s3_bucket.khaas_s3_instance.id

  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_public_access_block" "khaas_s3_public_access_block" {
  bucket                  = aws_s3_bucket.khaas_s3_instance.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "khaas_s3_server_side_encryption_configuration" {
  bucket = aws_s3_bucket.khaas_s3_instance.id
  rule {
    bucket_key_enabled = false
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
