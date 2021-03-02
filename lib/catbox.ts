import * as cdk from '@aws-cdk/core';
import * as lambda from '@aws-cdk/aws-lambda';
import * as iam from '@aws-cdk/aws-iam';
import * as sns from '@aws-cdk/aws-sns';
import * as sqs from '@aws-cdk/aws-sqs';
import * as dynamodb from '@aws-cdk/aws-dynamodb';
import * as apigateway from'@aws-cdk/aws-apigateway';
import {
  DynamoEventSource,
  SqsEventSource,
} from '@aws-cdk/aws-lambda-event-sources';
import * as events from '@aws-cdk/aws-events';
import * as eventsTargets from '@aws-cdk/aws-events-targets';
import { SqsSubscription } from '@aws-cdk/aws-sns-subscriptions';

import { LambdaFunction } from '@aws-cdk/aws-events-targets';
import * as path from 'path';
import * as fs from 'fs';

// Definitions
const scanQueueTimeout = cdk.Duration.seconds(300);
const inspectQueueTimeout = cdk.Duration.seconds(300);

interface CatBoxProps extends cdk.StackProps {
  readonly lambdaRoleARN?: string;
  readonly s3Region: string;
  readonly s3Bucket: string;
  readonly s3Prefix: string;

  // frontendBaseURL: string;
  // readonly secretARN: string;

  readonly sentryDSN?: string;
  readonly sentryENV?: string;
}

interface CatBoxQueues {
  readonly scanQueue: sqs.Queue;
  readonly inspectQueue: sqs.Queue;
  readonly scanDLQ: sqs.Queue;
  readonly inspectDLQ: sqs.Queue;
};

interface CatBoxFunctions {
  readonly apiHandler: lambda.Function;
  readonly enqueueScan: lambda.Function;
  readonly inspect: lambda.Function;
  readonly notify: lambda.Function;
  readonly scanImage: lambda.Function;
  readonly updateDB: lambda.Function;
};

export class CatboxStack extends cdk.Stack {
  readonly notifyTopic: sns.Topic;

  readonly metaTable: dynamodb.Table;
  readonly queues: CatBoxQueues;
  readonly functions: CatBoxFunctions;

  constructor(scope: cdk.Construct, id: string, props: CatBoxProps) {
    super(scope, id, props);

    // DynamoDB
    this.metaTable = new dynamodb.Table(this, 'metaTable', {
      partitionKey: { name: 'pk', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'sk', type: dynamodb.AttributeType.STRING },
      stream: dynamodb.StreamViewType.NEW_AND_OLD_IMAGES,
      timeToLiveAttribute: 'expires_at',
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
    });
    this.metaTable.addGlobalSecondaryIndex({
      indexName: 'secondary',
      partitionKey: { name: 'pk2', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'sk2', type: dynamodb.AttributeType.STRING },
    });

    // SQS
    this.queues = setupQueues(this);

    // TODO: create S3 bucket if S3Region, S3Bucket and S3Prefix is not provided

    this.functions = setupLambda(this, props, this.metaTable, this.queues);

    setupWeb(this, this.functions.apiHandler);
  }
}

function setupQueues(stack :cdk.Stack): CatBoxQueues {
  const scanDLQ = new sqs.Queue(stack, 'scanDLQ');
  const scanQueue = new sqs.Queue(stack, 'scanQueue', {
    visibilityTimeout: scanQueueTimeout,
    deadLetterQueue: {
      maxReceiveCount: 3,
      queue: scanDLQ,
    },
  });

  const inspectDLQ = new sqs.Queue(stack, 'inspectDLQ');
  const inspectQueue = new sqs.Queue(stack, 'inspectQueue', {
    visibilityTimeout: inspectQueueTimeout,
    deadLetterQueue: {
      maxReceiveCount: 3,
      queue: inspectDLQ,
    },
  });

  return {
    scanDLQ: scanDLQ,
    scanQueue: scanQueue,
    inspectDLQ: inspectDLQ,
    inspectQueue: inspectQueue,
  };
}

interface lambdaConfig {
  readonly funcName: string;
  readonly events: lambda.IEventSource[],
  readonly timeout: cdk.Duration;
  readonly memorySize?: number;
};

function setupLambda(stack: cdk.Stack, props: CatBoxProps, table: dynamodb.Table, queues: CatBoxQueues): CatBoxFunctions {
  // IAM role
  const lambdaRole = (props.lambdaRoleARN !== undefined) ? iam.Role.fromRoleArn(stack, 'LambdaRole', props.lambdaRoleARN, {
    mutable: false,
  }) : undefined;

  // Lambda functions
  const rootPath = path.resolve(__dirname, '..');
  const asset = lambda.Code.fromAsset(rootPath, {
    bundling: {
      image: lambda.Runtime.GO_1_X.bundlingDockerImage,
      user: 'root',
      command: ['make', 'asset'],
      environment: {
        GOARCH: 'amd64',
        GOOS: 'linux',
      },
    },
  });
  const baseEnvVars = {
    TABLE_NAME: table.tableName,
    S3_REGION: props.s3Region,
    S3_BUCKET: props.s3Bucket,
    S3_PREFIX: props.s3Prefix,

    SCAN_QUEUE_URL: queues.scanQueue.queueUrl,
    INSPECT_QUEUE_URL: queues.inspectQueue.queueUrl,

    SENTRY_DSN: props.sentryDSN || "",
    SENTRY_ENV: props.sentryENV || "",
  };

  const newLambda = function(stack: cdk.Stack, config: lambdaConfig): lambda.Function {
    return  new lambda.Function(stack, config.funcName, {
      runtime: lambda.Runtime.GO_1_X,
      handler: config.funcName,
      code: asset,
      role: lambdaRole,
      timeout: cdk.Duration.seconds(300),
      memorySize: config.memorySize || 256,
      environment: baseEnvVars,
      reservedConcurrentExecutions: 1,

      events: config.events,
    });
  };

  const functions = {
    apiHandler: newLambda(stack, {
      funcName: 'apiHandler',
      events: [],
      timeout: cdk.Duration.seconds(10),
    }),
    enqueueScan: newLambda(stack, {
      funcName: 'enqueueScan',
      events: [],
      timeout: cdk.Duration.seconds(10),
    }),
    inspect: newLambda(stack, {
      funcName: 'inspect',
      events: [new SqsEventSource(queues.inspectQueue)],
      timeout: inspectQueueTimeout,
    }),
    notify: newLambda(stack, {
      funcName: 'notify',
      events: [],
      timeout: cdk.Duration.seconds(30),
    }),
    scanImage: newLambda(stack, {
      funcName: 'scanImage',
      events: [new SqsEventSource(queues.scanQueue)],
      timeout: scanQueueTimeout,
      memorySize: 2046,
    }),
    updateDB: newLambda(stack, {
      funcName: 'updateDB',
      events: [],
      timeout: cdk.Duration.minutes(3),
    }),
  };

  new events.Rule(stack, 'periodicInvokeUpdateDB' + functions.enqueueScan, {
    schedule: events.Schedule.rate(cdk.Duration.hours(1)),
    targets: [new eventsTargets.LambdaFunction(functions.updateDB)],
  });

  new events.Rule(stack, 'periodicInvokeEnqueueScan' + functions.enqueueScan, {
    schedule: events.Schedule.rate(cdk.Duration.hours(24)),
    targets: [new eventsTargets.LambdaFunction(functions.updateDB)],
  });

  new events.Rule(stack, 'ECREvent', {
    eventPattern: {
      source: ['aws.ecr'],
      detail: { eventName: ['PutImage'] },
    },
    targets: [new eventsTargets.LambdaFunction(functions.enqueueScan)],
  });

  return functions;
}

function setupWeb(stack: cdk.Stack, apiHandler: lambda.Function) {
  const api = new apigateway.LambdaRestApi(stack, 'catboxAPI', {
    handler: apiHandler,
    proxy: false,
    cloudWatchRole: false,
    endpointTypes: [apigateway.EndpointType.PRIVATE],
    policy: new iam.PolicyDocument({
      statements: [
        new iam.PolicyStatement({
          actions: ['execute-api:Invoke'],
          resources: ['execute-api:/*/*'],
          effect: iam.Effect.ALLOW,
          principals: [new iam.AnyPrincipal()],
        }),
      ],
    }),
  });

  // UI assets
  api.root.addMethod('GET');
  api.root.addResource('js').addResource('bundle.js').addMethod("GET");

  // auth
  const auth = api.root.addResource('auth');
  auth.addMethod('GET');
  auth.addResource('logout').addMethod('GET');
  const authGoogle = auth.addResource('google');
  authGoogle.addMethod('GET');
  authGoogle.addResource('callback').addMethod('GET');

  // API
  const v1 = api.root.addResource('api').addResource('v1');

  // repository
  const repository = v1.addResource('repository');
  repository.addMethod('GET');
}
