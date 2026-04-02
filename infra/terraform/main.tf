module "eks" {
  source = "./modules/eks"

  project_name       = var.project_name
  cluster_version    = var.cluster_version
  node_instance_type = var.node_instance_type
  node_min           = var.node_min
  node_max           = var.node_max
  node_desired       = var.node_desired
}

module "rds" {
  source = "./modules/rds"

  project_name   = var.project_name
  engine_version = "16.4"
  instance_class = var.db_instance_class
  db_name        = var.db_name
  db_username    = var.db_username
  db_password    = var.db_password
  subnet_ids     = module.eks.private_subnet_ids
  vpc_id         = module.eks.vpc_id
  allowed_sg_id  = module.eks.node_security_group_id
}

module "msk" {
  source = "./modules/msk"

  project_name         = var.project_name
  kafka_version        = var.kafka_version
  broker_count         = var.kafka_broker_count
  broker_instance_type = var.kafka_broker_instance_type
  subnet_ids           = module.eks.private_subnet_ids
  vpc_id               = module.eks.vpc_id
  allowed_sg_id        = module.eks.node_security_group_id
}

module "elasticache" {
  source = "./modules/elasticache"

  project_name  = var.project_name
  node_type     = var.redis_node_type
  subnet_ids    = module.eks.private_subnet_ids
  vpc_id        = module.eks.vpc_id
  allowed_sg_id = module.eks.node_security_group_id
}

resource "helm_release" "prometheus" {
  name             = "prometheus"
  repository       = "https://prometheus-community.github.io/helm-charts"
  chart            = "kube-prometheus-stack"
  namespace        = "monitoring"
  create_namespace = true

  set {
    name  = "grafana.adminPassword"
    value = "admin"
  }
}
