// SPDX-License-Identifier: MIT

pragma solidity ^0.8.18;

import "../../lib/forge-std/src/Test.sol";

import {console} from "../../lib/forge-std/src/console.sol";

import {SC_deploy} from "../../script/SC_deploy.s.sol";

import {stableCoin} from "../../src/stableCoin.sol";
import {stableCoinEngine} from "../../src/StableCoinEngine.sol";
import {IERC20} from "../../node_modules/@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {ERC20Mock, ERC20} from "../../lib/openzeppelin-contracts/contracts/mocks/token/ERC20Mock.sol";

import {helper} from "../../script/helper_deploy.s.sol";

contract SCE_testing is Test {
    SC_deploy public deployedContracts;
    stableCoin Sc;
    stableCoinEngine SCE;
    helper helper_config;
    address wethUsdPriceFeed;
    address wbtcUsdPriceFeed;
    address weth;
    address wbtc;
    uint256 deployerKey;

    // for constructor test only
    address[] public tokenAddress_array;
    address[] public data_feed_array;

    address user = 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266;

    function setUp() public {
        deployedContracts = new SC_deploy();
        (SCE, Sc, helper_config) = deployedContracts.run();

        (wethUsdPriceFeed, wbtcUsdPriceFeed, weth, wbtc, deployerKey) = helper_config.activeConfig();
    }

    function test_constructor() external {
        tokenAddress_array.push(weth);
        tokenAddress_array.push(wbtc);

        data_feed_array.push(wethUsdPriceFeed);

        vm.expectRevert(stableCoinEngine.stableCoinEngine_UnequalArrayOfAddresses.selector);
        new stableCoinEngine(tokenAddress_array, data_feed_array, address(Sc));
    }

    function test_collateralInUSD() external view {
        uint256 weth_giving = 15e18;
        // 15e18 * 2000e8
        uint256 expectedAnswer = 30000 * 1e18;

        uint256 incomingAnswer = SCE.getCollateralTokenValue(weth, weth_giving);

        assert(expectedAnswer == incomingAnswer);
    }

    function test_deposit_and_mint() external {
        // weth ek user ke account me mint krrwayenge
        // fir ek prank through that user

        vm.startPrank(user);
        ERC20Mock(weth).mint(user, 100e18);
        console.log("user balance ", IERC20(weth).balanceOf(user));

        // approving 10 eth token to contract
        IERC20(weth).approve(address(SCE), 10e18);

        // trasfering 10 token to our contract
        SCE.depositCollateral(weth, 10e18);
        assert(10e18 == SCE.user_collateral_amount(user, weth));

        console.log("this is the user balance", SCE.user_collateral_amount(user, weth));

        // we can mint upto 10 ether * 2000 /2 = 10000$ max , after that we will get revrt error of health factor

        // amount will be in 1e18 terms , rest evry thing in 1e18 terms
        // aggregator answer in 1e8 , total collaterla value converted to 1e18
        SCE.mintSC(10000e18);
        assert(10000e18 == Sc.balanceOf(user));
        vm.stopPrank();
    }

    function test_how_much_token_in_thisMuchUSD() public {
        uint256 this_muchUSD = 2000e18;
        uint256 amount = SCE.howMuch_collateralTokens_in_thisMuchUSD(weth, this_muchUSD);

        uint256 expected = 1e18;

        assertEq(amount, expected);
    }

    function test_burn_and_Redeem_collateral() public {
        vm.startPrank(user);

        ERC20Mock(weth).mint(user, 100e18);

        ERC20Mock(weth).approve(address(SCE), 10e18);

        SCE.depositCollateral(address(weth), 10e18);

        SCE.mintSC(10000e18);

        IERC20(Sc).approve(address(SCE), 5000e18);

        SCE.burn_SC_and_redeem_collateral(address(weth), 5e18, 5000e18);

        assertEq(ERC20Mock(weth).balanceOf(user), 95e18) ;
    }
}
