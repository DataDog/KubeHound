# Deploying the cloud storage provider

## Deploying KHaaS bucket in AWS s3

Execute the following commands:

```bash
terraform init
terraform plan
terraform apply
```

> [!TIP]
> If you want to avoid the interactive prompt, create a `terraform.tfvars` file with the bucket name (i.e. `bucket=<your_bucket>`).

We advise you to use Terraform S3 and dynamoDB backend, [more info here](https://developer.hashicorp.com/terraform/language/settings/backends/s3).

```tf
  backend "s3" {
    key            = "terraform-state"
    region         = "us-east-1"
    bucket         = "<your_bucket>"
    encrypt        = true # Optional, S3 Bucket Server Side Encryption
    dynamodb_table = "<your_bucket>-terraform-state-lock"
  }
  ```

