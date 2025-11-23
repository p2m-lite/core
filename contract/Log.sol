// SPDX-License-Identifier: SEE LICENSE IN LICENSE
pragma solidity ^0.8.30;

contract Log {
  struct WaterLog {
    address recorder;
    uint256 phValue;
    uint256 turbidity;
    uint256 timestamp;
  }

  WaterLog[] public waterLogs;

  event LogStored(address indexed recorder, uint256 phValue, uint256 turbidity, uint256 timestamp);

  function storeLog(uint256 _phValue, uint256 _turbidity) public {
    WaterLog memory newLog = WaterLog({
      recorder: msg.sender,
      phValue: _phValue,
      turbidity: _turbidity,
      timestamp: block.timestamp
    });

    emit LogStored(msg.sender, _phValue, _turbidity, block.timestamp);
    waterLogs.push(newLog);
  }

  function getLogCount() public view returns (uint256) {
    return waterLogs.length;
  }

  function getLog(uint256 index) public view returns (WaterLog memory) {
    require(index < waterLogs.length, "Index out of bounds");
    return waterLogs[index];
  }

  function getLogsOf(address recorder) public view returns (WaterLog[] memory) {
    uint256 count = 0;
    for (uint256 i = 0; i < waterLogs.length; i++) {
      if (waterLogs[i].recorder == recorder) {
        count++;
      }
    }

    WaterLog[] memory result = new WaterLog[](count);
    uint256 j = 0;
    for (uint256 i = 0; i < waterLogs.length; i++) {
      if (waterLogs[i].recorder == recorder) {
        result[j] = waterLogs[i];
        j++;
      }
    }

    return result;
  }
}
