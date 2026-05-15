// SPDX-License-Identifier: MIT

pragma solidity >=0.6.0 <0.9.0;

contract SimpleStorage {
    struct Person {
        uint256 number;
        string name;
    }
    Person[] public people;
    mapping(string => uint256) public nameToNumber;
    
    function getAllPeople() public view returns(Person[] memory) {
        return people;
    }
    
    function addPerson(uint256 _number, string memory _name) public {
        people.push(Person({number: _number, name: _name}));
        nameToNumber[_name] = _number;
    }
}
