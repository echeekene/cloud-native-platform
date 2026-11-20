# Web traffic is public. DNS and NTP are limited to admin_cidrs on purpose:
# an open resolver or open NTP server on the internet gets abused for
# reflection and amplification DDoS within hours.

resource "aws_security_group" "node" {
  name        = "${var.name}-node"
  description = "k3s node running the DNS, NTP and status services"
  vpc_id      = aws_vpc.main.id

  tags = { Name = "${var.name}-node" }
}

resource "aws_vpc_security_group_ingress_rule" "http" {
  security_group_id = aws_security_group.node.id
  description       = "HTTP, redirected to HTTPS by Traefik"
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = 80
  to_port           = 80
}

resource "aws_vpc_security_group_ingress_rule" "https" {
  security_group_id = aws_security_group.node.id
  description       = "HTTPS status page"
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = 443
  to_port           = 443
}

locals {
  admin_ports = {
    ssh     = { protocol = "tcp", port = 22, description = "SSH for Ansible" }
    k8s_api = { protocol = "tcp", port = 6443, description = "Kubernetes API" }
    dns_udp = { protocol = "udp", port = 53, description = "DNS over UDP" }
    dns_tcp = { protocol = "tcp", port = 53, description = "DNS over TCP" }
    ntp     = { protocol = "udp", port = 123, description = "NTP" }
  }

  admin_rules = {
    for pair in setproduct(keys(local.admin_ports), var.admin_cidrs) :
    "${pair[0]}-${pair[1]}" => merge(local.admin_ports[pair[0]], { cidr = pair[1] })
  }
}

resource "aws_vpc_security_group_ingress_rule" "admin" {
  for_each = local.admin_rules

  security_group_id = aws_security_group.node.id
  description       = each.value.description
  cidr_ipv4         = each.value.cidr
  ip_protocol       = each.value.protocol
  from_port         = each.value.port
  to_port           = each.value.port
}

resource "aws_vpc_security_group_ingress_rule" "vpc_internal" {
  security_group_id = aws_security_group.node.id
  description       = "Anything from inside the VPC"
  cidr_ipv4         = var.vpc_cidr
  ip_protocol       = "-1"
}

resource "aws_vpc_security_group_egress_rule" "all" {
  security_group_id = aws_security_group.node.id
  description       = "Outbound for updates, image pulls, DNS over TLS and NTP upstreams"
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}
