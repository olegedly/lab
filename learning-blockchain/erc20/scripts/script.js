const hre = require("hardhat");

async function main() {
  const OlegToken = await hre.ethers.getContractFactory("OlegToken");
  const olegToken = await OlegToken.deploy();

  await olegToken.deployed();

  console.log("OlegToken deployed to:", olegToken.address);
  // console.log(hre.ethers);
  const signer = await hre.ethers.getSigner();
  // await olegToken.transfer(signer.address, 500);
  console.log((await olegToken.totalSupply()).toString());
  console.log((await olegToken.balanceOf(signer.address)).toString());
  console.log(await olegToken.symbol());
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
