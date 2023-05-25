locals {
  config = {
    region      = var.region
    account_id  = var.account_id
    rest_api_id = var.rest_api.id

    response_parameters = {
      "method.response.header.Access-Control-Allow-Headers" = true,
      "method.response.header.Access-Control-Allow-Methods" = true,
      "method.response.header.Access-Control-Allow-Origin"  = true,
      "method.response.header.Strict-Transport-Security"    = true
    }

    response_integration_parameters = {
      "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Amz-Date,Truora-API-Key,X-Api-Key,x-requested-with,Truora-Client,Truora-Priority'",
      "method.response.header.Access-Control-Allow-Methods" = "'GET,POST,OPTIONS'",
      "method.response.header.Access-Control-Allow-Origin"  = "'*'",
      "method.response.header.Strict-Transport-Security"    = "'max-age=63072000; includeSubdomains; preload'"
    }

    request_templates = {
      "application/json" = "{ \"statusCode\": 200 }"
    }

    response_templates = {
      "application/json" = ""
    }
  }
}