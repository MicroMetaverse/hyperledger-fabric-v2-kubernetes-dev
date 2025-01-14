import { Gateway, Wallets } from 'fabric-network';
import * as path from 'path';
import * as fs from 'fs';
import { FunctionDto } from 'src/dto';

export const invoke = async (invoke: FunctionDto) => {
  let msg = ``;

  try {
    // Load the network configuration
    const ccpPath = path.resolve(__dirname, process.env.HLF_CCP_PATH);
    const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

    // Create a new file system based wallet for managing identities.
    const walletPath = path.join(process.cwd(), 'wallet');
    const wallet = await Wallets.newFileSystemWallet(walletPath);
    console.log(`Wallet path: ${walletPath}`);

    // Check to see if we've already enrolled the user.
    const identity = await wallet.get(invoke.appUser);
    if (!identity) {
      msg = `An identity for the user ${invoke.appUser} does not exist in the wallet;Please register ther user ${invoke.appUser} first`;
      console.log(msg);
      return msg;
    }

    // Create a new gateway for connecting to our peer node.
    const gateway = new Gateway();
    await gateway.connect(ccp, {
      wallet,
      identity: invoke.appUser,
      discovery: { enabled: true, asLocalhost: false },
    });

    // Get the network (channel) our contract is deployed to.
    const network = await gateway.getNetwork(invoke.channelId);

    // Get the contract from the network.
    const contract = network.getContract(invoke.contractName);

    // Submit transaction
    const responseAsBytes = await contract.submitTransaction(
      invoke.func,
      ...invoke.args,
    );
    console.log(`Transaction has been submitted`);
    return responseAsBytes.toString();
  } catch (error) {
    msg = `Failed to submit transaction: ${error}`;
    console.error(msg);
    return msg;
  }
};
