const admin = require('firebase-admin');

let firebaseApp;

function initFirebase() {
    if (firebaseApp) return firebaseApp;
    
    const projectId = process.env.FIREBASE_PROJECT_ID;
    const privateKey = process.env.FIREBASE_PRIVATE_KEY?.replace(/\\n/g, '\n');
    const clientEmail = process.env.FIREBASE_CLIENT_EMAIL;
    
    if (!projectId || !privateKey || !clientEmail) {
        throw new Error('Missing Firebase environment variables');
    }
    
    firebaseApp = admin.initializeApp({
        credential: admin.credential.cert({
            projectId,
            privateKey,
            clientEmail
        })
    });
    
    return firebaseApp;
}

async function getUidFromIdentifier(identifier) {
    try {
        if (identifier.includes('@')) {
            const user = await admin.auth().getUserByEmail(identifier);
            return user.uid;
        } else if (/^\+?\d+$/.test(identifier)) {
            const user = await admin.auth().getUserByPhoneNumber(identifier);
            return user.uid;
        }
    } catch (err) {
        if (err.code === 'auth/user-not-found') {
            console.log("ℹ️ User not found for identifier:", identifier);
        } else {
            console.error("❌ Error fetching UID:", err.message);
        }
    }
    return null;
}

exports.handler = async (event) => {
    console.log('🔐 Authorizer started');
    
    const token = event.authorizationToken;
    if (!token || !token.startsWith('Bearer ')) {
        console.log('❌ Invalid token format');
        throw new Error('Unauthorized');
    }
    
    const idToken = token.substring(7);
    console.log('🔑 Token extracted, length:', idToken.length);
    
    try {
        initFirebase();
        console.log('✅ Firebase initialized');
        
        const decodedToken = await admin.auth().verifyIdToken(idToken);
        
        console.log(`✅ Token verified for user: ${decodedToken.email}`);
        
        // Extract ARN components
        const arnParts = event.methodArn.split(':');
        const apiGatewayArnTmp = arnParts[5].split('/');
        const awsAccountId = arnParts[4];
        const region = arnParts[3];
        const restApiId = apiGatewayArnTmp[0];
        const stage = apiGatewayArnTmp[1];
        
        const resourceArn = `arn:aws:execute-api:${region}:${awsAccountId}:${restApiId}/${stage}/*/*`;

        // 🔄 Resolve studentUid from X-Identifier header (if provided)
        const headers = event.headers || {};
        const identifier = headers['x-identifier'] || headers['X-Identifier'];
        let studentUid = null;

        if (identifier) {
            console.log("🔍 Resolving studentUid from identifier:", identifier);
            studentUid = await getUidFromIdentifier(identifier);
        }

        const policy = {
            principalId: decodedToken.uid,
            policyDocument: {
                Version: '2012-10-17',
                Statement: [{
                    Action: 'execute-api:Invoke',
                    Effect: 'Allow',
                    Resource: resourceArn
                }]
            },
            context: {
                email: decodedToken.email || '',
                phone_number: decodedToken.phone_number || '',
                uid: decodedToken.uid,
                studentUid: studentUid ?? null
            }
        };
        
        console.log('✅ Policy generated successfully with studentUid:', studentUid);
        return policy;
    } catch (error) {
        console.log(`❌ Error details: `, error);
        console.log(`❌ Token verification failed: ${error.message}`);
        throw new Error('Unauthorized');
    }
};
