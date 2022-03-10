import * as path from 'path';
import * as grpc from '@grpc/grpc-js';
import { promises as fs } from 'fs';

export function envOrDefault(key: string, defaultValue: string): string {
  return process.env[key] || defaultValue;
}

export const expressPort = envOrDefault('EXPRESS_PORT', '8000');

export const channelName = envOrDefault('CHANNEL_NAME', 'mychannel');
export const chaincodeName = envOrDefault('CHAINCODE_NAME', 'basic');
export const mspId = envOrDefault('MSP_ID', 'Org1MSP');

// Path to crypto materials.
export const cryptoPath = envOrDefault('CRYPTO_PATH', path.resolve(__dirname, '..', '..', '..', 'test-network', 'organizations', 'peerOrganizations', 'org1.example.com'));

// Path to user private key directory.
export const keyDirectoryPath = envOrDefault('KEY_DIRECTORY_PATH', path.resolve(cryptoPath, 'users', 'User1@org1.example.com', 'msp', 'keystore'));

// Path to user certificate.
export const certPath = envOrDefault('CERT_PATH', path.resolve(cryptoPath, 'users', 'User1@org1.example.com', 'msp', 'signcerts', 'cert.pem'));

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
  console.debug(`channelName:       ${channelName}`);
  console.debug(`chaincodeName:     ${chaincodeName}`);
  console.debug(`mspId:             ${mspId}`);
  console.debug(`cryptoPath:        ${cryptoPath}`);
  console.debug(`keyDirectoryPath:  ${keyDirectoryPath}`);
  console.debug(`certPath:          ${certPath}`);
  console.debug(`tlsCertPath:       ${tlsCertPath}`);
  console.debug(`peerEndpoint:      ${peerEndpoint}`);
  console.debug(`peerHostAlias:     ${peerHostAlias}`);
}

export async function newGrpcConnection(): Promise<grpc.Client> {
  const tlsRootCert = await fs.readFile(tlsCertPath);
  const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);
  return new grpc.Client(peerEndpoint, tlsCredentials, {
    'grpc.ssl_target_name_override': peerHostAlias,
  });
}
