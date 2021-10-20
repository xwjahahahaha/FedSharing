var Web3 = require('web3');
var abi = require('ethereumjs-abi');

var web3 = new Web3(Web3.givenProvider || 'http://localhost:8545');

console.log(web3.version);

poolManager = web3.eth.accounts.privateKeyToAccount("CB760ECDD8C99E0008B982193ECF936E5250FF45ECC8359F8BA46838D2FFF005");
miner1 = web3.eth.accounts.privateKeyToAccount("71F3BFA9B3BC97B7733CE17FB7B941B8AF9CC22ACE5402EFCB779791F21BEA0F");
miner2 = web3.eth.accounts.privateKeyToAccount("4C6A8B7CF974FDD76F53A5028F4B68FFBF60E981B9DD764C9E10F0ED6E9A968E");


contractAddress = "0xa0310dDBd9D5019421f2EB578b30C9066c1b8406";

function signFLState(poolManager, minerAddress, amount, contractAddress) {
    console.log("联邦学习账户:", minerAddress);
    var hash = "0x" + abi.soliditySHA3(
        ["address", "uint256", "address"], [contractAddress, amount, minerAddress]
    ).toString("hex");
    console.log("hash = ", hash)

    console.log("Pool Manager:", poolManager)
    web3.eth.personal.sign(hash, poolManager, (err, res) => {
        if (!err) {
            console.log(res)
        } else {
            console.log(err)
        }
    });
}

signFLState(poolManager.address, miner1.address, 90, contractAddress)