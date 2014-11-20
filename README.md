What Is Sia?
============

Sia (/'saɪə/) is a compensation-based platform for distributed cloud storage. Anybody can join Sia as a storage host, and anybody can rent storage from the network. It's like Airbnb for your hard drive: through crowd-sourcing, we can eliminate a lot of overhead and improve quality of service.

Sia is a robust platform that stores files across hundreds of machines. Sia uses M of N erasure coding, meaning that of N hosts, only M need to remain online for the whole file to be recovered. Lost file pieces can be restored at any time.

Sia is cheap. Very cheap. The amortized market cost of bulk storage is around $50 per TB, with the average disk lasting around 18 months. So on a monthly basis, the raw cost of hosting data is less than $3 per TB. After adding in Sia's redundancy and a profit margin, the cost should come to approximately $5 per TB-month, or $0.005 per GB-month. This is half of the price of Amazon Glacier, which charges 1 cent per GB-month.

Sia is fast. Every file is hosted across hundreds of machines, so downloads are highly parallel. In mosts cases, this should be enough to saturate your Internet connection. The power of distributed systems makes Sia both faster than Amazon S3 and cheaper than Amazon Glacier.

Sia is elastic. You can rent as much or as little storage as you want, and you pay for exactly what you are renting. You never need to guess whether you need the 100GB package or the 500GB package. Instead, you rent exactly as much as you use, expanding or contracting as you add and remove files. There is no fee for adjusting how much you are renting.

Sia is secure. By default, all data is encrypted on the client machine before being uploaded to the network. The encrypted data is then divided into pieces and distributed across hundreds of hosts. Only you can view the contents of your files. Sia is decentralized, which means that many different parties hold your files as opposed to a single company. You are protected against sudden changes in price or Terms of Service that can be problematic when relying on a single party.

Economic Model
==============

Sia compensates hosts using a cryptocurrency. This cryptocurrency, Siacoin, will be easily exchangeable for bitcoins and subsequently USD. When the currency launches, 300,000 siacoins will be mined per block, decreasing by 1 each block. By year 5, the annual inflation will be 4.4%.

Initially, storage can only be rented using siacoins. This provides inital value to the siacoin (demand resulting from needing the currency), however it also causes inconvenience to storage renters. Support for renting storage using Bitcoins through a two-way peg (a sidechain) is planned, and additionally direct support for the USD is also planned.

The primary source of income for the developers behind Sia is not premined currency, but rather a fee taken from the storage that is hosted on the network. This gives Sia a strong income based on the amount of storage being consumed on the network, and not tethered to the volatility of a cryptocurrency. Most importantly, this income allows us to pay full time developers to improve the network and utility of Sia.

Protocol
========

Sia operates in a very similar fashion to Bitcoin, using a POW blockchain and many of the same constants.

There are two major changes to the Bitcoin protocol. The first is that the scripting system has been replaced with a simpler multisignature system that operates much like a simplified p2sh transaction. The second major change is the introduction of storage contracts, which are the foundation of the decentralized storage platform.

A client will contact a host, requesting that the host store a file for a certain period of time at an agreed upon price. If the host agrees, a contract will be created and submitted to the blockchain. The host is then required to prove to the network that it is storing the file, and must submit storage proofs at a regular interval. If the host is storing the file, the host is guaranteed to be paid. If the host is not storing the file, then the host will not be paid, and the client will get a refund.

More details can be found by reading the whitepaper.

Project Status
==============

We have recently released a new whitepaper, and plan to be in closed beta by the end of November. If you would like to be involved in the early beta (and have a high tolerance for potentially catastrophic bugs), please contact us.

We hope to be in open beta by the end of the year.

Contact Information
===================

david@nebulouslabs.com

luke@nebulouslabs.com
