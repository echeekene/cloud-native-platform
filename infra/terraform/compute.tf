# Latest Ubuntu 24.04 LTS from Canonical
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_key_pair" "admin" {
  key_name   = "${var.name}-admin"
  public_key = var.ssh_public_key
}

resource "aws_instance" "node" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.public.id
  private_ip             = var.node_private_ip
  vpc_security_group_ids = [aws_security_group.node.id]
  iam_instance_profile   = aws_iam_instance_profile.node.name
  key_name               = aws_key_pair.admin.key_name
  monitoring             = true

  # IMDSv2 only: blocks SSRF style credential theft from the metadata service
  metadata_options {
    http_tokens                 = "required"
    http_put_response_hop_limit = 2 # pods need one extra hop
    http_endpoint               = "enabled"
  }

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 30
    encrypted             = true
    delete_on_termination = true
  }

  # Ansible does the real configuration; this only makes sure the box
  # is patched and has Python before Ansible connects.
  user_data = <<-EOT
    #!/bin/bash
    set -euo pipefail
    apt-get update -y
    DEBIAN_FRONTEND=noninteractive apt-get upgrade -y
    apt-get install -y python3
  EOT

  lifecycle {
    ignore_changes = [ami] # do not rebuild the node every time Canonical ships an AMI
  }

  tags = { Name = "${var.name}-node" }
}

resource "aws_eip" "node" {
  domain   = "vpc"
  instance = aws_instance.node.id
  tags     = { Name = "${var.name}-node" }

  depends_on = [aws_internet_gateway.main]
}
