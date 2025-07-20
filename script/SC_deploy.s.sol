// SPDX-License-Identifier: MIT

pragma solidity ^0.8.18;

import "../lib/forge-std/src/Script.sol";

import {stableCoin} from "../src/stableCoin.sol";
import {stableCoinEngine} from "../src/StableCoinEngine.sol";
import {helper} from "./helper_deploy.s.sol";

contract SC_deploy is Script {
    address[] public collateralTokensDataFeed;
    address[] public collateralTokens;

    function run() external returns (stableCoinEngine, stableCoin, helper) {
        vm.startBroadcast();

        helper helper_config = new helper();

        (address wethUsdPriceFeed, address wbtcUsdPriceFeed, address weth, address wbtc, uint256 deployerKey) =
            helper_config.activeConfig();

        collateralTokens = [weth, wbtc];
        collateralTokensDataFeed = [wethUsdPriceFeed, wbtcUsdPriceFeed];

        stableCoin Sc = new stableCoin();
        stableCoinEngine SCE = new stableCoinEngine(collateralTokens, collateralTokensDataFeed, address(Sc));
        Sc.transferOwnership(address(SCE));

        vm.stopBroadcast();

        return (SCE, Sc, helper_config);
    }
}
