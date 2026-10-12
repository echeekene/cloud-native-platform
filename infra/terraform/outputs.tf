output "node_public_ip" {
  description = "Point status.ekeneeche.com at this in Cloudflare"
  value       = aws_eip.node.public_ip
}

output "node_instance_id" {
  description = "Set as the PLATFORM_INSTANCE_ID repo variable"
  value       = aws_instance.node.id
}

output "github_deploy_role_arn" {
  description = "Set as the PLATFORM_DEPLOY_ROLE_ARN repo variable"
  value       = aws_iam_role.github_deploy.arn
}

output "ssh_command" {
  value = "ssh ubuntu@${aws_eip.node.public_ip}"
}

output "ansible_inventory" {
  description = "Paste into infra/ansible/inventory.ini"
  value       = "[platform]\nnode ansible_host=${aws_eip.node.public_ip} ansible_user=ubuntu\n"
}

output "site_cloudfront_domain" {
  value = try(data.aws_cloudfront_distribution.site[0].domain_name, null)
}

output "account_id" {
  value = data.aws_caller_identity.current.account_id
}
