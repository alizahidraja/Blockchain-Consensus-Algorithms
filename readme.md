# Blockchain Consensus Algorithms

With the fame of blockchain on the rise and Go becoming a popular programming language, I thought this is a good time to share some of my experience in the mentioned.

I made 2 Blockchain consensus algorithms that are working on Go

## 1. Proof of Donation

The idea is to distribute the wealth equally between the people who have little to their names. The miner is the person with most donated coins, and there is a fixed amount of mining reward in the start. The miner obviously will have a lot of coins at this point. Before the mining of the next block, the miner must choose an amount of coins that he must donate to everyone else equally in the chain. This amount(donation amount) will be the base of choosing the next miner. Anyone can donate any amount to the network and the person who donates the most will be chosen as the miner.

## 2. Random Winner

Random winner with specified prize pools of a certain minimum and maximum amount, winner gets 50% and the rest is divided to the contributors in the proportion they contributed.
