import { Contract } from '@hyperledger/fabric-gateway';
import { TextDecoder } from 'util';
import * as errors from './errors';

type UserData = {
  id: string,
  name: string,
  balance: string,
};

type PartsData = {
  total: number,
  high_quality: number,
  low_quality: number,
};

type VRUData = {
  critical: number,
  warning: number,
  highRisk: number,
  lowRisk: number,
  noRisk: number,
}

const utf8Decoder = new TextDecoder();
/**
 * Function to query if a user with the specified public key exists.
 */
export async function queryUsersByPublicKey(
  contract: Contract,
  publicKey: string,
): Promise<UserData | string> {
  try {
    console.debug('\n--> Evaluate Transaction: QueryUsersByPublicKey');
    const processedPublicKey = publicKey.replace(/(-----(BEGIN|END) CERTIFICATE-----|[\n\r])/g, '');
    const resultBytes = await contract.evaluateTransaction('QueryUsersByPublicKey', `${processedPublicKey}`);

    const resultJson = utf8Decoder.decode(resultBytes);
    const result = JSON.parse(resultJson);
    console.log('*** Result:', result);
    const { id, name, balance } = result;
    if (id === '' || name === '' || balance === '') {
      return errors.getErrorMessage('User does not exist.');
    }
    return { id, name, balance };
  } catch (e: unknown) {
    console.error(errors.getErrorMessage(e));
    return (errors.getErrorMessage(e));
  }
}

export async function queryVRUTimeRange(
  contract: Contract,
  start: string,
  end: string,
): Promise<VRUData | string> {
  try {
    console.debug('\n--> Evaluate Transaction: queryVRUTimeRange');
    const resultBytes = await contract.evaluateTransaction('GetAssetRiskInRange', `${start}`, `${end}`);
    const resultJson = utf8Decoder.decode(resultBytes);
    const result = JSON.parse(resultJson);
    console.log('*** Result:', result);
    return result;
  } catch (e: unknown) {
    console.error(errors.getErrorMessage(e));
    return (errors.getErrorMessage(e));
  }
}

export async function queryPartsTimeRange(
  contract: Contract,
  start: string,
  end: string,
): Promise<PartsData | string> {
  try {
    console.debug('\n--> Evaluate Transaction: queryPartsTimeRange');
    const resultBytes = await contract.evaluateTransaction('GetAssetQualityByRange', `${start}`, `${end}`);
    const resultJson = utf8Decoder.decode(resultBytes);
    const result = JSON.parse(resultJson);
    console.log('*** Result:', result[0]);
    return result[0];
  } catch (e: unknown) {
    console.error(errors.getErrorMessage(e));
    return (errors.getErrorMessage(e));
  }
}
