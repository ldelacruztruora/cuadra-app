resource "aws_api_gateway_resource" "v1" {
  rest_api_id = var.rest_api.id
  parent_id   = var.rest_api.root_resource_id
  path_part   = var.path
}


/*
  METHOD: OPTIONS
  PATH: /v1
  TYPE INTEGRATION: MOCK
*/
module "options" {
  source = "../../../../modules/apigateway/base-method"

  config           = local.config
  resource_id      = aws_api_gateway_resource.v1.id
  resource_path    = aws_api_gateway_resource.v1.path
  request_method   = "OPTIONS"
  type_integration = "MOCK"

  request_templates               = local.config.request_templates
  response_templates              = local.config.response_templates
  response_parameters             = local.config.response_parameters
  response_integration_parameters = local.config.response_integration_parameters
}

/*
  METHOD: GET
  PATH: /v1
  TYPE INTEGRATION: MOCK
*/
module "get" {
  source = "../../../../modules/apigateway/base-method"

  config = {
    region      = var.region
    account_id  = var.account_id
    rest_api_id = var.rest_api.id
  }

  authorization    = "NONE"
  resource_id      = var.rest_api.root_resource_id
  resource_path    = "/"
  request_method   = "GET"
  type_integration = "MOCK"

  request_templates = {
    "application/json" = "{ \"statusCode\": 200 }"
  }

  response_templates = {
    "application/json" = ""
  }

  response_parameters             = local.config.response_parameters
  response_integration_parameters = local.config.response_integration_parameters
}
