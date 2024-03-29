import * as path from 'path';

function envOrDefault(key: string, defaultValue: string): string {
  return process.env[key] || defaultValue;
}

export const expressPort = envOrDefault('EXPRESS_PORT', '8000');
export const userManagementServiceURL = envOrDefault('identity_endpoint', 'http://localhost:8000');

export const SLAChannelName = envOrDefault('fabric_sla_channel', 'sla');
export const SLAChaincodeName = envOrDefault('fabric_sla_contract', 'slasc_bridge');

export const VRUChannelName = envOrDefault('fabric_vru_channel', 'vru');
export const VRUChaincodeName = envOrDefault('fabric_vru_contract', 'vru_positions');

export const PartsChannelName = envOrDefault('fabric_parts_channel', 'parts');
export const PartsChaincodeName = envOrDefault('fabric_parts_contract', 'parts');

export const SLA2ChannelName = envOrDefault('fabric_sla2_channel', 'sla2.0');

export const SLA2Peer = envOrDefault('sla2_peer', 'grpc://org4-peer1:8051');

export const org1MSPId = envOrDefault('ORG1_MSP_ID', 'Org1MSP');
export const org2MSPId = envOrDefault('ORG2_MSP_ID', 'Org2MSP');
export const org3MSPId = envOrDefault('ORG3_MSP_ID', 'Org3MSP');
export const org4MSPId = envOrDefault('ORG4_MSP_ID', 'Org4MSP');

// Path to peer tls certificate.
export const org1TlsCertPath = envOrDefault('ORG1_TLS_CERT_PATH', '/fabric/tlscacerts/org1/tlsca-signcert.pem');
export const org2TlsCertPath = envOrDefault('ORG2_TLS_CERT_PATH', '/fabric/tlscacerts/org2/tlsca-signcert.pem');
export const org3TlsCertPath = envOrDefault('ORG3_TLS_CERT_PATH', '/fabric/tlscacerts/org3/tlsca-signcert.pem');
export const org4TlsCertPath = envOrDefault('ORG4_TLS_CERT_PATH', '/fabric/tlscacerts/org4/tlsca-signcert.pem');

// Gateway peer endpoint.
export const org1PeerEndpoint = envOrDefault('fabric_org1_gateway_hostport', 'localhost:7051');
export const org2PeerEndpoint = envOrDefault('fabric_org2_gateway_hostport', 'localhost:9051');
export const org3PeerEndpoint = envOrDefault('fabric_org3_gateway_hostport', 'localhost:11051');
export const org4PeerEndpoint = envOrDefault('fabric_org4_gateway_hostport', 'localhost:13051');

// Gateway peer SSL host name override.
export const org1PeerHostAlias = envOrDefault('fabric_org1_gateway_sslHostOverride', 'peer0.org1.example.com');
export const org2PeerHostAlias = envOrDefault('fabric_org2_gateway_sslHostOverride', 'peer0.org2.example.com');
export const org3PeerHostAlias = envOrDefault('fabric_org3_gateway_sslHostOverride', 'peer0.org3.example.com');
export const org4PeerHostAlias = envOrDefault('fabric_org4_gateway_sslHostOverride', 'peer0.org4.example.com');

export const org1Constants = {
  mspId: org1MSPId,
  tlsCertPath: org1TlsCertPath,
  peerEndpoint: org1PeerEndpoint,
  peerHostAlias: org1PeerHostAlias,
};

export const org2Constants = {
  mspId: org2MSPId,
  tlsCertPath: org2TlsCertPath,
  peerEndpoint: org2PeerEndpoint,
  peerHostAlias: org2PeerHostAlias,
};

export const org3Constants = {
  mspId: org3MSPId,
  tlsCertPath: org3TlsCertPath,
  peerEndpoint: org3PeerEndpoint,
  peerHostAlias: org3PeerHostAlias,
};

export const org4Constants = {
  mspId: org4MSPId,
  tlsCertPath: org4TlsCertPath,
  peerEndpoint: org4PeerEndpoint,
  peerHostAlias: org4PeerHostAlias,
};

export const orgConstants = [
  org1Constants,
  org2Constants,
  org3Constants,
  org4Constants,
];
