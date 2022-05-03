import { Contract } from '@hyperledger/fabric-gateway';
import { TextDecoder } from 'util';
import * as errors from './errors';

type UserData = {
  id: string,
  name: string,
  balance: string,
}
const utf8Decoder = new TextDecoder();
/**
 * Function to query if a user with the specified public key exists.
 */
export default async function queryUsersByPublicKey(
  contract: Contract,
  publicKey: string,
): Promise<UserData|string> {
  try {
    console.debug('\n--> Evaluate Trasnaction: QueryUsersByPublicKey');
    const processedPublicKey = publicKey.replace(/(-----(BEGIN|END) CERTIFICATE-----|[\n\r])/g, '');
    const resultBytes = await contract.evaluateTransaction('QueryUsersByPublicKey', `${processedPublicKey}`);

    const resultJson = utf8Decoder.decode(resultBytes);
    const result = JSON.parse(resultJson);
    console.log('*** Result:', result);
    const { id, name, balance } = result;
    return { id, name, balance };
  } catch (e: unknown) {
    console.error(errors.getErrorMessage(e));
    return (errors.getErrorMessage(e));
  }
}
