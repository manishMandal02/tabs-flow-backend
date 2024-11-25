import { Construct } from 'constructs';

import { aws_apigateway as apiGateway, aws_certificatemanager as acm } from 'aws-cdk-lib';
import { config } from '../../../config';

type RestApiProps = {
  stage: string;
  domainCertArn: string;
};

export class RestApi extends Construct {
  restAPI: apiGateway.RestApi;
  constructor(scope: Construct, props: RestApiProps, id = 'RestApi') {
    super(scope, id);

    this.restAPI = new apiGateway.RestApi(this, `${config.AppName}-${props.stage}`, {
      endpointTypes:
        props.stage === config.Stage.Test
          ? [apiGateway.EndpointType.REGIONAL]
          : [apiGateway.EndpointType.EDGE],
      defaultCorsPreflightOptions: {
        allowOrigins: apiGateway.Cors.ALL_ORIGINS,
        allowMethods: apiGateway.Cors.ALL_METHODS
      },
      deployOptions: {
        stageName: props.stage
      }
    });

    if (props.stage !== config.Stage.Test) {
      // get the certificate from the arn
      const certificate = acm.Certificate.fromCertificateArn(this, 'Certificate', props.domainCertArn);

      // Create a custom domain name for your API
      const domainName = new apiGateway.DomainName(this, 'CustomDomainName', {
        domainName: config.Env.API_DOMAIN_NAME,
        certificate: certificate,
        endpointType: apiGateway.EndpointType.EDGE
      });

      // Map the custom domain to your API
      new apiGateway.BasePathMapping(this, 'ApiMapping', {
        domainName: domainName,
        restApi: this.restAPI
      });
    }
  }
}
