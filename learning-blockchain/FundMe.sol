// SPDX-License-Identifier: MIT

pragma solidity >=0.6.6 <0.9.0;

import "@chainlink/contracts/src/v0.6/interfaces/AggregatorV3Interface.sol";

contract FundMe {
    address public owner;
    constructor() public {
        owner = msg.sender;
    }
    
    mapping (address => uint256) public addressToAmountFunded;
    address[] public funders;
    
    function fund() public payable {
        // uint256 minUSD = 50 * 10 ** 18;
        // require(getConvertedAmount(msg.value) >= minUSD, "You need to spend more ETH!");
        addressToAmountFunded[msg.sender] += msg.value;
        funders.push(msg.sender);
    }
    
    modifier onlyOwner {
        require(owner ==  msg.sender);
        _;
    }
    
    function withdraw() payable onlyOwner public {
        msg.sender.transfer(address(this).balance);
        for (uint256 funderIndex = 0; funderIndex < funders.length; funderIndex++) {
            addressToAmountFunded[funders[funderIndex]] = 0;
        }
        funders = new address[](0);
    }
    
    function getVersion() public view returns(uint256) {
        AggregatorV3Interface priceFeed = AggregatorV3Interface(0x8A753747A1Fa494EC906cE90E9f37563A8AF630e);
        return priceFeed.version();
    }
    
    function getPrice() public view returns(uint256) {
        AggregatorV3Interface priceFeed = AggregatorV3Interface(0x8A753747A1Fa494EC906cE90E9f37563A8AF630e);
        (,int256 answer,,,) = priceFeed.latestRoundData();
        return uint256(answer * 10 ** 10);
    }
    
    function getConvertedAmount(uint256 ethAmount) public view returns(uint256) {
        uint256 ethPrice = getPrice();
        uint256 ethAmountInUsd = (ethPrice * ethAmount) / (10 ** 18);
        return ethAmountInUsd;
    }
}