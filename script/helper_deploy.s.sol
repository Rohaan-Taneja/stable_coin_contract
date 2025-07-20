// SPDX-License-Identifier: MIT

pragma solidity ^0.8.18;

import {ERC20Mock, ERC20} from "../lib/openzeppelin-contracts/contracts/mocks/token/ERC20Mock.sol";

import {MockV3Aggregator} from "../test/mock/mockV3Aggregator.sol";

contract helper {
    struct deploymentConfig {
        address wethUsdPriceFeed;
        address wbtcUsdPriceFeed;
        address weth;
        address wbtc;
        uint256 deployerKey;
    }

    uint8 public constant DECIMALS = 8;
    int256 public constant ETH_USD_PRICE = 2000e8;
    int256 public constant BTC_USD_PRICE = 1000e8;
    uint256 public constant DEFAULT_PRIVATE_KEY = 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80;

    deploymentConfig public activeConfig;

    constructor() {
        if (block.chainid == 1115511) {
            activeConfig = return_sepolia_config();
        } else {
            activeConfig = create_and_return_anvil_config();
        }
    }

    function create_and_return_anvil_config() public returns (deploymentConfig memory anvilConfig) {
        ERC20 weth = new ERC20Mock();
        ERC20 wBTC = new ERC20Mock();

        MockV3Aggregator ethDataFeed = new MockV3Aggregator(DECIMALS, ETH_USD_PRICE);
        MockV3Aggregator BTC_DataFeed = new MockV3Aggregator(DECIMALS, BTC_USD_PRICE);

        anvilConfig = deploymentConfig({
            wethUsdPriceFeed: address(ethDataFeed),
            wbtcUsdPriceFeed: address(BTC_DataFeed),
            weth: address(weth), // erc20 contract deployed on anvil
            wbtc: address(wBTC), // erc20 contract deployed on anvil
            deployerKey: DEFAULT_PRIVATE_KEY
        });
    }

    function return_sepolia_config() public pure returns (deploymentConfig memory sepoliaconfig) {
        //  ERC20  weth = new ERC20Mock();
        //  ERC20  wBTC = new  ERC20Mock();

        // weth and wbtc contract deployment on sepolia then

        sepoliaconfig = deploymentConfig({
            wethUsdPriceFeed: address(0),
            wbtcUsdPriceFeed: address(0),
            weth: address(0), // erc20 contract deployed on sepolia
            wbtc: address(0), // erc20 contract deployed on sepolia
            deployerKey: 123456789
        });
    }
}
