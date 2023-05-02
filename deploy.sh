#!/bin/bash

gcloud run deploy robswebhub --source . --update-env-vars=[APP_ENVIRONMENT=production]
