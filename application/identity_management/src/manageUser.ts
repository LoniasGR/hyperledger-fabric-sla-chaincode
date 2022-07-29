import FabricCAServices from 'fabric-ca-client';
import { Identity, Wallet, Wallets } from 'fabric-network';
import { resolve, join } from 'path';
import { readFileSync } from 'fs';
import * as errors from './errors';

const ca: Array<FabricCAServices> = [];
const wallet: Array<Wallet> = [];

type KeysOrError = {
  error?: string
  publicKey?: string
  privateKey?: string
}

type x509Identity = {
  credentials: {
    certificate: string,
    privateKey: string,
  },
  mspId: string,
  type: string,
};

function toPEMFormat(str: string): string {
  try {
    return str.replace(/\\n/g, '\n');
  } catch (e: unknown) {
    console.error(errors.getErrorMessage(e));
    return errors.getErrorMessage(e);
  }
}

export async function prepareContext(org: number, ledger: string): Promise<void> {
  // load the network configuration
  const ccpPath = resolve(
    __dirname,
    '..',
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
  ca[org - 1] = new FabricCAServices(
    caInfo.url,
    { trustedRoots: caTLSCACerts, verify: false },
    caInfo.caName,
  );

  // Create a new file system based wallet for managing identities.
  const walletPath = join(process.cwd(), `wallet_${ledger}`);
  wallet[org - 1] = await Wallets.newFileSystemWallet(walletPath);
  console.debug(`Wallet path: ${walletPath}`);
}

export async function createUser(
  username: string,
  org: number,
  ledger: string,
) :Promise<KeysOrError> {
  try {
    // Check to see if we've already enrolled the user.
    const userIdentity = await wallet[org - 1].get(username);
    if (userIdentity) {
      console.error(`An identity for the user "${username}" already exists in wallet ${ledger}`);
      return { error: 'User already exists' };
    }

    // Check to see if we've already enrolled the admin user.
    const adminIdentity = await wallet[org - 1].get('admin');
    if (!adminIdentity) {
      console.error(`An identity for the admin user admin does not exist in the wallet ${ledger}`);
      console.error('Run the enrollAdmin.js application before retrying');
      return { error: 'Admin is not enrolled' };
    }

    // build a user object for authenticating with the CA
    const provider = wallet[org - 1].getProviderRegistry().getProvider(adminIdentity.type);
    const adminUser = await provider.getUserContext(adminIdentity, 'admin');

    // Register the user, enroll the user, and import the new identity into the wallet.
    const secret = await ca[org - 1].register({
      affiliation: 'org1.department1',
      enrollmentID: username,
      role: 'client',
    }, adminUser);
    const enrollment = await ca[org - 1].enroll({
      enrollmentID: username,
      enrollmentSecret: secret,
    });
    const userX509Identity = {
      credentials: {
        certificate: enrollment.certificate,
        privateKey: enrollment.key.toBytes(),
      },
      mspId: 'Org1MSP',
      type: 'X.509',
    };
    await wallet[org - 1].put(username, userX509Identity);
    return {
      publicKey: enrollment.certificate,
      privateKey: enrollment.key.toBytes(),
    };
  } catch (error) {
    console.error(`Failed to enroll user ${username}: ${error}`);
    return { error: 'Uknown failure' };
  }
}

export async function userExists(cert: string): Promise<{found: boolean, org: number}> {
  let found = false;
  let org = 0;
  const asyncUsers: Array<Promise<Array<string>>> = [];
  const asyncCredentials: Array<Promise<Identity | undefined>> = [];
  const credentials = [];
  for (let i = 0; i < wallet.length; i += 1) {
    asyncUsers[i] = wallet[i].list();
  }
  const users = await Promise.all(asyncUsers);
  console.log(users);
  for (let i = 0; i < wallet.length; i += 1) {
    for (let u = 0; u < users[i].length; u += 1) {
      asyncCredentials[u] = wallet[i].get(users[i][u]);
    }
    // eslint-disable-next-line no-await-in-loop
    credentials[i] = await Promise.all(asyncCredentials);
  }

  for (let i = 0; i < wallet.length; i += 1) {
    for (let u = 0; u < users.length; u += 1) {
      const user = credentials[i][u];
      if (user !== undefined) {
        const userJSON = JSON.stringify(user);
        const actualUser : x509Identity = JSON.parse(userJSON);
        if (toPEMFormat(cert) === toPEMFormat(actualUser.credentials.certificate)) {
          found = true;
          org = i + 1;
          break;
        }
      }
    }
  }
  return { found, org };
}
