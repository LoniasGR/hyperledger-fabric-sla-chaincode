import * as path from 'path';

function envOrDefault(key: string, defaultValue: string): string {
  return process.env[key] || defaultValue;
}
export const expressPort = envOrDefault('EXPRESS_PORT', '8000');

export const SLAChannelName = envOrDefault('SLA_CHANNEL_NAME', 'sla');
export const SLAChaincodeName = envOrDefault('SLA_CHAINCODE_NAME', 'slasc_bridge');

export const VRUChannelName = envOrDefault('VRU_CHANNEL_NAME', 'vru');
export const VRUChaincodeName = envOrDefault('VRU_CHAINCODE_NAME', 'vru_positions');

export const PartsChannelName = envOrDefault('PARTS_CHANNEL_NAME', 'parts');
export const PartsChaincodeName = envOrDefault('PARTS_CHAINCODE_NAME', 'parts');

export const mspId = envOrDefault('MSP_ID', 'Org1MSP');

// Path to crypto materials.
export const cryptoPath = envOrDefault('CRYPTO_PATH', path.resolve(__dirname, '..', '..', '..', 'organizations', 'peerOrganizations', 'org1.example.com'));

// Path to peer tls certificate.
export const tlsCertPath = envOrDefault('TLS_CERT_PATH', path.resolve(cryptoPath, 'peers', 'peer0.org1.example.com', 'tls', 'ca.crt'));

// Gateway peer endpoint.
export const peerEndpoint = envOrDefault('PEER_ENDPOINT', 'localhost:7051');

// Gateway peer SSL host name override.
export const peerHostAlias = envOrDefault('PEER_HOST_ALIAS', 'peer0.org1.example.com');
