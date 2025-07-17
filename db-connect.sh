#!/bin/bash

# Get bastion instance ID
INSTANCE_ID=$(aws ec2 describe-instances --filters "Name=tag:Name,Values=*Bastion*" "Name=instance-state-name,Values=running,stopped" --query "Reservations[0].Instances[0].InstanceId" --output text)

case "$1" in
  start)
    echo "Starting bastion host..."
    aws ec2 start-instances --instance-ids $INSTANCE_ID
    aws ec2 wait instance-running --instance-ids $INSTANCE_ID
    
    # Get DB config from Lambda
    DB_CONFIG=$(aws lambda get-function-configuration --function-name golang-upload-api --query 'Environment.Variables' --output json)
    DB_HOST=$(echo $DB_CONFIG | jq -r '.DB_HOST')
    DB_USER=$(echo $DB_CONFIG | jq -r '.DB_USER')
    DB_NAME=$(echo $DB_CONFIG | jq -r '.DB_NAME')
    DB_PASSWORD=$(echo $DB_CONFIG | jq -r '.DB_PASSWORD')
    
    echo "Connecting to database via bastion..."
    aws ssm start-session --target $INSTANCE_ID --document-name AWS-StartInteractiveCommand --parameters '{"command":["PGPASSWORD='$DB_PASSWORD' psql -h '$DB_HOST' -U '$DB_USER' -d '$DB_NAME'"]}' 
    ;;
  stop)
    echo "Stopping bastion host..."
    aws ec2 stop-instances --instance-ids $INSTANCE_ID
    ;;
  *)
    echo "Usage: $0 {start|stop}"
    ;;
esac