import { Construct } from 'constructs';
import { Duration, SecretValue, Stack, StackProps, aws_lambda as lambda } from 'aws-cdk-lib';

import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';

import * as cognito from 'aws-cdk-lib/aws-cognito';

type StatefulStackProps = StackProps & {
  stage: string;
  appName: string;
  googleClientId: string;
  googleClientSecret: string;
  sesEmail: string;
  emailQueueName: string;
};

export class StatefulStack extends Stack {
  // go lambda triggers

  constructor(scope: Construct, id: string, props: StatefulStackProps) {
    super(scope, id, props);

    const defineChallengeHandler = new GoFunction(this, 'DefineChallenge', {
      entry: 'cmd/auth/custom-auth-flow/define-challenge/main.go',
      runtime: lambda.Runtime.PROVIDED_AL2,
      timeout: Duration.seconds(30),
      memorySize: 128,
      environment: {
        EMAIL_QUEUE_NAME: props.emailQueueName
      }
    });

    const createChallengeHandler = new GoFunction(this, 'CreateChallenge', {
      entry: 'cmd/auth/custom-auth-flow/define-challenge/main.go',
      runtime: lambda.Runtime.PROVIDED_AL2,
      timeout: Duration.seconds(30),
      memorySize: 128
    });

    const userPoolName = `${props.stage}${props.appName}UserPool`;

    const userPool = new cognito.UserPool(this, userPoolName, {
      userPoolName,
      deletionProtection: props?.terminationProtection,
      autoVerify: {
        email: true
      },
      signInCaseSensitive: false,
      selfSignUpEnabled: true,
      signInAliases: {
        email: true
      },
      keepOriginal: {
        email: true
      },
      mfa: cognito.Mfa.OFF,
      passwordPolicy: {
        minLength: 8,
        requireLowercase: false,
        requireDigits: false,
        requireUppercase: false,
        requireSymbols: false
      },
      //   TODO - add lambda triggers
      lambdaTriggers: {
        defineAuthChallenge: defineChallengeHandler,
        createAuthChallenge: createChallengeHandler
      }
    });

    userPool.grant(defineChallengeHandler, 'cognito-idp:AdminCreateUser');
    userPool.grant(createChallengeHandler, 'cognito-idp:AdminCreateUser');

    // const role = new iam.Role(this, 'role', {
    //   assumedBy: new iam.ServicePrincipal('foo')
    // });

    // userPool.grant(role, 'cognito-idp:AdminCreateUser');

    const clientSecretValue = new SecretValue(props.googleClientSecret);

    const googleIdentityProvider = new cognito.UserPoolIdentityProviderGoogle(
      this,
      'GoogleIdentityProvider',
      {
        userPool,
        clientSecretValue,
        clientId: props.googleClientId,
        attributeMapping: {
          email: cognito.ProviderAttribute.GOOGLE_EMAIL,
          fullname: cognito.ProviderAttribute.GOOGLE_NAME,
          profilePicture: cognito.ProviderAttribute.GOOGLE_PICTURE
        }
      }
    );

    // attach google auth provider to created user pool
    userPool.registerIdentityProvider(googleIdentityProvider);

    const userPoolClient = new cognito.UserPoolClient(this, 'UserPoolClient', {
      userPool,
      authSessionValidity: Duration.minutes(15),
      idTokenValidity: Duration.days(1),
      refreshTokenValidity: Duration.days(180),
      supportedIdentityProviders: [
        cognito.UserPoolClientIdentityProvider.COGNITO,
        cognito.UserPoolClientIdentityProvider.GOOGLE
      ],
      generateSecret: false,
      oAuth: {
        callbackUrls: ['http://localhost:3000/callback', 'https://tabflows.com/auth'],
        logoutUrls: ['http://localhost:3000/logout', 'https://tabflows.com/logout'],
        flows: {
          authorizationCodeGrant: true
        },
        scopes: [cognito.OAuthScope.EMAIL, cognito.OAuthScope.OPENID, cognito.OAuthScope.PROFILE]
      },
      authFlows: {
        custom: true,
        userSrp: true
      }
    });

    // TODO - pass clientId to auth lambda
  }
}
