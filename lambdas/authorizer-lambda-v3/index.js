const admin = require('firebase-admin');

// Initialize Firebase Admin SDK
if (!admin.apps.length) {
    admin.initializeApp({
        credential: admin.credential.cert({
            type: "service_account",
            project_id: process.env.FIREBASE_PROJECT_ID,
            private_key_id: process.env.FIREBASE_PRIVATE_KEY_ID,
            private_key: process.env.FIREBASE_PRIVATE_KEY.replace(/\\n/g, '\n'),
            client_email: process.env.FIREBASE_CLIENT_EMAIL,
            client_id: process.env.FIREBASE_CLIENT_ID,
            auth_uri: "https://accounts.google.com/o/oauth2/auth",
            token_uri: "https://oauth2.googleapis.com/token",
            auth_provider_x509_cert_url: "https://www.googleapis.com/oauth2/v1/certs",
            client_x509_cert_url: process.env.FIREBASE_CLIENT_CERT_URL
        })
    });
}



exports.handler = async (event) => {
    console.log('🔐 V3 Authorizer received event:', JSON.stringify(event, null, 2));
    
    const token = event.authorizationToken;
    
    if (!token) {
        console.log('❌ No token provided');
        throw new Error('Unauthorized');
    }

    // Remove 'Bearer ' prefix if present
    const cleanToken = token.replace(/^Bearer\s+/i, '');
    
    try {
        // Verify Firebase token
        const decodedToken = await admin.auth().verifyIdToken(cleanToken);
        console.log('✅ Token verified for UID:', decodedToken.uid);
        
        let context = {
            uid: decodedToken.uid,
            email: decodedToken.email || '',
            phoneNumber: decodedToken.phone_number || ''
        };
        
        // Only for /v3/students/update path
        if (event.methodArn.includes('/v3/students/update')) {
            const email = event.queryStringParameters?.email;
            const phoneNumber = event.queryStringParameters?.phoneNumber;
            
            if (email || phoneNumber) {
                const targetUID = await getUid(email || phoneNumber);
                if (targetUID) {
                    context.targetUID = targetUID;
                }
            }
        }
        
        return {
            principalId: decodedToken.uid,
            policyDocument: {
                Version: '2012-10-17',
                Statement: [
                    {
                        Action: 'execute-api:Invoke',
                        Effect: 'Allow',
                        Resource: event.methodArn
                    }
                ]
            },
            context: context
        };
    } catch (error) {
        console.log('❌ Token verification failed:', error.message);
        throw new Error('Unauthorized');
    }
};

async function getUid(identifier) {
    try {
        let userRecord;

        if (identifier.includes("@")) {
            // It's an email
            userRecord = await admin.auth().getUserByEmail(identifier);
        } else if (identifier.startsWith("+")) {
            // It's a phone number
            userRecord = await admin.auth().getUserByPhoneNumber(identifier);
        } else {
            throw new Error("Invalid identifier. Must be a valid email or E.164 phone number.");
        }

        console.log("✅ UID:", userRecord.uid);
        return userRecord.uid;

    } catch (err) {
        console.error("❌ Error fetching UID:", err.message);
        return null;
    }
}