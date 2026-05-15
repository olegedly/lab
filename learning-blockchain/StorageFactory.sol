// SPDX-License-Identifier: MIT

pragma solidity >=0.6.0 <0.9.0;

import "./SimpleStorage.sol";

contract StorageFactory {
    SimpleStorage[] public simpleStorageArray;
    function createSimpleStorageContract() public {
        simpleStorageArray.push(new SimpleStorage());
    }
    function getContractInstance(uint256 _simpleStorageIndex) internal view returns(SimpleStorage) {
        return SimpleStorage(address(simpleStorageArray[_simpleStorageIndex]));
    }
    function sfAddPerson(uint256 _simpleStorageIndex, string memory _name, uint _number) public {
        SimpleStorage _simpleStorageInstance = getContractInstance(_simpleStorageIndex);
        _simpleStorageInstance.addPerson(_number, _name);
    }
    function sfViewPerson(uint256 _simpleStorageIndex, uint256 _personIndex) view public returns(SimpleStorage.Person memory) {
        SimpleStorage _simpleStorageInstance = getContractInstance(_simpleStorageIndex);
        return _simpleStorageInstance.getAllPeople()[_personIndex];
    }
}
