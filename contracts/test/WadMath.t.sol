// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";
import {WadMath} from "../src/WadMath.sol";

// forge-lint: disable-next-item(mixed-case-function)
contract TestWadMath {
    using WadMath for uint256;
    using WadMath for int256;

    function exposed_toWad_uint256(uint256 amount, uint8 decimals) external pure returns (uint256) {
        return amount.toWad(decimals);
    }

    function exposed_toWad_int256(int256 amount, uint8 decimals) external pure returns (int256) {
        return amount.toWad(decimals);
    }
}

contract WadMathTest is Test {
    TestWadMath public wadMath;

    uint8 immutable MAX_PRECISION = WadMath.MAX_PRECISION;

    function setUp() public {
        wadMath = new TestWadMath();
    }

    function test_toWad_uint256_success_with6Decimals() public view {
        uint256 result = wadMath.exposed_toWad_uint256(1_000_000, 6);
        assertEq(result, 1000000_000_000_000_000); // 18 - 6 = 12 zeros added
    }

    function test_toWad_uint256_success_withMaxPrecision() public view {
        uint256 amount = 123_456_789_012_345_678;
        uint256 result = wadMath.exposed_toWad_uint256(amount, MAX_PRECISION);
        assertEq(result, amount);
    }

    function test_toWad_uint256_revert_withOverPrecision() public {
        vm.expectRevert(WadMath.DecimalsExceedMaxPrecision.selector);
        wadMath.exposed_toWad_uint256(1000, 19);
    }

    function testFuzz_toWad_uint256(uint256 amount, uint8 decimals) public view {
        vm.assume(decimals <= MAX_PRECISION);

        if (decimals < MAX_PRECISION) {
            uint256 scaleFactor = 10 ** (MAX_PRECISION - decimals);
            vm.assume(amount <= type(uint256).max / scaleFactor);
        }

        uint256 result = wadMath.exposed_toWad_uint256(amount, decimals);

        if (decimals == MAX_PRECISION) {
            assertEq(result, amount);
        } else {
            uint256 expected = amount * (10 ** (MAX_PRECISION - decimals));
            assertEq(result, expected);
        }
    }

    function test_toWad_int256_success_with6Decimals() public view {
        int256 result = wadMath.exposed_toWad_int256(-123_456, 6);
        assertEq(result, -123456_000_000_000_000); // 18 - 6 = 12 zeros added
    }

    function test_toWad_int256_success_withMaxPrecision() public view {
        int256 amount = -123_456_789_012_345_678;
        int256 result = wadMath.exposed_toWad_int256(amount, MAX_PRECISION);
        assertEq(result, amount);
    }

    function test_toWad_int256_revert_withOverPrecision() public {
        vm.expectRevert(WadMath.DecimalsExceedMaxPrecision.selector);
        wadMath.exposed_toWad_int256(-1000, 19);
    }

    function testFuzz_toWad_int256(int256 amount, uint8 decimals) public view {
        vm.assume(decimals <= MAX_PRECISION);

        if (decimals < MAX_PRECISION) {
            int256 scaleFactor = int256(10 ** (MAX_PRECISION - decimals));
            if (amount > 0) {
                vm.assume(amount <= type(int256).max / scaleFactor);
            } else if (amount < 0) {
                vm.assume(amount >= type(int256).min / scaleFactor);
            }
        }

        int256 result = wadMath.exposed_toWad_int256(amount, decimals);

        if (decimals == MAX_PRECISION) {
            assertEq(result, amount);
        } else {
            int256 expected = amount * int256(10 ** (MAX_PRECISION - decimals));
            assertEq(result, expected);
        }
    }
}
