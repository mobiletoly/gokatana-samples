#!/bin/bash

IAMSERVICE_DATABASE_PASSWORD=postgres \
  IAMSERVICE_DATABASE_USER=postgres \
  IAMSERVICE_CREDENTIALS_JWTSECRET=secret \
  IAMSERVICE_GCLOUD_EMAIL_FROM=noreply@some-company-domain.com \
  IAMSERVICE_GCLOUD_EMAIL_USER=account@some-company-domain.com \
  IAMSERVICE_GCLOUD_SERVICEJSON="$(<./iamservice-credentials.json)" \
  air run --deployment=local
