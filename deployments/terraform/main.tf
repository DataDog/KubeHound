module "khaas" {
  source = "./create-s3"

  bucket_name = var.bucket_name
}
