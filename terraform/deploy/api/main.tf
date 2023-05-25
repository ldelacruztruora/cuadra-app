terraform {
   required_providers {
     aws = {
      source = "hashicorp/aws"
      version = "~> 4.67"
     }
   }
}

provider "aws" {
  region = var.region
}

module "apigateway" {
  source = "./apigateway"

  name               = var.name
  region             = var.region
  description        = var.description
  account_id            = var.account_id
  stage_name         = var.stage_name
  type               = var.type
  binary_media_types = var.binary_media_types
  path               = var.path
  service            = var.service
  metrics_enabled    = var.metrics_enabled
}