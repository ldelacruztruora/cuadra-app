resource "aws_api_gateway_method" "request_method" {
  http_method        = var.http_method
  rest_api_id        = var.api_id
  resource_id        = var.resource_id
  authorization      = var.authorization
  authorizer_id      = var.authorizer_id
  request_parameters = var.request_parameters
}

resource "aws_api_gateway_integration" "request_method_integration" {
  count = var.type_integration == "lambda" ? 1 : 0

  rest_api_id             = var.api_id
  resource_id             = var.resource_id
  http_method             = aws_api_gateway_method.request_method.http_method
  integration_http_method = var.integration_http_method
  type                    = var.type_request_integration
  uri                     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/arn:aws:lambda:${var.region}:${var.account_id}:function:${var.lambda_name}:${var.lambda_alias}/invocations"
}

resource "aws_lambda_permission" "allow_api_gateway" {
  count = var.type_integration == "lambda" ? 1 : 0

  function_name = var.lambda_name
  qualifier     = var.lambda_alias
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "arn:aws:execute-api:${var.region}:${var.account_id}:${var.api_id}/*/${aws_api_gateway_method.request_method.http_method}${var.resource_path}"
}