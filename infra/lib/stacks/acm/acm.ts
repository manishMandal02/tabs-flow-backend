import { Construct } from 'constructs';
import { Stack, StackProps } from 'aws-cdk-lib';
import { aws_certificatemanager as acm, aws_ssm as ssm } from 'aws-cdk-lib';
import { config } from '../../../config';

type ACMStackProps = StackProps & {
  stage: string;
};

export class ACMStack extends Stack {
  constructor(scope: Construct, id: string, props: ACMStackProps) {
    super(scope, id, props);

    const certificateName = `${config.AppName}/${props.stage}/api-domain-cert`;

    // Create an ACM certificate for api domain
    const certificate = new acm.Certificate(this, 'Certificate', {
      certificateName,
      domainName: config.Env.API_DOMAIN_NAME,
      validation: acm.CertificateValidation.fromEmail()
    });

    // save table arns in ssm parameters store
    new ssm.StringParameter(this, 'APIDomainCert', {
      parameterName: config.SSMParameterName.MainTableArn,
      stringValue: certificate.certificateArn,
      tier: ssm.ParameterTier.STANDARD
    });

    //* Info: Alternative way to verify domain name for aws certificate manager,
    // create a certificate stack
    // create a dns for subdomain with Route53 and then use the hosted zone to verify the domain
  }
}
