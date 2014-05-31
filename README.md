What Is Sia?
============

Sia (/'saɪə/) is a compensation-based platform for distributed cloud storage. Anybody can join Sia as a storage host, and anybody can rent storage from the network. It's like Airbnb for your hard drive: through crowd-sourcing, we can eliminate a lot of overhead and improve quality of service.

Sia is a robust platform that stores files across hundreds of machines. If any of these machines fail, corrupt, or disconnect, the network will automatically repair the lost files with no break in service. Even with these protections against widespread failure or host withdrawal, the total redundancy is only around 25%—substantially less expensive than using RAID or a mirrored backup.

Sia is cheap. Very cheap. The amortized market cost of bulk storage is around $50 per TB, with the average disk lasting around 18 months. So on a monthly basis, the raw cost of hosting data is less than $3 per TB. After adding in Sia's redundancy and a profit margin, the cost should come to approximately $5 per TB-month, or $0.005 per GB-month. This is half of the price of Amazon Glacier, which charges 1 cent per GB-month.

Sia is fast. Every file is hosted across hundreds of machines, so downloads are highly parallel. In mosts cases, this should be enough to saturate your Internet connection. The power of distributed systems makes Sia both faster than Amazon S3 and cheaper than Amazon Glacier.

Sia is elastic. You can rent as much or as little storage as you want, and you pay for exactly what you are renting. You never need to guess whether you need the 100GB package or the 500GB package. Instead, you rent exactly as much as you use, expanding or contracting as you add and remove files. There is no fee for adjusting how much you are renting.

Sia is secure. By default, all data is encrypted on the client machine before being uploaded to the network. The encrypted data is then divided into pieces and distributed across hundreds of hosts. Only you can view the contents of your files.

Economic Model
======================

Sia compensates hosts using a cryptocurrency. This cryptocurrency, Siacoin, will be easily exchangeable for bitcoins and subsequently USD. When the currency launches, 10,000 siacoins will be mined per day and distributed to hosts according to the volume of storage they offer to the network. The number of coins mined per day will decrease until the 4 year mark, at which point the network will mine coins such that the annual inflation rate is kept permanently at 5%. No coins will be premined.

Storage can only be rented using siacoins. This presents a major inconvenience for most people, so we will be providing a service that allows users to rent storage with USD, transparently exchanging dollars for siacoins behind the scenes.

Hosts on Sia have two incomes: mining and rental payments. Depending on the price of Siacoin vs. the price of storage, this could lead to a scenario where the network cost of storage is actually cheaper than the raw cost of the storage, due to the added bonus earned through mining.

Siacoins are intended to be a means to an end, not a store of value. That's one of the reasons why we made the currency permanently inflationary. Rampant speculation is a major contributor to the instability of most cryptocurrencies, which damages their credibility. Instead of speculating in Siacoin, clients are recommended to purchase only as many as they need to store their data for a comfortable period of time, and buy more as needed. This protects the client in the event of a price swing; if the value of Siacoin suddenly plummets, they won't get burned.

Such swings may not turn out to be a serious problem though. If there is a sudden explosion in the amount of cheap storage available (e.g. if there is a major breakthrough in storage density), the value of Siacoin is likely to drop significantly. But the price per TB in siacoins should not change at all. Even though your siacoins are only worth half of what they used to be worth, they can be used to rent just as much storage as before. This is why we recommend equating siacoins to bytes rather than dollars.

Fundamentally, this implies that Siacoin has inherent value. There is an explicit reason to use Siacoin, which is to rent storage from Sia. This is a resource that can only be accessed using siacoins, which gives it a minimum value; a siacoin will never be worth less than the storage that it can buy. Additionally, if the value of the Siacoin skyrockets, new hosts will be incentivized to join the network, and everybody wins.

Protocol
========

Sia works quite a bit differently than existing cryptocurrencies. Instead of using proof of work, Sia uses proof of storage. We enforce storage proofs by breaking hosts into sets of 128, called quorums. Each member of the quorum hosts 16GB; if your machine is offering more than 16GB, you participate in more than one quorum.

The quorums individually follow a solution to the Byzantine generals problem presented [here](doc/The Byzantine Generals Problem.pdf). This provides a solution that guarantees all honest participants in a quorum will revive the information. Though the solution is complex, we've optimized the networking such that each block only takes a few minutes, comparable to bitcoin's block rate. Blocks that do not contain storage proofs could operate at a much faster rate.

Quorums communicate through a tree. Quorums use a verification algorithm to confirm with high probability that all blocks they receive are honest. The big advantage is that in general, quorums can operate without caring about the blocks produced in other quorums. In every existing cryptocurrency, every miner must know about every transaction, which is expensive and limits the types and volume of transactions that can occur. The exact algorithms for the tree will be discussed at greater length in the as-yet-incomplete whitepaper.

Finally, Sia will support a scripting system that is much lighter and more powerful than the scripting in existing cryptocurrencies. For example, a user could write a script to automatically adjust what files they are storing based on the amount of Siacoin in their wallet. Such a system could potentially be used for complex computation, enabling fully decentralized dynamic web hosting.

In general, we assume that 1) less than 50% of the total hosts on the network will be dishonest, and 2) no quorum will have more than 80% dishonest hosts. If 1) is true, then 2) can be achieved by randomly placing hosts around the network. Sia's method of entropy generation is a complex topic that will be covered at length in the whitepaper.

Cryptography and Erasure Coding
===============================

For cryptography, Sia uses [libsodium](https://github.com/jedisct1/libsodium). This includes local random number generation, hashing, and signatures using Curve25519 elliptic curve cryptography. As far as we can tell, this is a secure library that follows the best practices. We are not experts in this area however. We want Sia to employ the absolute best practices when using cryptography, so if you see something that is amiss, please let us know.

For erasure coding, Sia uses a Cauchy Reed Solomon coding library called [longhair](https://github.com/catid/longhair).

Project Status
==============

Currently, we have most parts of the basic quorum implemented. We have no part of the tree implemented, nor any part of the Byzantine-tolerant DHT that will allow quorums to communicate. We have the erasure coding libraries established and most of the crypto libraries in place (no encryption yet).

We expect to have a basic demo complete by April 25th, but only of the quorums. We expect to have a beta with transactions, the quorum tree, and some scripting in place by the end of June. We expect to be launching the full system sometime in August.

Where Can I Learn More?
=======================

Primary Contact: David Vorick, david@nebulouslabs.com
Secondary Contact: Luke Champine, luke@nebulouslabs.com
