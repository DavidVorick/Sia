Sia Release Notes:

The release branch of Sia is meant to function in a real world scenario, with turbulence and high internet latency. The constants have been changed as follows:

Quorum size has been changed to 16, which brings us closer to the intended 128 of the full Sia release.
Step duration has been changed to 10 seconds, which means blocks will take about 170 seconds each. 10 seconds is needed to provide enough time for servers worldwide to communicate. Clock drift may be a significant issue during this release.

Atoms per quorum has been adjusted to 2^25, which results in a participant size of 1GB. This means that (depending on redundancy settings) the release quorum can hold between 1GB and 6GB of files. The sector size has been adjusted to 2^13 atoms, or 256kb. This means that the largest size for a single file is between 256kb and 1.5mb (again depending on redundancy settings). The client will be using a K of 3, which means the max file size is 768kb.

The minimum redundancy allowed currently is 16/6.

The sibing passive window has been adjusted to 20 blocks, or just under 1 hour. Siblings will need to download the quorum in this timeframe, which could mean up to 6GB of downloading, but will more realisticially max out around 2.5GB of downloading. Even at 6GB, only a 2mbps connection is needed to fit inside of the window.
