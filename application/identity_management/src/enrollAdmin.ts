/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
import FabricCAServices from 'fabric-ca-client';
import { Wallets } from 'fabric-network';
import path from 'path';
import fs from 'fs';

async function enrollAdmin(org: number, ledger: string): Promise<void> {
  try {
    // load the network configuration
    const ccpPath = path.resolve(
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
    const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

    // Create a new CA client for interacting with the CA.
    const caInfo = ccp.certificateAuthorities[`ca.org${org}.example.com`];
    const caTLSCACerts = caInfo.tlsCACerts.pem;
    const ca = new FabricCAServices(
      caInfo.url,
      { trustedRoots: caTLSCACerts, verify: false },
      caInfo.caName,
    );

    // Create a new file system based wallet for managing identities.
    const walletPath = path.join(process.cwd(), `wallet_${ledger}`);
    const wallet = await Wallets.newFileSystemWallet(walletPath);
    console.debug(`Wallet path: ${walletPath}`);

    // Check to see if we've already enrolled the admin user.
    const identity = await wallet.get('admin');
    if (identity) {
      console.debug(`An identity for the admin user admin already exists in wallet ${ledger}`);
      return;
    }

    // Enroll the admin user, and import the new identity into the wallet.
    const enrollment = await ca.enroll({ enrollmentID: 'admin', enrollmentSecret: 'adminpw' });
    const x509Identity = {
      credentials: {
        certificate: enrollment.certificate,
        privateKey: enrollment.key.toBytes(),
      },
      mspId: `Org${org}MSP`,
      type: 'X.509',
    };
    await wallet.put('admin', x509Identity);
    console.log(`Successfully enrolled admin user "admin" and imported it into the ${ledger} wallet`);
  } catch (error) {
    console.error(`Failed to enroll admin user "admin" in wallet ${ledger}: ${error}`);
    process.exit(1);
  }
}

export default enrollAdmin;
