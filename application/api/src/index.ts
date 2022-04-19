import * as grpc from '@grpc/grpc-js';
import { connect, Gateway } from '@hyperledger/fabric-gateway';
import 'dotenv/config';
import express from 'express';
import cors from 'cors';

import * as utils from './utils';
import queryUsersByPublicKey from './ledger';

const app = express();
app.use(express.json()); // for parsing application/json
app.use(cors());

interface Connections {
  gateway: Gateway,
  grpcClient: grpc.Client,
}

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

app.post('/balance', async (req, res) => {
  const { key, cert } = req.body;
  if (key === undefined) {
    return res.send({ success: false, error: 'Private key is missing' });
  }
  if (cert === undefined) {
    return res.send({ success: false, error: 'Public key is missing' });
  }
  const keyPEM = utils.toPEMFormat(key);
  const certPEM = utils.toPEMFormat(cert);

  const match = utils.keysMatch(keyPEM, certPEM);
  if (typeof match !== 'boolean') {
    return res.send({ success: false, error: match });
  }
  if (!match) {
    return res.send({ success: false, error: 'Public/private key missmatch' });
  }
  const { gateway, grpcClient } = await initialize(keyPEM, certPEM);
  try {
    // Get a network instance representing the channel where the smart contract is deployed.
    const network = gateway.getNetwork(utils.channelName);

    // Get the smart contract from the network.
    const contract = network.getContract(utils.chaincodeName);

    // Get the asset details by assetID.
    const user = await queryUsersByPublicKey(contract, utils.oneLiner(certPEM));
    return res.send({ success: true, user });
  } finally {
    gateway.close();
    grpcClient.close();
  }
});

app.listen(utils.expressPort, () => {
  console.debug(`⚡️[server]: Server is running at https://localhost:${utils.expressPort}`);
});
