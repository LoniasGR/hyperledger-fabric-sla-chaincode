import * as grpc from '@grpc/grpc-js';
import { connect, Gateway, Contract } from '@hyperledger/fabric-gateway';
import 'dotenv/config';
import express from 'express';
import cors from 'cors';

import axios from 'axios';
import { normalize } from 'path';
import * as utils from './utils';
import * as constants from './constants';
import {
  queryUsersByPublicKey, queryVRUTimeRange, queryPartsTimeRange, UserData,
} from './ledger';

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
  org?: number,
  username?: string,
};

async function initialize(key:string, cert:string, org: number): Promise<Connections> {
  console.log(org);
  await utils.displayInputParameters(org);

  const {
    mspId, tlsCertPath, peerEndpoint, peerHostAlias,
  } = constants.orgConstants[org - 1];

  // The gRPC client connection should be shared by all Gateway connections to this endpoint.
  const grpcClient = await utils.newGrpcConnection(
    tlsCertPath,
    peerEndpoint,
    peerHostAlias,
  );

  const gateway = connect({
    client: grpcClient,
    identity: utils.newIdentity(cert, mspId),
    signer: utils.newSigner(key),
    // Default timeouts for different gRPC calls
    evaluateOptions: () => ({ deadline: Date.now() + 5000 }), // 5 seconds
    endorseOptions: () => ({ deadline: Date.now() + 15000 }), // 15 seconds
    submitOptions: () => ({ deadline: Date.now() + 5000 }), // 5 seconds
    commitStatusOptions: () => ({ deadline: Date.now() + 60000 }), // 1 minute
  });

  return { gateway, grpcClient };
}

async function checkAndInitializeKeys(key: string, cert: string): Promise<GatewayOrError> {
  const keysWithStatus = utils.verifyKeys(key, cert);

  if (keysWithStatus.success === false) {
    const { success, error } = keysWithStatus;
    return { error: { success, error } };
  }

  const result = await axios.post(`${constants.userManagementServiceURL}/exists`, { cert });
  console.debug(result.data);
  if (result.data.success === false) {
    return { error: { success: false, error: result.data.error } };
  }

  if (result.data.exists === false) {
    return { error: { success: false, error: 'User does not exist' } };
  }
  const { organisation, username } = result.data;

  const { keyPEM, certPEM } = keysWithStatus;
  const { gateway, grpcClient } = await initialize(keyPEM, certPEM, organisation);
  return {
    gateway: {
      gateway, grpcClient, keyPEM, certPEM,
    },
    org: organisation,
    username,
  };
}

app.post('/init', async (req, res) => {
  const { key, cert } = req.body;
  const gatewayOrError : GatewayOrError = await checkAndInitializeKeys(key, cert);

  if (gatewayOrError.error !== undefined) {
    const { success, error } = gatewayOrError.error;
    return res.send({ success, error });
  }

  return res.send({
    success: true,
    organisation: gatewayOrError.org,
    username: gatewayOrError.username,
  });
});

app.post('/balance', async (req, res) => {
  const { key, cert } = req.body;
  const gatewayOrError = await checkAndInitializeKeys(key, cert);
  if (gatewayOrError.error !== undefined) {
    const { success, error } = gatewayOrError.error;
    return res.send({ success, error });
  }

  if (gatewayOrError.org !== 1) {
    return res.send({ success: false, error: 'User does not exist in this ledger' });
  }

  const gt: GatewayAndKeys = gatewayOrError.gateway!;
  const {
    gateway, grpcClient, certPEM,
  } = gt;

  try {
    // Get a network instance representing the channel where the smart contract is deployed.
    const network = gateway.getNetwork(constants.SLAChannelName);

    // Get the smart contract from the network.
    const contract = network.getContract(constants.SLAChaincodeName);

    // Get the asset details by assetID.
    const userOrError = await queryUsersByPublicKey(contract, utils.oneLiner(certPEM));
    if (typeof userOrError !== 'object') {
      return res.send({ success: false, error: userOrError });
    }
    return res.send({ success: true, user: userOrError });
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
    return res.send({ success, error });
  }

  if (gatewayOrError.org !== 2) {
    return res.send({ success: false, error: 'User does not exist in this ledger' });
  }

  const gt: GatewayAndKeys = gatewayOrError.gateway!;
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
    return res.send({ success, error });
  }

  if (gatewayOrError.org !== 3) {
    return res.send({ success: false, error: 'User does not exist in this ledger' });
  }

  const gt: GatewayAndKeys = gatewayOrError.gateway!;
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

app.post('/balance-sla2', async (req, res) => {
  const { key, cert } = req.body;
  const gatewayOrError = await checkAndInitializeKeys(key, cert);
  if (gatewayOrError.error !== undefined) {
    const { success, error } = gatewayOrError.error;
    return res.send({ success, error });
  }

  if (gatewayOrError.org !== 4) {
    return res.send({ success: false, error: 'User does not exist in this ledger' });
  }

  let ccs: Array<string> = [];
  try {
    ccs = await utils.connectWithPeer();
    console.log(`Chaincodes: ${ccs}`);
  } catch (e: any) {
    console.error(e);
    res.status(e).send({ success: false, error: 'Could not connect with peer' });
  }

  const gt: GatewayAndKeys = gatewayOrError.gateway!;
  const {
    gateway, grpcClient, certPEM,
  } = gt;

  try {
    const errors: Array<string> = [];

    // Get a network instance representing the channel where the smart contract is deployed.
    const network = gateway.getNetwork(constants.SLA2ChannelName);

    const contracts: Array<Contract | null> = ccs.map((cc) => {
      try {
        // Get the smart contract from the network.
        return network.getContract(cc);
      } catch {
        console.error(`Contract ${cc} not found`);
        return null;
      }
    });
    const actualContracts: Array<Contract> = contracts.filter(
      (contract): contract is Contract => contract !== null,
    );
    if (actualContracts.length === 0) {
      return res.send({ success: false, error: 'No contracts found in channel' });
    }

    const user = await Promise.all(
      actualContracts.map(async (contract) => {
        const userOrError = await queryUsersByPublicKey(contract, utils.oneLiner(certPEM));
        if (typeof userOrError === 'string') {
          console.error(userOrError);
          errors.push(userOrError);
          return null;
        }
        return userOrError as UserData;
      }),
    );

    const actualUser = user.filter((u): u is UserData => u !== null)
      .reduce((prevUser, u) => ({
        id: prevUser.id,
        name: prevUser.name,
        balance: prevUser.balance + u.balance,
      }));

    return res.send({ success: true, user: actualUser });
  } finally {
    gateway.close();
    grpcClient.close();
  }
});

app.listen(constants.expressPort, () => {
  console.debug(`⚡️[server]: Server is running at https://localhost:${constants.expressPort}`);
});
