variable "api_id" {
  type = string
}

variable "resource_id" {
  type = string
}

variable "http_method" {
  type = string
}

variable "status_code" {
  default = null
}

variable "response_parameters" {
  default = {}
}

variable "response_models" {
  default = {}
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