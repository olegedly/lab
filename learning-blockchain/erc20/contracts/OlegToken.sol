//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.0;

import "hardhat/console.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract OlegToken is ERC20, Ownable {
    constructor() ERC20("OlegToken", "OLEG") {
        _mint(owner(), 500);
    }
}
