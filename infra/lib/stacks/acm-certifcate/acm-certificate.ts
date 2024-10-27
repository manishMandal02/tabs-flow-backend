import { Construct } from 'constructs';
import { Stack, StackProps, aws_dynamodb, aws_ssm as ssm, aws_certificatemanager as acm } from 'aws-cdk-lib';
import { config } from '../../../config';

type ACMCertificateStackProps = StackProps & {
  stage: string;
};

export class ACMCertificateStack extends Stack {
  constructor(scope: Construct, id: string, props: ACMCertificateStackProps) {
    super(scope, id, props);

    // Create an ACM certificate for your domain
    const certificate = new acm.Certificate(this, 'API_DOMAIN_Certificate', {
      domainName: config.Env.API_DOMAIN_NAME,
      validation: acm.CertificateValidation.fromEmail(),
      certificateName: config.AppName + '-api-domain-cert'
    });

    new ssm.StringParameter(this, 'MainTableArn', {
      parameterName: config.SSMParameterName.APIDomainCertificateArn,
      stringValue: certificate.certificateArn,
      tier: ssm.ParameterTier.STANDARD
    });
  }
}
