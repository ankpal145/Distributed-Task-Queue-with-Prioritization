#!/bin/bash

# Example script to submit tasks to the queue

API_URL="http://localhost:8080/api/v1"

# Submit a high-priority email task
echo "Submitting email task..."
curl -X POST "$API_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "priority": 30,
    "payload": {
      "to": "user@example.com",
      "subject": "Important Notification",
      "body": "This is a high-priority email."
    },
    "max_retries": 3,
    "timeout": 300000000000
  }'

echo -e "\n"

# Submit a normal-priority report generation task
echo "Submitting report task..."
curl -X POST "$API_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "report",
    "priority": 20,
    "payload": {
      "report_type": "monthly",
      "month": "January",
      "year": 2024
    },
    "max_retries": 2,
    "timeout": 600000000000
  }'

echo -e "\n"

# Submit a critical data processing task
echo "Submitting data processing task..."
curl -X POST "$API_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "data_processing",
    "priority": 40,
    "payload": {
      "dataset": "user_analytics",
      "operation": "aggregate"
    },
    "max_retries": 5,
    "timeout": 900000000000
  }'

echo -e "\n"

echo "Tasks submitted successfully!"

