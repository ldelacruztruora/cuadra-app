
// request
resource "aws_api_gateway_method" "request_method" {
  count = var.request_method != "" ? 1 : 0

  rest_api_id        = var.config.rest_api_id
  authorizer_id      = var.request_method == "OPTIONS" || var.authorization == "NONE" ? "" : var.config.authorizer_id
  authorization      = var.request_method == "OPTIONS" ? "NONE" : var.authorization
  resource_id        = var.resource_id
  http_method        = var.request_method
  request_parameters = var.request_parameters
}

resource "aws_api_gateway_integration" "request_method_integration" {
  count = var.lambda_name != "" ? 1 : 0

  rest_api_id             = var.config.rest_api_id
  resource_id             = var.resource_id
  http_method             = aws_api_gateway_method.request_method[0].http_method
  integration_http_method = var.integration_http_method
  type                    = var.type_integration
  uri                     = "arn:aws:apigateway:${var.config.region}:lambda:path/2015-03-31/functions/arn:aws:lambda:${var.config.region}:${var.config.account_id}:function:${var.lambda_name}:${var.lambda_alias}/invocations"
}

resource "aws_lambda_permission" "allow_api_gateway" {
  count = var.lambda_name != "" ? 1 : 0

  function_name = var.lambda_name
  qualifier     = var.lambda_alias
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "arn:aws:execute-api:${var.config.region}:${var.config.account_id}:${var.config.rest_api_id}/*/${aws_api_gateway_method.request_method[0].http_method}${var.resource_path}"
}

// response
resource "aws_api_gateway_method_response" "method_response" {
  rest_api_id         = var.config.rest_api_id
  resource_id         = var.resource_id
  http_method         = aws_api_gateway_method.request_method[0].http_method
  status_code         = var.status_code
  response_parameters = var.response_parameters
  response_models     = var.response_models
}

resource "aws_api_gateway_integration" "integration" {
  count = var.type_integration == "MOCK" ? 1 : 0

  rest_api_id = var.config.rest_api_id
  resource_id = var.resource_id
  http_method = aws_api_gateway_method.request_method[0].http_method
  type        = "MOCK"

  request_templates    = var.request_templates
  passthrough_behavior = var.passthrough_behavior
}

resource "aws_api_gateway_integration_response" "integration_response" {
  rest_api_id         = var.config.rest_api_id
  resource_id         = var.resource_id
  http_method         = aws_api_gateway_method.request_method[0].http_method
  status_code         = aws_api_gateway_method_response.method_response.status_code
  response_parameters = var.response_integration_parameters
  response_templates  = var.response_templates
}