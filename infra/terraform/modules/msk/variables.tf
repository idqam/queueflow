variable "project_name" {
  type = string
}

variable "kafka_version" {
  type = string
}

variable "broker_count" {
  type = number
}

variable "broker_instance_type" {
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
