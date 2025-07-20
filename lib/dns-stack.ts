import * as cdk from 'aws-cdk-lib';
import * as route53 from 'aws-cdk-lib/aws-route53';
import { Construct } from 'constructs';

export class DnsStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create hosted zone for gradeup.guru
    const hostedZone = new route53.HostedZone(this, 'GradeUpZone', {
      zoneName: 'gradeup.guru'
    });

    // SPF record
    new route53.TxtRecord(this, 'SpfRecord', {
      zone: hostedZone,
      values: ['v=spf1 include:_spf.firebasemail.com ~all']
    });

    // Firebase verification
    new route53.TxtRecord(this, 'FirebaseRecord', {
      zone: hostedZone,
      values: ['firebase=gothinkerstech']
    });

    // DKIM records
    new route53.CnameRecord(this, 'DkimRecord1', {
      zone: hostedZone,
      recordName: 'firebase1._domainkey',
      domainName: 'mail-gradeup-guru.dkim1._domainkey.firebasemail.com'
    });

    new route53.CnameRecord(this, 'DkimRecord2', {
      zone: hostedZone,
      recordName: 'firebase2._domainkey',
      domainName: 'mail-gradeup-guru.dkim2._domainkey.firebasemail.com'
    });

    // Output nameservers
    new cdk.CfnOutput(this, 'NameServers', {
      value: hostedZone.hostedZoneNameServers?.join(', ') || '',
      description: 'Update these nameservers at your domain registrar'
    });
  }
}