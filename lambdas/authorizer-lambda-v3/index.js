const admin = require('firebase-admin');
const querystring = require('querystring');

// 🔹 Initialize Firebase Admin SDK
function initFirebase() {
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
}

initFirebase();

// 🔹 Extract UID and email from identifier (email or phone)
async function getUserFromIdentifier(identifier) {
  try {
    let user;
    if (identifier.includes('@')) {
      user = await admin.auth().getUserByEmail(identifier);
    } else if (/^\+?[0-9]{10,15}$/.test(identifier)) {
      user = await admin.auth().getUserByPhoneNumber(identifier);
    } else {
      throw new Error("Invalid identifier format");
    }
    return { uid: user.uid, email: user.email || '' };
  } catch (err) {
    console.error('❌ User lookup failed:', err.message);
    return null;
  }
}

// 🔹 Lambda handler
exports.handler = async (event) => {
  console.log('🔐 Authorizer Event:', JSON.stringify(event, null, 2));

  const token = event.authorizationToken;

  if (!token) {
    console.log('❌ No token provided');
    throw new Error('Unauthorized');
  }

  const cleanToken = token.replace(/^Bearer\s+/i, '');

  try {
    const decoded = await admin.auth().verifyIdToken(cleanToken);
    console.log('✅ Token verified. UID:', decoded.uid);

    let context = {
      uid: decoded.uid,
      email: decoded.email || '',
      phoneNumber: decoded.phone_number || ''
    };

    // 🔸 Check if `identifier` is present in queryStringParameters (for /lookup use case)
    const identifierParam = extractIdentifier(event);
    console.log('🔍 Identifier param:', identifierParam);
    if (identifierParam) {
      const targetUser = await getUserFromIdentifier(identifierParam);
      console.log('👤 Target user:', targetUser);
      if (targetUser) {
        context.targetUID = targetUser.uid;
        context.targetEmail = targetUser.email;
        console.log('✅ Context updated with targetUID:', targetUser.uid, 'targetEmail:', targetUser.email);
      } else {
        console.log('❌ Could not resolve identifier to user');
        throw new Error('Unauthorized');
      }
    }

    const resourceArn = event.methodArn.replace(/\/[^/]+\/[^/]+$/, '/*/*');

    return {
      principalId: decoded.uid,
      policyDocument: {
        Version: '2012-10-17',
        Statement: [
          {
            Action: 'execute-api:Invoke',
            Effect: 'Allow',
            Resource: resourceArn
          }
        ]
      },
      context
    };
  } catch (err) {
    console.error('❌ Authorization failed:', err.message);
    throw new Error('Unauthorized');
  }
};

// 🔹 Extract identifier from request (GET /lookup?identifier=...)
function extractIdentifier(event) {
  if (!event.path || !event.methodArn.includes('/lookup')) return null;

  const queryString = event.path.includes('?') ? event.path.split('?')[1] : '';
  const queryParams = querystring.parse(queryString);

  if (queryParams.identifier) {
    return decodeURIComponent(queryParams.identifier);
  }

  return null;
}
