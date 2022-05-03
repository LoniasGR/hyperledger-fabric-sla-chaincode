import * as path from 'path';
import * as crypto from 'crypto';
import * as grpc from '@grpc/grpc-js';
import { Identity, Signer, signers } from '@hyperledger/fabric-gateway';
import { promises as fs } from 'fs';

import * as errors from './errors';

export function envOrDefault(key: string, defaultValue: string): string {
  return process.env[key] || defaultValue;
}

export const expressPort = envOrDefault('EXPRESS_PORT', '8000');

export const channelName = envOrDefault('CHANNEL_NAME', 'mychannel');
export const chaincodeName = envOrDefault('CHAINCODE_NAME', 'basic');
export const mspId = envOrDefault('MSP_ID', 'Org1MSP');

// Path to crypto materials.
export const cryptoPath = envOrDefault('CRYPTO_PATH', path.resolve(__dirname, '..', '..', '..', '..', 'test-network', 'organizations', 'peerOrganizations', 'org1.example.com'));

// Path to peer tls certificate.
export const tlsCertPath = envOrDefault('TLS_CERT_PATH', path.resolve(cryptoPath, 'peers', 'peer0.org1.example.com', 'tls', 'ca.crt'));

// Gateway peer endpoint.
export const peerEndpoint = envOrDefault('PEER_ENDPOINT', 'localhost:7051');

// Gateway peer SSL host name override.
export const peerHostAlias = envOrDefault('PEER_HOST_ALIAS', 'peer0.org1.example.com');

/**
 * displayInputParameters() will print the global scope parameters used by the main driver routine.
 */
export async function displayInputParameters(): Promise<void> {
  console.debug('*********************************');
  console.debug('**      INPUT PARAMETERS       **');
  console.debug(`channelName:       ${channelName}`);
  console.debug(`chaincodeName:     ${chaincodeName}`);
  console.debug(`mspId:             ${mspId}`);
  console.debug(`cryptoPath:        ${cryptoPath}`);
  console.debug(`tlsCertPath:       ${tlsCertPath}`);
  console.debug(`peerEndpoint:      ${peerEndpoint}`);
  console.debug(`peerHostAlias:     ${peerHostAlias}`);
  console.debug('*********************************');
}

export async function newGrpcConnection(): Promise<grpc.Client> {
  const tlsRootCert = await fs.readFile(tlsCertPath);
  const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);
  return new grpc.Client(peerEndpoint, tlsCredentials, {
    'grpc.ssl_target_name_override': peerHostAlias,
  });
}

export function newIdentity(cert: string):Identity {
  const credentials = Buffer.from(cert);
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

export function keysMatch(key:string, cert: string): boolean|string {
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
