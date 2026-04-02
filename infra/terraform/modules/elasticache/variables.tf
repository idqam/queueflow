variable "project_name" {
  type = string
}

variable "node_type" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

variable "vpc_id" {
  type = string
}

variable "allowed_sg_id" {
  type = string
}
