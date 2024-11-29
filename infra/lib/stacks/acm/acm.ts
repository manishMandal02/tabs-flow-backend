import { Construct } from 'constructs';
import { Stack, StackProps, aws_certificatemanager as acm, custom_resources } from 'aws-cdk-lib';

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

    // set the certificate arn to a  ssm parameter in the main region
    new custom_resources.AwsCustomResource(this, 'PutParameterAPIDomainCert', {
      onCreate: {
        service: 'SSM',
        action: 'putParameter',
        parameters: {
          Name: config.SSMParameterName.APIDomainCertArn,
          Value: certificate.certificateArn,
          Type: 'String',
          Overwrite: true
        },
        region: config.Env.AWS_REGION,
        physicalResourceId: custom_resources.PhysicalResourceId.of('PutParameterAPIDomainCert')
      },
      policy: custom_resources.AwsCustomResourcePolicy.fromSdkCalls({
        resources: custom_resources.AwsCustomResourcePolicy.ANY_RESOURCE
      })
    });
  }
}
