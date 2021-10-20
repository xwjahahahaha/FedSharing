// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.7.6 <0.9.0;
import './SafeMath.sol';

contract Ownable{
    address public owner;
  
    constructor() {
        owner = msg.sender;
    }

    modifier onlyOwner(){
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }

    function transferOwnership(address newOwner) public onlyOwner{
        require(newOwner != address(0), "can't transfer owner for 0x0 address.");
        owner = newOwner;
    }
}

abstract contract ERC20{
    uint public totalSupply;
    function transfer(address to, uint value) virtual public;
    function balanceOf(address owner) virtual public view returns(uint);
    function allowance(address owner, address spender) virtual public view returns(uint);
    function transferFrom(address from, address to, uint value) virtual public;
    function approve(address spender, uint value) virtual public;
    event Transfer(address indexed from, address indexed to, uint value);
    event Approval(address indexed owner, address indexed spender, uint value);
 }

contract FSCoin is Ownable, ERC20{
    using SafeMath for uint;
    mapping(address => uint) public balances;
    mapping(address => mapping(address => uint)) public allowed;
    uint public constant MAX_UINT = 2**256-1;
    
    modifier onlyPayloadSize(uint size){
        require(!(msg.data.length < size+4), "Invalid short address");
        _;
    }
    
    constructor(uint _totalSupply){
        totalSupply = _totalSupply;
        balances[owner] = _totalSupply;
    }

    function transfer(address _to, uint _value) override public onlyPayloadSize(2 * 32){
        require(_value > 0, "Invalid transfer value.");
        balances[msg.sender] = balances[msg.sender].sub(_value);
        balances[_to] = balances[_to].add(_value);
        emit Transfer(msg.sender, _to, _value);
    }
    
    function balanceOf(address owner) override public view returns(uint){
        return balances[owner];
    }
    
    function transferFrom(address _from, address _to, uint _value) override public onlyPayloadSize(2 * 32){
        uint _allowance = allowed[_from][msg.sender];
        if (_allowance < MAX_UINT){
            allowed[_from][msg.sender] = _allowance.sub(_value);
        }
        balances[_from] = balances[_from].sub(_value);
        balances[_to] = balances[_to].add(_value);
        emit Transfer(_from, _to, _value);
    }
    
    function approve(address _spender, uint _value) override public onlyPayloadSize(2 * 32){
        require(!(_value != 0 && allowed[msg.sender][_spender] != 0), "You have only one chance to approve , you can only change it to 0 later");
        allowed[msg.sender][_spender] = _value;
        emit Approval(msg.sender, _spender, _value);
    }
    
    function allowance(address _owner, address _spender) override public view returns(uint remaining){
        return allowed[_owner][_spender];
    }
    
}