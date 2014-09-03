Usage Notes:

Run 'client-cli' to bring up the home menu. Type 'h' or 'help' at any menu screen to see what your options are. First you need to connect to the network, which you can do by typing 'c'. You will be asked to provide a port number to listen on, this is what port your local computer will listen on. Then you will be asked to provide a hostname, port number, and ID to connect to. Using the values '23.239.14.98', '9989', '1' will connect you to the official Sia test network. Alternatively, you can join a different network or create your own. When creating your own, you do not need to do the 'connect' step.

Once connected to the network, you need a wallet. This wallet will be used to pay you for providing storage to the network. To get a wallet, type 'g' from the home screen, and put in a wallet ID that you would like to have. Your requested ID may already be taken, which means you need to try again with a different ID. Creating a wallet will take a few minutes.

Once you have a wallet, you are ready to switch to server mode and create a server. Switch to server mode by typing 's', and join the network with a new participant by typing 'j' (for join). This will ask you to name your server, and to provide a directory for the server. Creating a server will take several minutes, perhaps as many as 6. There is only room on the official server for 16 participants, it may be full when you try to join. This number will be vastly increased in future releases.

To do things like upload, download, and send money, you need to load a wallet that you control. To do this, go to the home screen (press 'q' if you are on another screen) and type 'l', putting in the wallet you wish to load. This will bring you to the wallet menu. Type 'h' for a list of available actions.

At present, the client is not terribly intuitive. We have tried to make the help menus (available by pressing 'h') reasonably useful. Please feel free to play around, and remember to file a bug report if something breaks. This is an early alpha, things are likely to break. Mostly, have fun!

Developer Notes:

The release branch of Sia is meant to function in a real world scenario, with turbulence and high internet latency. The constants have been changed as follows:

Quorum size has been changed to 16, which brings us closer to the intended 128 of the full Sia release.
Step duration has been changed to 10 seconds, which means blocks will take about 170 seconds each. 10 seconds is needed to provide enough time for servers worldwide to communicate. Clock drift may be a significant issue during this release.

Atoms per quorum has been adjusted to 2^25, which results in a participant size of 1GB. This means that (depending on redundancy settings) the release quorum can hold between 1GB and 6GB of files. The sector size has been adjusted to 2^13 atoms, or 256kb. This means that the largest size for a single file is between 256kb and 1.5mb (again depending on redundancy settings). The client will be using a K of 3, which means the max file size is 768kb.

The minimum redundancy allowed currently is 16/6.

The sibing passive window has been adjusted to 20 blocks, or just under 1 hour. Siblings will need to download the quorum in this timeframe, which could mean up to 6GB of downloading, but will more realisticially max out around 2.5GB of downloading. Even at 6GB, only a 2mbps connection is needed to fit inside of the window.
