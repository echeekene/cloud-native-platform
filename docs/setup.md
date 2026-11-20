# Setup

## Prerequisites

Terraform 1.10+, Ansible, AWS CLI signed in to the account, Docker, Go 1.24, and `dig`.

## 1. State bucket (once)

```bash
cd infra/terraform/bootstrap
terraform init
terraform apply -var state_bucket_name=<unique-name>
```

## 2. Infrastructure

```bash
cd infra/terraform
cp backend.hcl.example backend.hcl            # fill in the bucket
cp terraform.tfvars.example terraform.tfvars  # fill in your IP, SSH key, site bucket
terraform init -backend-config=backend.hcl
terraform plan
terraform apply
```

With `manage_site_bucket = true`, the plan should show the portfolio bucket being **imported**, not created or destroyed. Stop and check if it says otherwise.

## 3. DNS and TLS in Cloudflare

1. Add an `A` record `status` pointing to the `node_public_ip` output, proxied (orange cloud).
2. SSL/TLS mode: Full (strict).
3. Create an Origin Certificate for `status.ekeneeche.com` and save the cert and key.
4. After k3s is up (step 5), store it in the cluster:

```bash
kubectl -n platform create secret tls status-origin-tls --cert origin.pem --key origin.key
```

## 4. GitHub

Repository variables:

| Name | Value |
|---|---|
| `PLATFORM_DEPLOY_ROLE_ARN` | `github_deploy_role_arn` output |
| `PLATFORM_INSTANCE_ID` | `node_instance_id` output |

Create a `production` environment with yourself as required reviewer, so every deploy waits for a click.

Push to `main` once so images exist in GHCR, then make the `status-api` and `chrony` packages public (or add an image pull secret).

## 5. Configure the node

```bash
cd infra/ansible
terraform -chdir=../terraform output -raw ansible_inventory > inventory.ini
ansible-galaxy collection install -r requirements.yml
ansible-playbook site.yml
```

Create the Grafana admin password:

```bash
kubectl -n monitoring create secret generic grafana-admin --from-literal=password="$(openssl rand -base64 24)"
```

## 6. Check it

```bash
./scripts/smoke-test.sh <node_public_ip>
kubectl -n monitoring port-forward svc/grafana 3000:3000
```

## Tear down

```bash
terraform destroy
```

The portfolio bucket has `prevent_destroy`, so remove it from state first if it is managed: `terraform state rm 'aws_s3_bucket.site'` (and the two related resources).
