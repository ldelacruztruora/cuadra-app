resource "aws_api_gateway_rest_api" "api" {
  name               = var.name
  description        = var.description
  binary_media_types = var.binary_media_types
  tags               = var.tags
  endpoint_configuration {
    types = [var.type]
  }
}

resource "aws_api_gateway_deployment" "deployment" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  description = var.description

  variables = var.deployment_variables

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_method_settings" "settings" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = aws_api_gateway_stage.stage.stage_name
  method_path = "*/*"

  settings {
    metrics_enabled = var.metrics_enabled
    logging_level   = var.logging_level
  }
}

resource "aws_api_gateway_stage" "stage" {
  deployment_id = aws_api_gateway_deployment.deployment.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
  stage_name    = var.stage_name

  cache_cluster_size = var.cache_cluster_size
}

resource "aws_api_gateway_gateway_response" "custom_response" {
  count         = length(var.default_responses)
  rest_api_id   = aws_api_gateway_rest_api.api.id
  status_code   = var.default_responses[count.index].status_code
  response_type = var.default_responses[count.index].response_type

  response_templates  = var.default_responses[count.index].response_templates
  response_parameters = var.default_responses[count.index].response_parameters
}
