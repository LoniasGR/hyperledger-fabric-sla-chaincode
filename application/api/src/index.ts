import * as grpc from '@grpc/grpc-js';
import pkg from '@hyperledger/fabric-gateway';
import 'dotenv/config';
import express from 'express';
import * as utils from './utils';

const app = express();


async function initialize(): Promise<void> {
    await utils.displayInputParameters();

    // The gRPC client connection should be shared by all Gateway connections to this endpoint.
    const client = await utils.newGrpcConnection();

    app.listen(utils.expressPort, () => {
        console.debug(`⚡️[server]: Server is running at https://localhost:${utils.expressPort}`);
      });
}
