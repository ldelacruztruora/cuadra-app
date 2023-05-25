module "rest_api"  {
   source = "../../../modules/apigateway/rest-api"

   name               = var.name
   type               = var.type
   description        = var.description
   stage_name         = var.stage_name
   binary_media_types = var.binary_media_types
   metrics_enabled    = var.metrics_enabled
   logging_level      = "ERROR"
   cache_cluster_size = "0.5"

   deployment_variables = {
     "version" = "0.1"
   }
}

module "resources" {
  source = "./resources/"

  rest_api   = module.rest_api.api
  path       = var.path
  region     = var.region
  account_id = var.account_id
}