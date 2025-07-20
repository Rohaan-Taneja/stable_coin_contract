// SPDX-License-Identifier: MIT

// Layout of Contract:
// version
// imports
// interfaces, libraries, contracts
// errors
// Type declarations
// State variables
// Events
// Modifiers
// Functions

// Layout of Functions:
// constructor
// receive function (if exists)
// fallback function (if exists)
// external
// public
// internal
// private
// view & pure functions

import "../node_modules/@openzeppelin/contracts/utils/ReentrancyGuard.sol";

import "./stableCoin.sol";

import {console} from "../lib/forge-std/src/console.sol";

import {IERC20} from "../node_modules/@openzeppelin/contracts/token/ERC20/IERC20.sol";

import {AggregatorV3Interface} from
    "../lib/chainlink-evm/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";

pragma solidity ^0.8.18;

contract stableCoinEngine is ReentrancyGuard {
    // ***********************
    // Errors
    // ***********************.
    error stableCoinEngine_nonZeroValueNeeded();

    error stableCoinEngine_notTakingThisToken_asCollateral();

    error stableCoinEngine_UnequalArrayOfAddresses();

    error stableCoinEngine_transferFailed();

    error SC_cannotMintThisMuchToken_askingAmount_Greater_than_threshold();

    error SC_engine_mintFailed();

    error SC_transferFailed();

    error InsufficientBalance();

    error cannot_withdrawThisMuchCollateraLTokens_burnSome_stableCoins_toWithdrawCollateralTokens();

    error collateral_TokenTransferFailed();

    error failed_to_burn_SCtokens();

    error cannotLiquidate_HF_isInSafeZone();

    // ***********************
    // state variables
    // ***********************
    mapping(address tokenAddress => address dataFeedAddress) private s_dataFeed;

    address[] collateral_TokensAddresses;

    stableCoin private sc_coin;

    mapping(address user => mapping(address collateralTokenAddress => uint256 amount)) public user_collateral_amount;

    mapping(address user => uint256 SC) private user_SC_balance;

    //
    uint256 private constant eth_precision = 1e18;

    uint256 private constant dataFeed_precision = 1e8;

    // ***********************
    // Events
    // ***********************

    event collateralTokenDeposited(address indexed user, address indexed collateralTokenAddress, uint256 amount);

    event Successfull_collateralWithdrawal(
        address indexed user, uint256 indexed collateralAmount_to_withdraw, address indexed collateralTokenAddress
    );

    event LiquidatedCollateral(
        address Liquidator, address User_withUnderCollateral, uint256 debtAmount, address collateralTokenAddress
    );

    // ***********************
    // modifiers
    // ***********************
    modifier nonZero(uint256 value) {
        if (value <= 0) {
            revert stableCoinEngine_nonZeroValueNeeded();
        }
        _;
    }

    modifier acceptedCollateral(address collateralToken) {
        if (s_dataFeed[collateralToken] == address(0)) {
            revert stableCoinEngine_notTakingThisToken_asCollateral();
        }
        _;
    }

    // ***********************
    // constructor
    // ***********************
    constructor(
        address[] memory collateralTokenAddresses,
        address[] memory dataFeed_of_each_collateral,
        address SC_address
    ) {
        if (collateralTokenAddresses.length != dataFeed_of_each_collateral.length) {
            revert stableCoinEngine_UnequalArrayOfAddresses();
        }

        // storing data in the mapping
        for (uint256 i = 0; i < collateralTokenAddresses.length; i++) {
            s_dataFeed[collateralTokenAddresses[i]] = dataFeed_of_each_collateral[i];

            // storing all the accepted collateral token addresses
            collateral_TokensAddresses.push(collateralTokenAddresses[i]);
        }

        sc_coin = stableCoin(SC_address);
    }

    // ***********************
    // External functions
    // ***********************

    //
    function deposit_and_mint_SC(address collateralTokenAddress, uint256 amount, uint256 amountToMint) public {
        depositCollateral(collateralTokenAddress, amount);
        mintSC(amountToMint);
    }

    /**
     * @param collateralTokenAddress => token address of the collateral that we are depositing
     * @param amount => how much token are depositing
     * note this will expect you have approve amount tokens to contract and it will transferfrom your wallet and update your status
     *      in our contract
     */
    function depositCollateral(address collateralTokenAddress, uint256 amount)
        public
        payable
        nonReentrant
        nonZero(amount)
        acceptedCollateral(collateralTokenAddress)
    {
        user_collateral_amount[msg.sender][collateralTokenAddress] += amount;

        bool success = IERC20(collateralTokenAddress).transferFrom(msg.sender, address(this), amount);

        if (!success) {
            revert stableCoinEngine_transferFailed();
        }

        emit collateralTokenDeposited(msg.sender, collateralTokenAddress, amount);
    }

    /**
     * @param collateralTokenAddress which colllateral token user2 is liquidating
     * @param user for which user ,this liquidation is
     * @param debtAmount how much stable coin is given back to SC engine(in 1e18 format)
     * Note this will take stable coin from user , liquidate the user debt
     * and give (amount+10) of collateral to user who is doing this liquidating process*
     * it follows CEI
     */
    function LiquidateCollateral(address collateralTokenAddress, address user, uint256 debtAmount)
        external
        acceptedCollateral(collateralTokenAddress)
        nonZero(debtAmount)
        nonReentrant
    {
        // if greater than 1, we will revert , if less than 1 , we will proceed to liquidation process
        if (!revert_if_HealthFactor_lessThan_1(user)) {
            revert cannotLiquidate_HF_isInSafeZone();
        }
        require(debtAmount <= user_SC_balance[user], "user debt is less that provided");

        // taking and burning Stbale coin from msg.sender
        user_SC_balance[user] -= debtAmount;
        bool success = sc_coin.transferFrom(msg.sender, address(this), debtAmount);

        if (!success) {
            revert SC_transferFailed();
        }
        sc_coin.burnTokens(address(this), debtAmount);

        // transfering collateral tokens to liquidaton( debt_amount + 10 => 11*debt_amount/10)
        uint256 collateral_to_transfer_to_liquidator =
            howMuch_collateralTokens_in_thisMuchUSD(collateralTokenAddress, 11 * debtAmount / 10);

        // if collateral value is gone below the debt amount .eg 100$ eth => 20$ eth ,and SC taken = 50.
        require(
            collateral_to_transfer_to_liquidator <= user_collateral_amount[user][collateralTokenAddress],
            "user dont have debt amount equivalent balance as collateral"
        );

        // tranferring debt amount + 10% bonus of collateral token to liquidator
        IERC20(collateralTokenAddress).transfer(msg.sender, collateral_to_transfer_to_liquidator);

        emit LiquidatedCollateral(msg.sender, user, debtAmount, collateralTokenAddress);
    }

    /**
     * @param collateralTokenAddress => the address of the collateral token
     * @param collateralAmount_to_withdraw => how much collateral token you want to redeem from you total collateral
     * @param amount => this is in 1e18 format and user will approve amount token and contract will burn it
     *
     */
    function burn_SC_and_redeem_collateral(
        address collateralTokenAddress,
        uint256 collateralAmount_to_withdraw,
        uint256 amount
    ) public {
        burn_SC_Tokens(amount);
        withdrawCollateral(collateralTokenAddress, collateralAmount_to_withdraw);
    }

    /**
     * @param collateralTokenAddress => the address of the collateral token
     * @param collateralAmount_to_withdraw => how much collateral token you want to redeem from you total collateral
     * Note => this will give user back its collateral , if its health factor is not breaking ( after redeem still greater than 1
     *  it follows CEI
     *
     */
    function withdrawCollateral(address collateralTokenAddress, uint256 collateralAmount_to_withdraw)
        public
        nonZero(collateralAmount_to_withdraw)
        acceptedCollateral(collateralTokenAddress)
        nonReentrant
    {
        if (user_collateral_amount[msg.sender][collateralTokenAddress] < collateralAmount_to_withdraw) {
            revert InsufficientBalance();
        }

        user_collateral_amount[msg.sender][collateralTokenAddress] -= collateralAmount_to_withdraw;

        bool success = IERC20(collateralTokenAddress).transfer(msg.sender, collateralAmount_to_withdraw);

        if (!success) {
            revert collateral_TokenTransferFailed();
        }

        // checking health factor
        if (revert_if_HealthFactor_lessThan_1(msg.sender)) {
            revert cannot_withdrawThisMuchCollateraLTokens_burnSome_stableCoins_toWithdrawCollateralTokens();
        }

        emit Successfull_collateralWithdrawal(msg.sender, collateralAmount_to_withdraw, collateralTokenAddress);
    }

    /**
     * @param amount => this is in 1e18 format and user will approve amount token and contract will burn it
     * Note => this will burn SC token from user account
     * this follows CEI
     */
    function burn_SC_Tokens(uint256 amount) internal nonZero(amount) {
        if (user_SC_balance[msg.sender] < amount) {
            revert InsufficientBalance();
        }

        user_SC_balance[msg.sender] -= amount;

        bool success = sc_coin.transferFrom(msg.sender, address(this), amount);
        if (!success) {
            revert SC_transferFailed();
        }

        sc_coin.burnTokens(address(this), amount);

        //    bool success = sc_coin.burnTokens(msg.sender, amount);
    }

    /**
     * @param amountToMint => it shoudl be in 1e18 precision
     *  note => this will check if healthfactor is greater that 1, if yes
     *          then it will deposit amount tokens (in smallest unit/1e18) to user belance
     *
     *
     */
    function mintSC(uint256 amountToMint) public nonZero(amountToMint) {
        user_SC_balance[msg.sender] += amountToMint;

        // if true , it means HF < 1
        if (revert_if_HealthFactor_lessThan_1(msg.sender)) {
            revert SC_cannotMintThisMuchToken_askingAmount_Greater_than_threshold();
        }

        // amounttoMint in 1e18
        bool minted = sc_coin.mintStableCoin(msg.sender, amountToMint);
        if (!minted) {
            revert SC_engine_mintFailed();
        }
    }

    // calculate total collateral value (all tokens)
    function calculate_TotalCollateralValue_In_USD(address user) public view returns (uint256) {
        uint256 totalCollateralValueInUSD = 0;

        for (uint256 i = 0; i < collateral_TokensAddresses.length; i++) {
            // tempAmount is also in 1e18 format
            uint256 tempAmount = user_collateral_amount[user][collateral_TokensAddresses[i]];

            if (tempAmount != 0) {
                // tokenaddress , amount  => calculate through data feed
                uint256 thisCollateralTotalAmount = getCollateralTokenValue(collateral_TokensAddresses[i], tempAmount);

                totalCollateralValueInUSD += thisCollateralTotalAmount;
            }
        }

        return totalCollateralValueInUSD;
    }

    // data feed function to calculate token value in usd*token_amount through CHAINLINK DATAFEED
    // gives answer in e18
    /**
     * @param collateralTokenAddress => then token whose current price we want to know from price feed
     * @param amount => how much tokens we have (x*1e18 format)
     * note this will get current collateral token value in usd and then return whats the value of amount no of tokens
     */
    function getCollateralTokenValue(address collateralTokenAddress, uint256 amount) public view returns (uint256) {
        AggregatorV3Interface dataFeed = AggregatorV3Interface(s_dataFeed[collateralTokenAddress]);

        (
            /* uint80 roundId */
            ,
            int256 answer,
            /*uint256 startedAt*/
            ,
            /*uint256 updatedAt*/
            ,
            /*uint80 answeredInRound*/
        ) = dataFeed.latestRoundData();

        console.log("this is the amount from aggregator", (uint256(answer) * 1e10 * amount) / 1e18);

        // answer will come in 1e8 precision , we will multiply by 1e10 to convert it into 1e18 precison , then multiply by amount_of_token_deposited(in 1e18 format)
        // the output is in 1e36 , so we will divide by 1e18 to get the final asnwer in 1e18 format
        // now the answer will be in 1e18 precision
        return ((uint256(answer) * 1e10) * amount) / 1e18; //result in 1e18 precision
    }

    /**
     * @param collateralTokenAddress => The collateral token that I want anser for
     * @param thisUsdAmount => how much collateral token in this much usd (in 1e18 format precision)
     * note => this function will calculate how much collateral tokens are there in this_much_usd
     */
    function howMuch_collateralTokens_in_thisMuchUSD(address collateralTokenAddress, uint256 thisUsdAmount)
        public
        view
        returns (uint256 result)
    {
        // coll_amount
        // ----------   X  (debtAmount X  1.1)       1.1debtAmount = amount + 10% bonus
        // coll_in_USD

        uint256 value_of_1_collateralToken_in_USD = getCollateralTokenValue(collateralTokenAddress, 1);

        result = thisUsdAmount / value_of_1_collateralToken_in_USD;
    }

    function revert_if_HealthFactor_lessThan_1(address user) internal view returns (bool) {
        // its basically caclulation total collateral in usd/2 , we will give halff usd value
        uint256 totalCollateralValueInUSD_threshold = (calculate_TotalCollateralValue_In_USD(user) * 50) / 100;

        require(user_SC_balance[user] != 0, "user stable coin balance is zero");

        // total kitna de sakte hai /kitne maang rha  ? >=1 , (500 de sakte hai , usse kmm mangega to dedenge else nhi)
        // both up and down are in 1e18 precision
        uint256 healthFactor = (totalCollateralValueInUSD_threshold) / user_SC_balance[user];
        return healthFactor < 1;

        // kitne krr sakte hai /(kitne leliye + kitne maang rha hai) >1 for taking more , below it is undercollateralize
        // kitne krr sakta - (kitne krr liye hai) > (or kitne krna cha rha hai) => if true then mint else revert
    }
}
