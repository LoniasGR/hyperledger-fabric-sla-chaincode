import FabricCAServices from 'fabric-ca-client';
import { Wallets } from 'fabric-network';
import { resolve, join } from 'path';
import { readFileSync } from 'fs';

let ca;
let wallet;
const dirname = resolve();

async function prepareContext() {
  // load the network configuration
  const ccpPath = resolve(
    dirname,
    '..',
    '..',
    '..',
    'test-network',
    'organizations',
    'peerOrganizations',
    'org1.example.com',
    'connection-org1.json',
  );
  const ccp = JSON.parse(readFileSync(ccpPath, 'utf8'));

  // Create a new CA client for interacting with the CA.
  const caInfo = ccp.certificateAuthorities['ca.org1.example.com'];
  const caTLSCACerts = caInfo.tlsCACerts.pem;
  ca = new FabricCAServices(
    caInfo.url,
    { trustedRoots: caTLSCACerts, verify: false },
    caInfo.caName,
  );

  // Create a new file system based wallet for managing identities.
  const walletPath = join(process.cwd(), 'wallet');
  wallet = await Wallets.newFileSystemWallet(walletPath);
  console.log(`Wallet path: ${walletPath}`);
}

async function createUser(username) {
  try {
    // Check to see if we've already enrolled the user.
    const userIdentity = await wallet.get(username);
    if (userIdentity) {
      console.error(`An identity for the user "${username}" already exists in the wallet`);
      return { error: 'User already exists' };
    }

    // Check to see if we've already enrolled the admin user.
    const adminIdentity = await wallet.get('admin');
    if (!adminIdentity) {
      console.error('An identity for the admin user "admin" does not exist in the wallet');
      console.error('Run the enrollAdmin.js application before retrying');
      return { error: 'Admin is not enrolled' };
    }

    // build a user object for authenticating with the CA
    const provider = wallet.getProviderRegistry().getProvider(adminIdentity.type);
    const adminUser = await provider.getUserContext(adminIdentity, 'admin');

    // Register the user, enroll the user, and import the new identity into the wallet.
    const secret = await ca.register({
      affiliation: 'org1.department1',
      enrollmentID: username,
      role: 'client',
    }, adminUser);
    const enrollment = await ca.enroll({
      enrollmentID: username,
      enrollmentSecret: secret,
    });
    const x509Identity = {
      credentials: {
        certificate: enrollment.certificate,
        privateKey: enrollment.key.toBytes(),
      },
      mspId: 'Org1MSP',
      type: 'X.509',
    };
    await wallet.put(username, x509Identity);
    return {
      publicKey: enrollment.certificate,
      privateKey: enrollment.key.toBytes(),
    };
  } catch (error) {
    console.error(`Failed to enroll user ${username}: ${error}`);
    return { error: 'Uknown failure' };
  }
}

export { prepareContext, createUser };
