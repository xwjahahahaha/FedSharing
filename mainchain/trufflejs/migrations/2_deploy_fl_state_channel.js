const fs = require('fs');
const iconv = require('iconv-lite');
const FL_StateChannel = artifacts.require("FL_StateChannel");
const FSCoin = artifacts.require("FSCoin");
const flConfigFilePath = "../../configs/fl_conf.json";
const hostConfigFilePath = "../../configs/host_conf.json";


module.exports = function(deployer) {
    // 前置工作：读取配置文件
    var flJson = loadjson(flConfigFilePath);
    var hostJson = loadjson(hostConfigFilePath);
    minerAddrList = []
    for (var i = 0; i < hostJson.miners.length; i++) {
        minerAddrList.push(hostJson.miners[i].ether_account);
    }
    // 解锁账户权限
    web3.eth.personal.unlockAccount(hostJson.pool_manager.ether_account, "test password!", 1000).then(console.log('Pool manager account unlocked!'));
    web3.eth.personal.unlockAccount(hostJson.service_demander.ether_account, "test password!", 1000).then(console.log('Service demander account unlocked!'));
    // 部署FL_StateChannel合约
    console.log("FSCoin contract already deployed Address : ", hostJson.network.fscoin_contract.address);
    // var flstatechannel, fscoin;
    deployer.deploy(FL_StateChannel,
            hostJson.network.fscoin_contract.address,
            flJson.task.id,
            flJson.model_name, ["batch_size:" + flJson.batch_size.toString(), "lr:" + flJson.lr.toString(), "momentum:" + flJson.momentum.toString(), "lambda:" + flJson.lambda.toString()], [flJson.global_epochs, flJson.local_epochs],
            flJson.task.v,
            flJson.task.proportion,
            minerAddrList,
            hostJson.service_demander.ether_account, [hostJson.state_channel.expiretion.start, hostJson.state_channel.expiretion.end], { overwrite: true, gas: 9999999, from: hostJson.pool_manager.ether_account, gasPrice: 100000 })
        .then(async() => {
            let fscoinInstance = await FSCoin.deployed();
            // approve
            await fscoinInstance.approve(FL_StateChannel.address, flJson.task.v, { from: hostJson.service_demander.ether_account, gas: 99999, gasPrice: 10000 })
                .then((res) => { console.log("Service demander success approved, tx is ", res.tx.toString()) }).catch(err => console.log(err));
            await fscoinInstance.approve(FL_StateChannel.address, flJson.task.v, { from: hostJson.pool_manager.ether_account, gas: 99999, gasPrice: 10000 })
                .then((res) => { console.log("Pool manager success approved, tx is ", res.tx.toString()) }).catch(err => console.log(err));
            console.log("All accounts were approved.");
            // deposit
            let flStateChannel = await FL_StateChannel.deployed();
            await flStateChannel.depositV({ from: hostJson.service_demander.ether_account, gas: 99999999, gasPrice: 100000 })
                .then((res) => { console.log(`Service demander success deposit ${flJson.task.v} fscoin, tx is ${res.tx.toString()}`); }).catch(err => console.log(err));
            await flStateChannel.deposit({ from: hostJson.pool_manager.ether_account, gas: 99999999, gasPrice: 100000 })
                .then((res) => { console.log(`Pool manager success deposit ${flJson.task.v} fscoin, tx is ${res.tx.toString()}`); }).catch(err => console.log(err));
            let pm_deposit = await flStateChannel.deposit_value({ from: hostJson.pool_manager.ether_account, gas: 99999999, gasPrice: 100000 }).catch(err => console.log(err));
            console.log(`The actual deposit amount of Pool manager is ${pm_deposit} fscoin.`);
            let state = await flStateChannel.channel_state({ from: hostJson.pool_manager.ether_account, gas: 99999999, gasPrice: 100000 }).catch(err => console.log(err));
            if (state == 1) {
                console.log(`Now FL_State_Channel state is [Deposited].`);
                console.log("success deposited.");
            } else {
                throw new Error(`Deposit error, now FL_State_Channel state is ${state}, Please try again.`);
            }
        });
}

function loadjson(filepath) {
    var data;
    try {
        var jsondata = iconv.decode(fs.readFileSync(filepath, "binary"), "utf8");
        data = JSON.parse(jsondata);
    } catch (err) {
        console.log(err);
    }

    return data;
}

function savejson(filepath, data) {
    var datastr = JSON.stringify(data, null, 4);
    if (datastr) {
        try {
            fs.writeFileSync(filepath, datastr);
        } catch (err) {}
    }
}