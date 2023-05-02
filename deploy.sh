#!/bin/bash

gcloud run deploy --source . --update-env-vars APP_ENVIRONMENT=production --port 8080 robswebhub
