import * as grpc from '@grpc/grpc-js';
import {
  connect, Contract, Identity, Signer, signers,
} from '@hyperledger/fabric-gateway';
import envOrDefault from './utils';

const channelName = envOrDefault('CHANNEL_NAME', 'mychannel');
const chaincodeName = envOrDefault('CHAINCODE_NAME', 'basic');
const mspId = envOrDefault('MSP_ID', 'Org1MSP');
