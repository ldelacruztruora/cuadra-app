# Find a certificate that is issued
data "aws_acm_certificate" "cert" {
  domain      = length(var.cert_name) == 0 ? var.domain_name : var.cert_name
  statuses    = ["ISSUED"]
  most_recent = true
}

# The domain name to use with api-gateway
resource "aws_api_gateway_domain_name" "api_gateway_domain_name" {
  domain_name     = var.domain_name
  certificate_arn = data.aws_acm_certificate.cert.arn
  security_policy = var.security_policy
}

resource "aws_api_gateway_base_path_mapping" "staging_base_path_mapping" {
  api_id      = var.api_id
  stage_name  = var.stage_name
  domain_name = aws_api_gateway_domain_name.api_gateway_domain_name.domain_name
}