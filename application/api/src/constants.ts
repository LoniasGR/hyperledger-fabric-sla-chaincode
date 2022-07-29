import * as path from 'path';

function envOrDefault(key: string, defaultValue: string): string {
  return process.env[key] || defaultValue;
}

export const expressPort = envOrDefault('EXPRESS_PORT', '8000');
export const userManagementServiceURL = envOrDefault('USER_MANAGEMENT_URL', 'http://localhost:3000');

export const SLAChannelName = envOrDefault('SLA_CHANNEL_NAME', 'sla');
export const SLAChaincodeName = envOrDefault('SLA_CHAINCODE_NAME', 'slasc_bridge');

export const VRUChannelName = envOrDefault('VRU_CHANNEL_NAME', 'vru');
export const VRUChaincodeName = envOrDefault('VRU_CHAINCODE_NAME', 'vru_positions');

export const PartsChannelName = envOrDefault('PARTS_CHANNEL_NAME', 'parts');
export const PartsChaincodeName = envOrDefault('PARTS_CHAINCODE_NAME', 'parts');

export const org1MSPId = envOrDefault('ORG1_MSP_ID', 'Org1MSP');
export const org2MSPId = envOrDefault('ORG2_MSP_ID', 'Org2MSP');
export const org3MSPId = envOrDefault('ORG3_MSP_ID', 'Org3MSP');

// Path to crypto materials.
export const org1CryptoPath = envOrDefault('ORG1_CRYPTO_PATH', path.resolve(__dirname, '..', '..', '..', '..', 'test-network', 'organizations', 'peerOrganizations', 'org1.example.com'));
export const org2CryptoPath = envOrDefault('ORG2_CRYPTO_PATH', path.resolve(__dirname, '..', '..', '..', '..', 'test-network', 'organizations', 'peerOrganizations', 'org2.example.com'));
export const org3CryptoPath = envOrDefault('ORG3_CRYPTO_PATH', path.resolve(__dirname, '..', '..', '..', '..', 'test-network', 'organizations', 'peerOrganizations', 'org3.example.com'));

// Path to peer tls certificate.
export const org1TlsCertPath = envOrDefault('ORG1_TLS_CERT_PATH', path.resolve(org1CryptoPath, 'peers', 'peer0.org1.example.com', 'tls', 'ca.crt'));
export const org2TlsCertPath = envOrDefault('ORG2_TLS_CERT_PATH', path.resolve(org2CryptoPath, 'peers', 'peer0.org2.example.com', 'tls', 'ca.crt'));
export const org3TlsCertPath = envOrDefault('ORG3_TLS_CERT_PATH', path.resolve(org3CryptoPath, 'peers', 'peer0.org3.example.com', 'tls', 'ca.crt'));

// Gateway peer endpoint.
export const org1PeerEndpoint = envOrDefault('ORG1_PEER_ENDPOINT', 'localhost:7051');
export const org2PeerEndpoint = envOrDefault('ORG2_PEER_ENDPOINT', 'localhost:9051');
export const org3PeerEndpoint = envOrDefault('ORG3_PEER_ENDPOINT', 'localhost:11051');

// Gateway peer SSL host name override.
export const org1PeerHostAlias = envOrDefault('ORG1_PEER_HOST_ALIAS', 'peer0.org1.example.com');
export const org2PeerHostAlias = envOrDefault('ORG2_PEER_HOST_ALIAS', 'peer0.org2.example.com');
export const org3PeerHostAlias = envOrDefault('ORG3_PEER_HOST_ALIAS', 'peer0.org3.example.com');

export const org1Constants = {
  mspId: org1MSPId,
  cryptoPath: org1CryptoPath,
  tlsCertPath: org1TlsCertPath,
  peerEndpoint: org1PeerEndpoint,
  peerHostAlias: org1PeerHostAlias,
};

export const org2Constants = {
  mspId: org2MSPId,
  cryptoPath: org2CryptoPath,
  tlsCertPath: org2TlsCertPath,
  peerEndpoint: org2PeerEndpoint,
  peerHostAlias: org2PeerHostAlias,
};

export const org3Constants = {
  mspId: org3MSPId,
  cryptoPath: org3CryptoPath,
  tlsCertPath: org3TlsCertPath,
  peerEndpoint: org3PeerEndpoint,
  peerHostAlias: org3PeerHostAlias,
};

export const orgConstants = [
  org1Constants,
  org2Constants,
  org3Constants,
];
