// SPDX-License-Identifier: SEE LICENSE IN LICENSE
pragma solidity ^0.8.30;

import "./Log.sol";

contract P2M is Log {
  function getTimestamp() public view returns (uint256) {
    return block.timestamp;
  }
}
