const fs = require('fs');
const iconv = require('iconv-lite');
const FSCoin = artifacts.require("FSCoin");
const LibSafeMath = artifacts.require("SafeMath");
const flConfigFilePath = "../configs/fl_conf.json";
const hostConfigFilePath = "../configs/host_conf.json";
const fscoinOwner = "0x67b690d8B2EDbfBF8172FBA5Cd99C5e69Cd09035"
const initTotalSupply = 100000000
const initialBalance = 1000

module.exports = function(deployer) {
    // 前置工作：读取配置文件
    var flJson = loadjson(flConfigFilePath);
    var hostJson = loadjson(hostConfigFilePath);
    if (hostJson.network.fscoin_contract.address != "") {
        return
    }
    minerAddrList = []
    for (var i = 0; i < hostJson.miners.length; i++) {
        minerAddrList.push(hostJson.miners[i].ether_account);
    }
    // 部署SafeMath库
    console.log("Push LibSafeMath...");
    deployer.deploy(LibSafeMath, { overwrite: true, gas: 9999999, from: fscoinOwner, gasPrice: 100000 });
    deployer.link(LibSafeMath, FSCoin);
    // 部署FSCoin合约
    console.log("Push new FSCoin contract...");
    deployer.deploy(FSCoin, initTotalSupply, { overwrite: true, gas: 9999999, from: fscoinOwner, gasPrice: 100000 })
        .then(() => {
            hostJson.network.fscoin_contract.address = FSCoin.address;
            hostJson.network.fscoin_contract.total_supply = initTotalSupply;
            savejson(hostConfigFilePath, hostJson);
            FSCoin.deployed().then((fscoinInstance) => {
                // 初始化各个参与者金额
                promises = []
                for (var i = 0; i < minerAddrList.length; i++) {
                    promises.push(new Promise((resolve, reject) => {
                        fscoinInstance.transfer(minerAddrList[i], initialBalance, { from: fscoinOwner, gas: 9999999, gasPrice: 100000 }).then(
                            fscoinInstance.balances(minerAddrList[i], { from: minerAddrList[i] })
                            .then(res => {
                                resolve(res.toNumber())
                            })
                        ).catch(err => {
                            console.log("err:", err);
                            reject(err);
                        })
                    }))
                }
                promises.push(new Promise((resolve, reject) => {
                    fscoinInstance.transfer(hostJson.pool_manager.ether_account, flJson.task.v * 3, { from: fscoinOwner, gas: 9999999, gasPrice: 100000 }).then(
                        fscoinInstance.balances(hostJson.pool_manager.ether_account, { from: hostJson.pool_manager.ether_account })
                        .then(res => {
                            resolve(res.toNumber())
                        })
                    ).catch(err => {
                        console.log("err:", err);
                        reject(err);
                    })
                }))
                promises.push(new Promise((resolve, reject) => {
                    fscoinInstance.transfer(hostJson.service_demander.ether_account, flJson.task.v * 3, { from: fscoinOwner, gas: 9999999, gasPrice: 100000 }).then(
                        fscoinInstance.balances(hostJson.service_demander.ether_account, { from: hostJson.service_demander.ether_account })
                        .then(res => {
                            resolve(res.toNumber())
                        })
                    ).catch(err => {
                        console.log("err:", err);
                        reject(err);
                    })
                }))
                Promise.all(promises)
                    .then((res) => {
                        console.log("success init accounts.");
                        console.log(res);
                    }).catch(e => console.log(e))
            })
        })
};

function loadjson(filepath) {
    var data;
    try {
        var jsondata = iconv.decode(fs.readFileSync(filepath, "binary"), "utf8");
        data = JSON.parse(jsondata);
    } catch (err) {
        console.log("read json file err : ", err);
    }

    return data;
}

function savejson(filepath, data) {
    var datastr = JSON.stringify(data, null, 4);
    if (datastr) {
        try {
            fs.writeFileSync(filepath, datastr);
        } catch (err) {
            console.log("save json file err : ", err);
        }
    }
}