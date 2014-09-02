Sia Release Notes:

The release branch of Sia is meant to function in a real world scenario, with turbulence and high internet latency. The constants have been changed as follows:

# The step duration has been changed to 10 seconds, which gives the hosts a bit of drift room, and plenty of time to pass heartbeats around. The Quorum Size is set to 16, because I want to see how stable things are when there are actually 16 full participants in different parts of the world. This results in a block time of 160 seconds and an upload time of 320 seconds, which I think is tolerable at this stage.
Quorum Size: 16
Step Duration: 10 seconds

# Each participant is expected to contribute 1GB at this point, and each file can be up to 768kb in size, which leaves room for lots of files in the quorum. At the StandardK of 3, you can fit about 4000 files on the quorum.
Atoms Per Quorum: 2^25 (1 GB)
Atoms Per Sector: 8192 (256kb)
Standard K: 3 (corresponding to max size of 768kb, max of 3900 files @ 768kb)

# Files are not currently allowed to be stored at a lower redundancy than 16/6
Max K: 6 (no code exists to utilize this however)

# Siblings have 5 blocks, or about 12 minutes to download all the files that they are missing. This is at most 1GB.
Sibling Passive Window: 5

# This should probably stay approx. the same size as the Sibling Passive Window.
Snapshot Length: 5
