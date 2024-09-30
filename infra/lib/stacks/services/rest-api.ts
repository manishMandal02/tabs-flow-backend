import { Construct } from 'constructs';

import {
  aws_apigateway as apiGateway,
  aws_certificatemanager as acm,
  CfnOutput,
  aws_route53
} from 'aws-cdk-lib';
import { config } from '../../../config';

type RestApiProps = {
  stage: string;
};

export class RestApi extends Construct {
  restAPI: apiGateway.RestApi;
  constructor(scope: Construct, props: RestApiProps, id = 'RestApi') {
    super(scope, id);

    this.restAPI = new apiGateway.RestApi(this, `${config.AppName}-${props.stage}`, {
      deployOptions: {
        stageName: props.stage
      }
    });

    // Create an ACM certificate for your domain
    const certificate = new acm.Certificate(this, 'Certificate', {
      domainName: config.Env.API_DOMAIN,
      validation: acm.CertificateValidation.fromEmail(),
      certificateName: config.AppName + 'api-cert'
    });

    //* Info: Alternative way to verify domain name for aws certificate manager,
    // create a certificate stack
    // create a dns for subdomain with Route53 and then use the hosted zone to verify the domain

    // Create a custom domain name for your API
    const domainName = new apiGateway.DomainName(this, 'CustomDomainName', {
      domainName: config.Env.API_DOMAIN,
      certificate: certificate,
      endpointType: apiGateway.EndpointType.REGIONAL
    });

    // Map the custom domain to your API
    new apiGateway.BasePathMapping(this, 'ApiMapping', {
      domainName: domainName,
      restApi: this.restAPI
    });

    //* Add the API GW domain name as cname record in your dns

    // Output the Regional Domain Name
    // new CfnOutput(this, 'RegionalDomainName', {
    //   value: domainName.domainNameAliasDomainName,
    //   description: 'Regional Domain Name Alias'
    // });
  }
}
