import FabricCAServices from 'fabric-ca-client';
import { Wallet, Wallets } from 'fabric-network';
import { resolve, join } from 'path';
import { readFileSync } from 'fs';

let ca: Array<FabricCAServices> = [];
let wallet: Array<Wallet> = [];

type KeysOrError = {
  error?: string
  publicKey?: string
  privateKey?: string
}

async function prepareContext(org: number, ledger: string): Promise<void> {
  // load the network configuration
  const ccpPath = resolve(
    __dirname,
    '..',
    '..',
    '..',
    'test-network',
    'organizations',
    'peerOrganizations',
    `org${org}.example.com`,
    `connection-org${org}.json`,
  );
  const ccp = JSON.parse(readFileSync(ccpPath, 'utf8'));

  // Create a new CA client for interacting with the CA.
  const caInfo = ccp.certificateAuthorities[`ca.org${org}.example.com`];
  const caTLSCACerts = caInfo.tlsCACerts.pem;
  ca[org-1] = new FabricCAServices(
    caInfo.url,
    { trustedRoots: caTLSCACerts, verify: false },
    caInfo.caName,
  );

  // Create a new file system based wallet for managing identities.
  const walletPath = join(process.cwd(), `wallet_${ledger}`);
  wallet[org-1] = await Wallets.newFileSystemWallet(walletPath);
  console.debug(`Wallet path: ${walletPath}`);
}

async function createUser(username: string, org: number, ledger: string): Promise<KeysOrError> {
  try {
    // Check to see if we've already enrolled the user.
    const userIdentity = await wallet[org-1].get(username);
    if (userIdentity) {
      console.error(`An identity for the user "${username}" already exists in wallet ${ledger}`);
      return { error: 'User already exists' };
    }

    // Check to see if we've already enrolled the admin user.
    const adminIdentity = await wallet[org-1].get('admin');
    if (!adminIdentity) {
      console.error(`An identity for the admin user admin does not exist in the wallet ${ledger}`);
      console.error('Run the enrollAdmin.js application before retrying');
      return { error: 'Admin is not enrolled' };
    }

    // build a user object for authenticating with the CA
    const provider = wallet[org-1].getProviderRegistry().getProvider(adminIdentity.type);
    const adminUser = await provider.getUserContext(adminIdentity, 'admin');

    // Register the user, enroll the user, and import the new identity into the wallet.
    const secret = await ca[org-1].register({
      affiliation: 'org1.department1',
      enrollmentID: username,
      role: 'client',
    }, adminUser);
    const enrollment = await ca[org-1].enroll({
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
    await wallet[org-1].put(username, x509Identity);
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
