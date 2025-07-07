import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as s3 from 'aws-cdk-lib/aws-s3';

export class BucketsStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const names = [
      "CLS6-TELUGU", "CLS6-HINDI", "CLS6-ENGLISH", "CLS6-MATHS", "CLS6-SCIENCE", "CLS6-SOCIAL",
      "CLS7-TELUGU", "CLS7-HINDI", "CLS7-ENGLISH", "CLS7-MATHS", "CLS7-SCIENCE", "CLS7-SOCIAL",
      "CLS8-TELUGU", "CLS8-HINDI", "CLS8-ENGLISH", "CLS8-MATHS", "CLS8-SCIENCE", "CLS8-SOCIAL",
      "CLS9-TELUGU", "CLS9-HINDI", "CLS9-ENGLISH", "CLS9-MATHS", "CLS9-SCIENCE", "CLS9-SOCIAL",
      "CLS10-TELUGU", "CLS10-HINDI", "CLS10-ENGLISH", "CLS10-MATHS", "CLS10-SCIENCE", "CLS10-SOCIAL",
      "CLS10-BRIDGE", "CLS10-POLYTECHNIC", "CLS10-FORMULAS",
      "CLS11-MPC-PHYSICS", "CLS11-MPC-MATHS1A", "CLS11-MPC-MATHS1B", "CLS11-MPC-CHEMISTRY",
      "CLS11-MPC-EAMCET", "CLS11-MPC-JEEMAINS", "CLS11-MPC-JEEADV",
      "CLS12-MPC-PHYSICS", "CLS12-MPC-MATHS2A", "CLS12-MPC-MATHS2B", "CLS12-MPC-CHEMISTRY",
      "CLS12-MPC-EAMCET", "CLS12-MPC-JEEMAINS", "CLS12-MPC-JEEADV",
      "CLS11-BIPC-PHYSICS", "CLS11-BIPC-BOTANY", "CLS11-BIPC-ZOOLOGY", "CLS11-BIPC-CHEMISTRY",
      "CLS11-BIPC-EAPCET", "CLS11-BIPC-NEET",
      "CLS12-BIPC-PHYSICS", "CLS12-BIPC-BOTANY", "CLS12-BIPC-ZOOLOGY", "CLS12-BIPC-CHEMISTRY",
      "CLS12-BIPC-EAPCET", "CLS12-BIPC-NEET"
    ];

    for (const name of names) {
      const cleanName = name.toLowerCase().replace(/[^a-z0-9]/g, '-');

      new s3.Bucket(this, `Bucket-${cleanName}`, {
        bucketName: cleanName,
        versioned: true,
        removalPolicy: cdk.RemovalPolicy.DESTROY,
        autoDeleteObjects: true,
      });
    }
  }
}
