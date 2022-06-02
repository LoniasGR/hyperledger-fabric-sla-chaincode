import * as grpc from '@grpc/grpc-js';
import { connect, Gateway, GatewayError } from '@hyperledger/fabric-gateway';
import 'dotenv/config';
import express from 'express';
import cors from 'cors';

import * as utils from './utils';
import * as constants from './constants';
import { queryUsersByPublicKey, queryVRUTimeRange, queryPartsTimeRange } from './ledger';

const app = express();
app.use(express.json()); // for parsing application/json
app.use(cors());

interface Connections {
  gateway: Gateway,
  grpcClient: grpc.Client,
}

type Error = {
  success: boolean,
  error: string,
};

type GatewayAndKeys = {
  gateway: Gateway,
  grpcClient: grpc.Client,
  keyPEM: string,
  certPEM: string,
};

type GatewayOrError = {
  error?: Error,
  gateway?: GatewayAndKeys,
};

async function initialize(key:string, cert:string): Promise<Connections> {
  await utils.displayInputParameters();

  // The gRPC client connection should be shared by all Gateway connections to this endpoint.
  const client = await utils.newGrpcConnection();
  const gateway = connect({
    client,
    identity: utils.newIdentity(cert),
    signer: utils.newSigner(key),
    // Default timeouts for different gRPC calls
    evaluateOptions: () => ({ deadline: Date.now() + 5000 }), // 5 seconds
    endorseOptions: () => ({ deadline: Date.now() + 15000 }), // 15 seconds
    submitOptions: () => ({ deadline: Date.now() + 5000 }), // 5 seconds
    commitStatusOptions: () => ({ deadline: Date.now() + 60000 }), // 1 minute
  });

  return { gateway, grpcClient: client };
}

async function checkAndInitializeKeys(key: string, cert: string): Promise<GatewayOrError> {
  const keysWithStatus = utils.verifyKeys(key, cert);

  if (keysWithStatus.success === false) {
    const { success, error } = keysWithStatus;
    return { error: { success, error } };
  }

  const { keyPEM, certPEM } = keysWithStatus;
  const { gateway, grpcClient } = await initialize(keyPEM, certPEM);
  return {
    gateway: {
      gateway, grpcClient, keyPEM, certPEM,
    },
  };
}

app.post('/balance', async (req, res) => {
  const { key, cert } = req.body;
  const gatewayOrError = await checkAndInitializeKeys(key, cert);
  if (gatewayOrError.error !== undefined) {
    const { success, error } = gatewayOrError.error;
    res.send({ success, error });
  }
  const gt = gatewayOrError.gateway!;
  const {
    gateway, grpcClient, certPEM,
  } = gt;

  try {
    // Get a network instance representing the channel where the smart contract is deployed.
    const network = gateway.getNetwork(constants.SLAChannelName);

    // Get the smart contract from the network.
    const contract = network.getContract(constants.SLAChaincodeName);

    // Get the asset details by assetID.
    const user = await queryUsersByPublicKey(contract, utils.oneLiner(certPEM));
    return res.send({ success: true, user });
  } finally {
    gateway.close();
    grpcClient.close();
  }
});

app.post('/vru/incidents', async (req, res) => {
  const {
    key, cert, startDate, endDate,
  } = req.body;
  const gatewayOrError = await checkAndInitializeKeys(key, cert);
  if (gatewayOrError.error !== undefined) {
    const { success, error } = gatewayOrError.error;
    res.send({ success, error });
  }
  const gt = gatewayOrError.gateway!;
  const {
    gateway, grpcClient,
  } = gt;

  try {
    // Get a network instance representing the channel where the smart contract is deployed.
    const network = gateway.getNetwork(constants.VRUChannelName);

    // Get the smart contract from the network.
    const contract = network.getContract(constants.VRUChaincodeName);

    // Get the asset details by assetID.
    const assets = await queryVRUTimeRange(contract, startDate, endDate);
    return res.send({ success: true, assets });
  } finally {
    gateway.close();
    grpcClient.close();
  }
});

app.post('/parts', async (req, res) => {
  const {
    key, cert, startDate, endDate,
  } = req.body;

  const gatewayOrError = await checkAndInitializeKeys(key, cert);
  if (gatewayOrError.error !== undefined) {
    const { success, error } = gatewayOrError.error;
    res.send({ success, error });
  }
  const gt = gatewayOrError.gateway!;
  const {
    gateway, grpcClient,
  } = gt;

  try {
    // Get a network instance representing the channel where the smart contract is deployed.
    const network = gateway.getNetwork(constants.PartsChannelName);

    // Get the smart contract from the network.
    const contract = network.getContract(constants.PartsChaincodeName);

    // Get the asset details by assetID.
    const assets = await queryPartsTimeRange(contract, startDate, endDate);
    return res.send({ success: true, assets });
  } finally {
    gateway.close();
    grpcClient.close();
  }
});

app.listen(constants.expressPort, () => {
  console.debug(`⚡️[server]: Server is running at https://localhost:${constants.expressPort}`);
  utils.displayInputParameters();
});
