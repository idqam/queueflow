variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "project_name" {
  type    = string
  default = "queueflow"
}

variable "cluster_version" {
  type    = string
  default = "1.29"
}

variable "node_instance_type" {
  type    = string
  default = "t3.medium"
}

variable "node_min" {
  type    = number
  default = 2
}

variable "node_max" {
  type    = number
  default = 10
}

variable "node_desired" {
  type    = number
  default = 2
}

variable "db_instance_class" {
  type    = string
  default = "db.t3.medium"
}

variable "db_name" {
  type    = string
  default = "queueflow"
}

variable "db_username" {
  type    = string
  default = "queueflow"
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "kafka_version" {
  type    = string
  default = "3.5.1"
}

variable "kafka_broker_instance_type" {
  type    = string
  default = "kafka.t3.small"
}

variable "kafka_broker_count" {
  type    = number
  default = 3
}

variable "redis_node_type" {
  type    = string
  default = "cache.t3.micro"
}
