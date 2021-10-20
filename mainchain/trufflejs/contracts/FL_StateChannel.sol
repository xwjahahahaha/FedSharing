// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.7.6 <0.9.0;

import './SafeMath.sol';
import './FSCoin.sol';

contract FL_StateChannel {
    
    using SafeMath for uint;
    
    address public FSCoinAddr;
    
    // task
    uint public taskID; 
    string public model_name; 
    string[] public super_paras; 
    uint public global_epoch; 
    uint public local_epoch; 
    uint public V;
    uint[4] public proportion;
    uint internal sum;
    
    // character
    address public service_demander;
    address public pool_manager;
    address[] public participates;
    
    // training reward
    bool public isConversion;
    mapping(address => bool) public conversioned;
    uint public conversioned_count;
    uint precision = 10000;
    
    // deposit
    bool public depositV_flag;
    uint public deposit_value;
    
    // state channel over time
    enum Channel_State {inited, deposited, openConversion, closeConversion, modelUploaded, modelCommitted, channelClosed}
    Channel_State public channel_state;        
    uint public expiration_start;
    uint public expiration_end;
    
    // model
    bytes[] internal model_src;
    string internal file_format;
    
    // spvPoof
    bool public destoried;
    
    event DepositV(uint taskID, address service_demander, uint amount);
    event DestorySideReward(uint taskID, uint timestamp, uint amount);
    event Deposit(uint taskID, address pool_manager, uint amount);
    event Conversion(uint taskID, address participate, uint amount);
    event UploadModel(uint taskID, address pool_manager, uint timestamp);
    event CommitModel(uint taskID, address service_demander, uint timestamp);
    event CloseChannel(uint taskID, address pool_manager, uint timestamp);
    
    modifier onlyManager{
        require(msg.sender == pool_manager, "Only pool manager can call this function.");
        _;
    }
    
    modifier onlyDemander{
        require(msg.sender == service_demander, "Only service demander can call this function.");
        _;
    }

    constructor(
        address _FSCoinAddr,
        uint _taskID,
        string memory _model_name,
        string[] memory _super_paras, 
        uint[2] memory _epochs,
        uint _v,
        uint[4] memory _proportion, 
        address[] memory _participates, 
        address _service_demander, 
        uint[2] memory _expiration
    ){
        require(_expiration[0] < _expiration[1], "Invalid expiration settings.");
        FSCoinAddr = _FSCoinAddr;
        taskID = _taskID;
        model_name = _model_name;
        super_paras = _super_paras;
        global_epoch = _epochs[0];
        local_epoch = _epochs[1];
        V = _v;
        proportion = _proportion;
        for (uint i=0; i<_proportion.length; i++){
            sum = sum.add(proportion[i]);
        }
        service_demander = _service_demander;
        pool_manager = msg.sender;
        participates = _participates;
        for (uint i=0; i<_participates.length; i++) {
            conversioned[participates[i]] = false;
        }
        expiration_start = block.timestamp + _expiration[0];
        expiration_end = block.timestamp + _expiration[1];
        channel_state = Channel_State.inited;
    }
    
    function depositV() public onlyDemander {
        FSCoin(FSCoinAddr).transferFrom(msg.sender, address(this), V);
        depositV_flag = true;
        emit DepositV(taskID, msg.sender, V);
        // destroy remaining SideReward FSCoin, because there are cross to side-blockchain.
        uint side_reward = proportion_count(2);
        FSCoin(FSCoinAddr).transfer(address(0), side_reward);
        destoried = true;
        emit DestorySideReward(taskID, block.timestamp, side_reward);
    }
    
    function deposit() public onlyManager {
        require(depositV_flag, "Need to pledge V first.");
        deposit_value = proportion_count(3);
        FSCoin(FSCoinAddr).transferFrom(msg.sender, address(this), deposit_value);
        channel_state = Channel_State.deposited;
        emit Deposit(taskID, msg.sender, deposit_value);
    }
    
    function contractBal() public view returns(uint){
        return FSCoin(FSCoinAddr).balanceOf(address(this));
    }
    
    function proportion_count(uint rewardKind) internal view returns(uint) {
        require(rewardKind < 4 && rewardKind >= 0, "Invalid rewardKind input.");
        uint propo = proportion[rewardKind].mul(precision).div(sum);
        return propo.mul(V).div(precision); 
    }
    
    function isValidSignature(uint256 amount, bytes memory signature) internal view returns(bool) {
        bytes32 message = prefixed(keccak256(abi.encodePacked(this, amount, msg.sender)));
        return recoverSigner(message, signature) == pool_manager;
    }
    
    function prefixed(bytes32 hash) internal pure returns (bytes32) {
        return keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", hash));
    }

    function recoverSigner(bytes32 message, bytes memory signature) internal pure returns (address) {
        (uint8 v, bytes32 r, bytes32 s) = splitSignature(signature);
        return ecrecover(message, v, r, s);
    }

    function splitSignature(bytes memory signature) internal pure returns (uint8 v, bytes32 r, bytes32 s) {
        require(signature.length == 65);
        assembly {
            r := mload(add(signature, 32))
            s := mload(add(signature, 64))
            v := byte(0, mload(add(signature, 96)))
        }
        return (v, r, s);
    }
    
    function openConversion() public onlyManager {
        require(channel_state == Channel_State.deposited, "Pledge a ratio of V first.");
        isConversion = true;
        channel_state = Channel_State.openConversion;
    }
    
    function extendConversionEnd(uint extendTime) public onlyManager {
        require(block.timestamp <= expiration_end, "Conversion period is end.");
        expiration_end = expiration_end.add(extendTime);
    }
    
    function conversion(uint participate_id, uint amount, bytes memory signature) public {
        require(block.timestamp <= expiration_end, "Conversion period is end.");
        require(block.timestamp >= expiration_start || isConversion, "Conversion period has't started.");
        channel_state = Channel_State.openConversion;
        require(participates[participate_id] == msg.sender, "Your are not participate or invalid id input");
        require(!conversioned[msg.sender], "Your already conversioned.");
        require(isValidSignature(amount, signature));
        FSCoin(FSCoinAddr).transfer(msg.sender, amount);
        emit Conversion(taskID, msg.sender, amount);
        conversioned_count ++;
        if (conversioned_count == participates.length) {
            channel_state = Channel_State.closeConversion;
        }
    }
    
    function uploadModel(bytes[] memory _model_src, string memory _file_format) public onlyManager {
        require(channel_state == Channel_State.closeConversion, "Wait all of participate convert");
        model_src = _model_src;
        file_format = _file_format;
        channel_state == Channel_State.modelUploaded;
        emit UploadModel(taskID, msg.sender, block.timestamp);
    }
    
    function getModelData() public view onlyDemander returns(bytes[] memory, string memory){
        require(channel_state == Channel_State.modelUploaded, "Wait pool manager uploaded data");
        return (model_src, file_format);
    }
    
    function commitModel() public onlyDemander {
        require(channel_state == Channel_State.modelUploaded, "Wait pool manager uploaded data");
        channel_state = Channel_State.modelCommitted;
        emit CommitModel(taskID, msg.sender, block.timestamp);
    }
    
    function closeChannel() public onlyManager {
        require(channel_state == Channel_State.modelCommitted, "Wait server demander commit model data");
        FSCoin(FSCoinAddr).transfer(msg.sender, deposit_value);
        channel_state = Channel_State.channelClosed;
        emit CloseChannel(taskID, msg.sender, block.timestamp);
        selfdestruct(payable(msg.sender));
    } 
}