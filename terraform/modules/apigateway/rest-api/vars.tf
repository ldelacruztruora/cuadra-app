variable "tags" {
  type    = map(string)
  default = {}
}

variable "name" {
  type = string
}

variable "description" {
  type = string
}

variable "type" {
  type = string
}

variable "binary_media_types" {
  type    = list(string)
  default = []
}

variable "stage_name" {
  type = string
}

variable "cache_cluster_size" {
  type = string
}

variable "metrics_enabled" {
  type = bool
}

variable "logging_level" {
  type = string
}

variable "deployment_variables" {
  type = map(string)
}

variable "default_responses" {
  type = list(object({
    status_code         = optional(string)
    response_type       = string
    response_templates  = map(string)
    response_parameters = optional(map(string))
  }))

  default = []
}