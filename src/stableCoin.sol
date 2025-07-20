// SPDX-License-Identifier: MIT

pragma solidity ^0.8.18;

import "../node_modules/@openzeppelin/contracts/token/ERC20/ERC20.sol";

import "../node_modules/@openzeppelin/contracts/access/Ownable.sol";

contract stableCoin is ERC20, Ownable {
    error stableCoin_cannotBeMinted_to_null_address(address userAddress);

    error stableCoin_accountAddress_cannot_be_null();

    error cannot_MINT_0_TOKENS();

    error cannot_Burn_0_TOKENS();

    error cannot_burnMoreThan_AccountBalance();

    // change it to governance engine address
    constructor() ERC20("stableCoin", "_coin") Ownable(msg.sender) {}

    function mintStableCoin(address user_address, uint256 no_of_tokens) external onlyOwner returns (bool) {
        if (user_address == address(0)) {
            revert stableCoin_cannotBeMinted_to_null_address(user_address);
        }
        if (no_of_tokens <= 0) {
            revert cannot_MINT_0_TOKENS();
        }

        _mint(user_address, no_of_tokens);
        return true;
    }

    function burnTokens(address user_Address, uint256 no_of_tokens) external onlyOwner returns (bool) {
        if (user_Address == address(0)) {
            revert stableCoin_accountAddress_cannot_be_null();
        }
        if (balanceOf(user_Address) < no_of_tokens) {
            revert cannot_burnMoreThan_AccountBalance();
        }
        if (no_of_tokens <= 0) {
            revert cannot_Burn_0_TOKENS();
        }

        _burn(user_Address, no_of_tokens);
        return true;
    }
}
