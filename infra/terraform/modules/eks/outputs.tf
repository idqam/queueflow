output "cluster_endpoint" {
  value = aws_eks_cluster.main.endpoint
}

output "cluster_name" {
  value = aws_eks_cluster.main.name
}

output "cluster_ca" {
  value = aws_eks_cluster.main.certificate_authority[0].data
}

output "cluster_token" {
  value     = data.aws_eks_cluster_auth.main.token
  sensitive = true
}

output "vpc_id" {
  value = aws_vpc.main.id
}

output "private_subnet_ids" {
  value = aws_subnet.private[*].id
}

output "node_security_group_id" {
  value = aws_eks_node_group.main.resources[0].remote_access_security_group_id
}
