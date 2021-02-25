#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { CatboxStack } from '../lib/catbox';

const app = new cdk.App();
new CatboxStack(app, process.env.STACK_NAME!, {
    lambdaRoleARN: process.env.LAMBDA_ROLE!,
    s3Region: process.env.S3_REGION!,
    s3Bucket: process.env.S3_BUCKET!,
    s3Prefix: process.env.S3_PREFIX!,

    sentryDSN: process.env.SENTRY_DSN,
    sentryENV: process.env.SENTRY_ENV,

    tags: {
        Project: 'security-vulnmgmt',
    },
});
