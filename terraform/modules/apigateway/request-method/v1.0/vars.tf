# Optional
variable "type_request_integration" {
  default = "AWS"
}

variable "http_method" {
  default = "GET"
}

variable "integration_http_method" {
  type = string
}

# Required
variable "region" {
  type = string
}

variable "account_id" {
  type = string
}

variable "api_id" {
  type        = string
  description = "The ID of the associated REST API"
}

variable "resource_id" {
  type        = string
  description = "The API resource ID"
}

variable "authorization" {
  type    = string
  default = "NONE"
}

variable "resource_path" {
  type        = string
  description = "The API resource path"
}

variable "type_integration" {
  type = string
}

variable "type_integration_mock" {
  default = "MOCK"
}

variable "authorizer_id" {
  type    = string
  default = ""
}

variable "querystring" {
  type    = map(any)
  default = {}
}

variable "request_parameters" {
  type    = map(any)
  default = {}
}

variable "lambda_name" {
  default = "function"
}

variable "lambda_alias" {
  default = "current"
}