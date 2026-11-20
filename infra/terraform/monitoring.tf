# Host level safety net. App level alerts live in Prometheus.

resource "aws_sns_topic" "alerts" {
  name = "${var.name}-alerts"
}

resource "aws_sns_topic_subscription" "email" {
  count     = var.alert_email == "" ? 0 : 1
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}

# If the underlying AWS hardware fails, recover the instance onto new
# hardware automatically. Private IP, EIP and EBS volume all carry over.
resource "aws_cloudwatch_metric_alarm" "system_check" {
  alarm_name          = "${var.name}-node-system-check"
  alarm_description   = "Hardware level failure: auto recover the node"
  namespace           = "AWS/EC2"
  metric_name         = "StatusCheckFailed_System"
  statistic           = "Maximum"
  period              = 60
  evaluation_periods  = 2
  threshold           = 1
  comparison_operator = "GreaterThanOrEqualToThreshold"
  dimensions          = { InstanceId = aws_instance.node.id }
  alarm_actions       = ["arn:aws:automate:${var.region}:ec2:recover", aws_sns_topic.alerts.arn]
}

# If the OS hangs, reboot it.
resource "aws_cloudwatch_metric_alarm" "instance_check" {
  alarm_name          = "${var.name}-node-instance-check"
  alarm_description   = "OS level failure: reboot the node"
  namespace           = "AWS/EC2"
  metric_name         = "StatusCheckFailed_Instance"
  statistic           = "Maximum"
  period              = 60
  evaluation_periods  = 3
  threshold           = 1
  comparison_operator = "GreaterThanOrEqualToThreshold"
  dimensions          = { InstanceId = aws_instance.node.id }
  alarm_actions       = ["arn:aws:automate:${var.region}:ec2:reboot", aws_sns_topic.alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "cpu" {
  alarm_name          = "${var.name}-node-cpu-high"
  alarm_description   = "CPU above 85 percent for 10 minutes"
  namespace           = "AWS/EC2"
  metric_name         = "CPUUtilization"
  statistic           = "Average"
  period              = 300
  evaluation_periods  = 2
  threshold           = 85
  comparison_operator = "GreaterThanThreshold"
  dimensions          = { InstanceId = aws_instance.node.id }
  alarm_actions       = [aws_sns_topic.alerts.arn]
}
