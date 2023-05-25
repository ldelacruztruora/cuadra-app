variable "domain_name" {
  type = string
}

variable "api_id" {
  type = string
}

variable "stage_name" {
  type = string
}

variable "cert_name" {
  default = ""
}

variable "security_policy" {
  type    = string
  default = "TLS_1_2"
}


variable "tags" {}
