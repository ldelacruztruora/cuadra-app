output "api" {
  value = aws_api_gateway_rest_api.api
}

output "stage" {
  value = aws_api_gateway_stage.stage
}

output "deployment" {
  value = aws_api_gateway_deployment.deployment
}
