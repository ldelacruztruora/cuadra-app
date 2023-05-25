resource "aws_api_gateway_method_response" "method_response" {
  rest_api_id         = var.api_id
  resource_id         = var.resource_id
  http_method         = var.http_method
  status_code         = var.status_code
  response_parameters = var.response_parameters
  response_models     = var.response_models
}

resource "aws_api_gateway_integration" "integration" {
  count = var.http_method == "OPTIONS" ? 1 : 0

  rest_api_id = var.api_id
  resource_id = var.resource_id
  http_method = var.http_method
  type        = "MOCK"

  request_templates    = var.request_templates
  passthrough_behavior = var.passthrough_behavior
}

resource "aws_api_gateway_integration_response" "integration_response" {
  rest_api_id         = var.api_id
  resource_id         = var.resource_id
  http_method         = aws_api_gateway_method_response.method_response.http_method
  status_code         = aws_api_gateway_method_response.method_response.status_code
  response_parameters = var.response_integration_parameters
  response_templates  = var.response_templates
}