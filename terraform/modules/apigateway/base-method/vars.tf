variable "resource_id" {
  type = string
}

variable "resource_path" {
  type = string
}

variable "config" {}

variable "authorization" {
  default = "CUSTOM"
}

variable "authorization_none" {
  default = "NONE"
}

variable "integration_http_method" {
  default = "POST"
}

variable "lambda_name" {
  default = ""
}

variable "lambda_alias" {
  default = "current"
}

variable "request_method" {
  default = ""
}

variable "type_integration" {
  default = "AWS_PROXY"
}

variable "request_parameters" {
  default = {}
}

variable "response_parameters" {
  default = {}
}

variable "response_models" {
  default = null
}

variable "response_integration_parameters" {
  default = {}
}

variable "response_templates" {
  default = {}
}

variable "request_templates" {
  default = {}
}

variable "passthrough_behavior" {
  default = "WHEN_NO_MATCH"
}

variable "status_code" {
  type    = number
  default = 200
}

variable "tags" {
  type    = map(string)
  default = {}
}