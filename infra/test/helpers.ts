import { NotificationsService } from './../lib/stacks/services/notifications';
import { EmailService } from './../lib/stacks/services/email';
import { Match, Template } from 'aws-cdk-lib/assertions';
import { config } from '../config';

export const verifyLambdaSQSPermission = (template: Template, service: string): boolean => {
  let iamStatement: any[] = [];

  let verifiedSQSIamPolicy = false;

  // verify lambda permission for sqs
  const policies = template.findResources('AWS::IAM::Policy', {});

  for (const policy of Object.values(policies)) {
    const PolicyDocument = policy['Properties']['PolicyDocument'] as any;

    if ((policy['Properties']['PolicyName'] as string).startsWith('LambdaRoleDefaultPolicy')) {
      const statement = PolicyDocument['Statement'] as any;
      const actions = statement[0]['Action'] as string[];
      if (Array.isArray(statement) && Array.isArray(actions) && actions[0].startsWith('sqs:')) {
        iamStatement = statement as any[];
        break;
      }
    }
  }

  if (!iamStatement.length) {
    return false;
  }

  for (const statement of iamStatement) {
    if (statement.Effect !== 'Allow' || !Array.isArray(statement.Action)) continue;

    const actions = statement.Action as string[];

    if (!actions[0].startsWith('sqs:')) continue;

    // iam policy for sqs permission
    expect(actions).toHaveLength(5);

    expect(actions).toContain('sqs:ReceiveMessage');
    expect(actions).toContain('sqs:DeleteMessage');

    // verify policy resource
    if (Array.isArray(statement.Resource)) {
      // multiple resources
      const resources = statement.Resource as { [key: string]: string[] }[];
      for (const resource of resources) {
        if (resource['Fn::GetAtt'][0].includes(service) && resource['Fn::GetAtt'][1] === 'Arn') {
          verifiedSQSIamPolicy = true;
          break;
        }
      }
    } else {
      if (
        statement.Resource['Fn::GetAtt'][0].includes(service) &&
        statement.Resource['Fn::GetAtt'][1] === 'Arn'
      ) {
        verifiedSQSIamPolicy = true;
        break;
      }
    }
  }

  return verifiedSQSIamPolicy;
};

type AssertLambdaFunctionProps = {
  template: Template;
  service: string;
  stage: string;
  env: Record<string, any>;
  name?: string;
};

export const assertLambdaFunction = ({ template, service, name, stage, env }: AssertLambdaFunctionProps) => {
  template.hasResourceProperties('AWS::Lambda::Function', {
    FunctionName: name ?? `${service}_${stage}`,
    Handler: 'bootstrap',
    Architectures: [config.Lambda.Architecture.name],
    Runtime: config.Lambda.Runtime.name,
    MemorySize: config.Lambda.MemorySize,
    Timeout: config.Lambda.Timeout.toSeconds(),
    Environment: {
      Variables: env
    }
  });
};

export const assertSQSQueue = (template: Template, queueName: string, service: string) => {
  template.hasResourceProperties('AWS::SQS::Queue', {
    QueueName: queueName,
    VisibilityTimeout: 300,
    DelaySeconds: 1,
    // assert dead letter queue
    RedrivePolicy: {
      deadLetterTargetArn: {
        'Fn::GetAtt': [Match.stringLikeRegexp(service), 'Arn']
      },
      maxReceiveCount: 3
    }
  });
};

export const assertLambdaSQSEventSourceMapping = (template: Template, service: string) => {
  // assert lambda sqs event source
  template.hasResourceProperties('AWS::Lambda::EventSourceMapping', {
    FunctionName: {
      Ref: Match.stringLikeRegexp(service)
    },
    BatchSize: 1,
    EventSourceArn: {
      'Fn::GetAtt': [Match.stringLikeRegexp(service), 'Arn']
    }
  });
};

type AssertLambdaAPIGatewayIntegrationProps = {
  template: Template;
  service: string;
  baseURL: string;
  hasAuthorization?: boolean;
};

export const assertLambdaAPIGatewayIntegration = ({
  template,
  service,
  baseURL,
  hasAuthorization = false
}: AssertLambdaAPIGatewayIntegrationProps) => {
  // assert api gw /auth resource
  template.hasResourceProperties('AWS::ApiGateway::Resource', {
    ParentId: {
      'Fn::GetAtt': [Match.stringLikeRegexp(config.AppName), 'RootResourceId']
    },
    PathPart: baseURL
  });

  // assert api gw {proxy+} resource
  template.hasResourceProperties('AWS::ApiGateway::Resource', {
    PathPart: '{proxy+}',
    ParentId: {
      Ref: Match.stringLikeRegexp(config.AppName) && Match.stringLikeRegexp(baseURL)
    }
  });

  const authorizerId = {
    Ref: Match.stringLikeRegexp(config.AppName) && Match.stringLikeRegexp('Authorizer')
  };

  // assert api gw ANY method
  template.hasResourceProperties('AWS::ApiGateway::Method', {
    HttpMethod: 'ANY',
    AuthorizationType: hasAuthorization ? 'CUSTOM' : 'NONE',
    ...(hasAuthorization ? { AuthorizerId: authorizerId } : {}),
    Integration: {
      IntegrationHttpMethod: 'POST',
      Type: 'AWS_PROXY',
      Uri: {
        'Fn::Join': [
          '',
          [
            'arn:',
            {
              Ref: 'AWS::Partition'
            },
            Match.anyValue(),
            {
              'Fn::GetAtt': [Match.stringLikeRegexp(service), 'Arn']
            },
            '/invocations'
          ]
        ]
      }
    },
    ResourceId: {
      Ref: Match.stringLikeRegexp(config.AppName) && Match.stringLikeRegexp(baseURL)
    },
    RestApiId: {
      Ref: Match.stringLikeRegexp(config.AppName)
    }
  });
};
