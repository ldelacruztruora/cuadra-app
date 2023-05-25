variable "region" {
   type    = string
   default = "us-east-1"
}

variable "name" {
  type = string
}

variable "description" {
  type = string
}

variable "account_id" {
  type = string
}

variable "type" {
  type = string
}

variable "binary_media_types" {
  type    = list(string)
  default = []
}

variable "path" {
  type = string
}

variable "service" {
  type = string
}

variable "stage_name" {
  type = string
}

variable "metrics_enabled" {
  type = bool
}