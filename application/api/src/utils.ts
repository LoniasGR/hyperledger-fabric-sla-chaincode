import * as crypto from 'crypto';
import * as grpc from '@grpc/grpc-js';
import { Identity, Signer, signers } from '@hyperledger/fabric-gateway';
import { promises as fs } from 'fs';

import * as errors from './errors';
import * as constants from './constants';

export type KeysWithStatus = {
  keyPEM: string,
  certPEM: string,
  success: boolean,
  error: string,
};

/**
 * displayInputParameters() will print the global scope parameters used by the main driver routine.
 */
export async function displayInputParameters(): Promise<void> {
  console.debug('*********************************');
  console.debug('**      INPUT PARAMETERS       **');
  console.debug(`SLA channelName:       ${constants.SLAChannelName}`);
  console.debug(`SLA chaincodeName:     ${constants.SLAChaincodeName}`);
  console.debug(`VRU channelName:       ${constants.VRUChannelName}`);
  console.debug(`VRU chaincodeName:     ${constants.VRUChaincodeName}`);
  console.debug(`Parts channelName:       ${constants.PartsChannelName}`);
  console.debug(`Parts chaincodeName:     ${constants.PartsChaincodeName}`);
  console.debug(`mspId:             ${constants.mspId}`);
  console.debug(`cryptoPath:        ${constants.cryptoPath}`);
  console.debug(`tlsCertPath:       ${constants.tlsCertPath}`);
  console.debug(`peerEndpoint:      ${constants.peerEndpoint}`);
  console.debug(`peerHostAlias:     ${constants.peerHostAlias}`);
  console.debug('*********************************');
}

export async function newGrpcConnection(): Promise<grpc.Client> {
  const tlsRootCert = await fs.readFile(constants.tlsCertPath);
  const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);
  return new grpc.Client(constants.peerEndpoint, tlsCredentials, {
    'grpc.ssl_target_name_override': constants.peerHostAlias,
  });
}

export function newIdentity(cert: string):Identity {
  const credentials = Buffer.from(cert);
  const { mspId } = constants;
  return { mspId, credentials };
}

export function newSigner(privateKeyPem: string): Signer {
  const privateKey = crypto.createPrivateKey(privateKeyPem);
  const privateKeySigner = signers.newPrivateKeySigner(privateKey);
  return privateKeySigner;
}

export function toPEMFormat(str: string): string {
  try {
    return str.replace(/\\n/g, '\n');
  } catch (e: unknown) {
    console.error(errors.getErrorMessage(e));
    return errors.getErrorMessage(e);
  }
}

export function oneLiner(str: string): string {
  return str.replace(/\n/g, '');
}

export function keysMatch(key:string, cert: string): boolean | string {
  try {
    const publicKeyFromPrivate = crypto.createPublicKey(key);
    const publicKey = crypto.createPublicKey(cert);
    const exportedPublicKeyFromPrivate = publicKeyFromPrivate.export({ type: 'spki', format: 'pem' });
    const exportedPublicKey = publicKey.export({ type: 'spki', format: 'pem' });
    return (exportedPublicKeyFromPrivate === exportedPublicKey);
  } catch (e: unknown) {
    console.error(errors.getErrorMessage(e));
    return errors.getErrorMessage(e);
  }
}

export function verifyKeys(key: string, cert: string): KeysWithStatus {
  if (key === undefined) {
    return {
      success: false, error: 'Private key is missing', keyPEM: '', certPEM: '',
    };
  }
  if (cert === undefined) {
    return {
      success: false, error: 'Public key is missing', keyPEM: '', certPEM: '',
    };
  }
  const keyPEM = toPEMFormat(key);
  const certPEM = toPEMFormat(cert);

  const match = keysMatch(keyPEM, certPEM);
  if (typeof match !== 'boolean') {
    return {
      success: false, error: match, keyPEM: '', certPEM: '',
    };
  }
  if (!match) {
    return {
      success: false, error: 'Public/private key mismatch', keyPEM: '', certPEM: '',
    };
  }

  return {
    success: true, keyPEM, certPEM, error: '',
  };
}
